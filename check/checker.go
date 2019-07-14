package check

import (
	"net"
)

type Proxy struct {
	Addr net.IP
	Port int
	Type string
}

type Checker interface {
	Check(checkHost, testHost string) bool
	Type() string
}
