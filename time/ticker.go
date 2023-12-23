package time

import (
	"context"
	"errors"
	"time"

	"github.com/kwitsch/rueidistools/helper"
	"github.com/kwitsch/rueidistools/model"
	"github.com/redis/rueidis"
)

// Ticker is a ticker that uses redis keyspace events to tick
type Ticker struct {
	r   tickerRuntime
	C   <-chan time.Time
	Err <-chan error
}

// NewTicker creates a new ticker instance
func NewTicker(ctx context.Context, d time.Duration, name, prefix string, client rueidis.Client) (Ticker, error) {
	duration := helper.DurationToTTL(d)
	if duration <= 0 {
		return Ticker{}, errors.New("the ticker duration hast to be at least one second")
	}

	cctx, cancel := context.WithCancel(ctx)

	dClient, dcCancel := client.Dedicate()

	if err := helper.EnableExpiredNKE(ctx, client); err != nil {
		defer cancel()
		defer dcCancel()

		return Ticker{}, err
	}

	cChan := make(chan time.Time, 1)
	errChan := make(chan error, 1)
	t := Ticker{
		r: tickerRuntime{
			key:      model.NewKey(name, prefix),
			duration: duration,
			client:   dClient,
			u:        make(chan int64, 1),
			d:        make(chan bool, 1),
			c:        cChan,
			err:      errChan,
		},
		C:   cChan,
		Err: errChan,
	}

	go t.r.run(cctx, cancel, dcCancel)

	return t, nil
}

// Reset resets the ticker duration to the given duration
func (t *Ticker) Reset(d time.Duration) {
	t.r.u <- helper.DurationToTTL(d)
}

// Close closes the ticker
func (t *Ticker) Close() {
	t.r.d <- true
}

// tickerRuntime is the internal runtime of a ticker
type tickerRuntime struct {
	key      model.Key
	duration int64
	client   rueidis.DedicatedClient
	u        chan int64
	d        chan bool
	c        chan time.Time
	err      chan error
}

// run is the main loop of the ticker
func (r *tickerRuntime) run(ctx context.Context, ctxCancel context.CancelFunc, clientCancel func()) {
	// close rueidis client on exit
	defer clientCancel()
	// close context on exit
	defer ctxCancel()
	// close channels on exit
	defer close(r.u)
	defer close(r.d)
	defer close(r.c)

	// subscribe to key space events for the ticker key
	if err := r.client.Receive(ctx,
		r.client.B().Psubscribe().Pattern(r.key.KeySpacePattern()).Build(),
		func(m rueidis.PubSubMessage) {
			r.c <- time.Now()
		}); err != nil {
		r.err <- err
	}

	// set the ticker key
	if err := r.set(ctx, false); err != nil {
		r.err <- err
	}

	for {
		select {
		// update the duration of the ticker key
		case d := <-r.u:
			r.duration = d
			r.raiseError(r.set(ctx, true))
			// close the ticker if an error occurred
		case <-r.err:
			r.d <- true
			// send a tick
		case <-r.c:
			r.raiseError(r.set(ctx, false))
			// remove the ticker key if the ticker is closed
		case <-r.d:
			r.raiseError(r.remove(ctx))

			return
			// close the ticker if the context is done
		case <-ctx.Done():
			return
		}
	}
}

// raiseError sends an error to the error channel if the error is not nil
func (r *tickerRuntime) raiseError(err error) {
	if err != nil {
		r.err <- err
	}
}

// set sets the ticker key to the current time and the duration as TTL
func (r *tickerRuntime) set(ctx context.Context, overwrite bool) error {
	if !overwrite {
		exists, err := r.client.Do(ctx,
			r.client.B().
				Exists().
				Key(r.key.String()).
				Build()).AsBool()
		if err != nil {
			return err
		}

		if exists {
			return nil
		}
	}

	res := r.client.Do(ctx,
		r.client.B().Set().
			Key(r.key.String()).
			Value(helper.TTLToString(r.duration)).
			ExSeconds(r.duration).
			Build())

	return res.Error()
}

// remove removes the ticker key
func (r *tickerRuntime) remove(ctx context.Context) error {
	res := r.client.Do(ctx,
		r.client.B().Del().
			Key(r.key.String()).
			Build())

	return res.Error()
}
