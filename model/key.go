package model

import (
	"fmt"
	"strings"
)

type Key struct {
	key          string
	spacePattern string
}

func NewKey(parts ...string) *Key {
	res := Key{
		key: strings.Join(parts, ":"),
	}

	res.spacePattern = fmt.Sprintf("__keyspace@?__:%s", res.key)

	return &res
}

func (k *Key) Key() string {
	return k.key
}

func (k *Key) KeySpacePattern() string {
	return k.spacePattern
}
