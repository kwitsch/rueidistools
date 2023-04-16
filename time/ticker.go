package time

import (
	"errors"
	gotime "time"

	"github.com/kwitsch/rueidistools/model"
	"github.com/rueian/rueidis"
)

type Ticker struct {
	key          *model.Key
	client       rueidis.DedicatedClient
	clientCancel func()
	duration     model.TTL
	C            chan gotime.Time
}

func NewTicker(d gotime.Duration, name, prefix string, client rueidis.Client) (*Ticker, error) {
	ttl := model.TTL(d)
	if ttl.SecondsUI32() == 0 {
		return nil, errors.New("the ticker duration hast to be at least one second")
	}

	dClient, dcCancel := client.Dedicate()
	channel := make(chan gotime.Time, 1)
	res := Ticker{
		key:          model.NewKey(prefix, name),
		client:       dClient,
		clientCancel: dcCancel,
		duration:     ttl,
		C:            channel,
	}

	return &res, nil
}

func (t *Ticker) Close() {
	defer t.client.Close()
	defer t.clientCancel()
	defer close(t.C)
}
