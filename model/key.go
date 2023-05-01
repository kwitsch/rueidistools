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

func (k *Key) String() string {
	return k.key
}

func (k *Key) KeySpacePattern() string {
	return k.spacePattern
}

func (k *Key) NewSubkey(parts ...string) *Key {
	ip := []string{k.key}

	ip = append(ip, parts...)

	return NewKey(ip...)
}
