package time

import (
	"context"
	"errors"
	gotime "time"

	"github.com/kwitsch/rueidistools/helper"
	"github.com/kwitsch/rueidistools/model"
	"github.com/rueian/rueidis"
)

type Ticker struct {
	key          *model.Key
	duration     model.TTL
	client       rueidis.DedicatedClient
	clientCancel func()
	ctxCa        context.CancelFunc
	C            chan gotime.Time
	Err          chan error
}

func NewTicker(d model.TTL, name, prefix string, client rueidis.Client) (*Ticker, error) {
	if d.SecondsUI32() == 0 {
		return nil, errors.New("the ticker duration hast to be at least one second")
	}

	ctx, ctxCa := context.WithCancel(context.Background())

	dClient, dcCancel := client.Dedicate()
	res := Ticker{
		key:          model.NewKey(prefix, name),
		duration:     d,
		client:       dClient,
		clientCancel: dcCancel,
		ctxCa:        ctxCa,
		C:            make(chan gotime.Time, 1),
		Err:          make(chan error, 1),
	}

	helper.EnableExpiredNKE(ctx, client)

	go func() {
		res.client.Receive(ctx,
			res.client.B().Psubscribe().Pattern(res.key.KeySpacePattern()).Build(),
			func(m rueidis.PubSubMessage) {
				res.C <- gotime.Now()
				res.set(ctx)
			})
	}()

	if !res.exists(ctx) {
		res.set(ctx)
	}

	return &res, nil
}

func (t *Ticker) Reset(d model.TTL) {
	t.duration = d

	t.set(context.Background())
}

func (t *Ticker) Close() {
	defer t.client.Close()
	defer t.clientCancel()
	defer t.ctxCa()
	defer close(t.C)

}

func (t *Ticker) exists(ctx context.Context) bool {
	res, err := t.client.Do(ctx,
		t.client.B().
			Exists().
			Key(t.key.String()).
			Build()).AsBool()

	if err != nil {
		t.Err <- err

		return false
	}

	return res
}

func (t *Ticker) set(ctx context.Context) {
	res := t.client.Do(ctx,
		t.client.B().Set().
			Key(t.key.String()).
			Value(t.duration.String()).
			ExSeconds(t.duration.SecondsI64()).
			Build())

	if res.Error() != nil {
		t.Err <- res.Error()
	}
}
