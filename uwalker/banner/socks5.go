package banner

import (
	"uwalker/socks5"
)

type Socks5 struct {
}

func (s *Socks5) Read(data []byte) ([]byte, int, bool) {
	if len(data) > 1 && data[0] == socks5.VersionByte && data[1] == 0x00 {
		return nil, len(data), true
	}
	return nil, 0, false
}

func (s *Socks5) Init() []byte {
	return socks5.NoAuthHeader
}

func (s *Socks5) Proto() string {
	return "socks5"
}
