package storage

import (
	"github.com/pkg/errors"
	"net"
)

type preparer interface {
	// prepare is called to init the engine (create tables, collections, etc.)
	prepare() error
}

type Engine interface {
	SaveBanner(ip net.IP, port uint16, proto string) error
	preparer
}

type Store struct {
	engine Engine
}

func NewStore(engine Engine) error {
	err := engine.prepare()
	if err != nil {
		return errors.Wrap(err, "error initializing the store")
	}
	return nil
}

func (s *Store) PersistBanner(ip net.IP, port uint16, proto string) error {
	return s.engine.SaveBanner(ip, port, proto)
}
