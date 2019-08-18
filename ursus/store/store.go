package store

import (
	"net"
	"time"
)

type Store interface {
	Save(proxy Proxy) error
	Close()
}

type Proxy struct {
	Addr    net.IP    `json:"addr"`
	Port    uint16    `json:"port"`
	Proto   string    `json:"proto"`
	Updated time.Time `json:"updated"`
}
