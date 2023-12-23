package time

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/kwitsch/rueidistools/model"
	"github.com/kwitsch/rueidistools/util"
	"github.com/redis/rueidis"
)

// Ticker is a ticker that uses redis keyspace events to tick
type Ticker struct {
	r tickerRuntime    // ticker runtime
	C <-chan time.Time // The channel on which the ticks are delivered.
}

// NewTicker creates a new ticker instance and starts it.
func NewTicker(ctx context.Context, d time.Duration, name, prefix string, client rueidis.Client) (Ticker, error) {
	duration := util.DurationToTTL(d)
	if duration <= 0 {
		return Ticker{}, errors.New("the ticker duration hast to be at least one second")
	}

	// create dedicated client for the ticker
	dClient, dcCancel := client.Dedicate()

	// enable expired key events
	if err := util.EnableExpiredNKE(ctx, client); err != nil {
		defer dcCancel()

		return Ticker{}, err
	}

	// create ticker channel
	cChan := make(chan time.Time, 1)

	// create ticker instance
	t := Ticker{
		r: tickerRuntime{
			key:      model.NewKey(name, prefix),
			duration: duration,
			client:   dClient,
			done:     make(chan bool, 1),
			c:        cChan,
		},
		C: cChan,
	}

	// initialize the ticker runtime
	if err := t.r.init(ctx); err != nil {
		defer dcCancel()

		return Ticker{}, err
	}

	// start the ticker runtime
	go t.r.run(ctx, dcCancel)

	return t, nil
}

// Reset resets the ticker duration to the given duration
func (t *Ticker) Reset(ctx context.Context, d time.Duration) error {
	duration := util.DurationToTTL(d)
	if duration <= 0 {
		return errors.New("the ticker duration hast to be at least one second")
	}

	t.r.duration = duration

	return t.r.set(ctx, true)
}

// Stop turns off a ticker. After Stop, no more ticks will be sent.
// Stop does not close the channel, to prevent a concurrent goroutine
// reading from the channel from seeing an erroneous "tick".
func (t *Ticker) Stop(ctx context.Context) error {
	// send stop signal to the runtime
	defer t.r.stop()
	// remove the ticker key
	return t.r.remove(ctx)
}

// tickerRuntime is the internal runtime of a ticker
type tickerRuntime struct {
	key      model.Key               // ticker key
	duration int64                   // ticker duration
	client   rueidis.DedicatedClient // dedicated client for the ticker
	done     chan bool               // stop signal channel
	c        chan time.Time          // ticker channel
}

func (r *tickerRuntime) init(ctx context.Context) error {
	// subscribe to key space events for the ticker key
	if err := r.client.Receive(ctx,
		r.client.B().Psubscribe().Pattern(r.key.KeySpacePattern()).Build(),
		func(m rueidis.PubSubMessage) {
			// send the current time to the ticker channel
			r.c <- time.Now()
			// set the ticker key again and close the ticker if it fails
			if err := r.set(ctx, false); err != nil {
				r.done <- true
			}
		}); err != nil {
		return err
	}

	// set the ticker key
	if err := r.set(ctx, false); err != nil {
		return err
	}

	return nil
}

// run is the main loop of the ticker
func (r *tickerRuntime) run(ctx context.Context, clientCancel func()) {
	// close rueidis client on exit
	defer clientCancel()

	for {
		select {
		// close the ticker if stop signal is received
		case <-r.done:
			return
			// close the ticker if the context is done
		case <-ctx.Done():
			return
		}
	}
}

// stop sends a stop signal to the ticker runtime
func (r *tickerRuntime) stop() {
	r.done <- true
}

// set sets the ticker key to the current duration and also the duration as TTL.
// If overwrite is false, the key is only set if it does not exist.
// If it exists and the stored duration is different from the current duration, the current duration is updated.
func (r *tickerRuntime) set(ctx context.Context, overwrite bool) error {
	if !overwrite {
		res, err := r.client.Do(ctx,
			r.client.B().
				Get().
				Key(r.key.String()).
				Build()).ToMessage()
		if err != nil {
			return err
		}

		if !res.IsNil() {
			if val, err := res.AsInt64(); err == nil && val != r.duration {
				r.duration = val
			}

			return nil
		}
	}

	res := r.client.Do(ctx,
		r.client.B().Set().
			Key(r.key.String()).
			Value(strconv.FormatInt(r.duration, 10)).
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
