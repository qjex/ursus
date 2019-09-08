package store

import (
	"context"
	"net"
	"time"
)

type ProxyStore interface {
	Save(ctx context.Context, proxy Proxy) error
	FindAll(ctx context.Context, page, pageSize int64) ([]Proxy, error)
	Close()
}

type Proxy struct {
	Addr    net.IP    `json:"addr"`
	Port    uint16    `json:"port"`
	Proto   string    `json:"proto"`
	Updated time.Time `json:"updated"`
}
