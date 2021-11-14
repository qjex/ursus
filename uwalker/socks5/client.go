package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	VersionByte    = 0x05
	cmdConnect     = 0x01
	addrTypeDomain = 0x03
)

var NoAuthHeader = []byte{VersionByte, 0x01, 0x00}

type Client struct {
	net.Conn
	handshakeDone bool
	connected     bool
	closed        bool
}

func (sc *Client) handShake() error {
	if sc.handshakeDone {
		return nil
	}

	// only no-password auth
	if _, err := sc.Conn.Write(NoAuthHeader); err != nil {
		return err
	}

	buf := make([]byte, 512)

	if _, err := io.ReadFull(sc.Conn, buf[:2]); err != nil {
		return err
	}

	if buf[0] != VersionByte {
		return fmt.Errorf("error socks version %d", buf[0])
	}

	if buf[1] != 0x00 && buf[1] != 0x02 {
		return fmt.Errorf("server return with code %d", buf[1])
	}

	if buf[1] == 0x00 {
		sc.handshakeDone = true
		return nil
	}

	return fmt.Errorf("rejected")
}

// Dial to the addr from socks server,
// this is net.Dial style,
// can call sc.Connect instead
func (sc *Client) Dial(network, addr string) (net.Conn, error) {
	switch network {
	case "tcp":
	default:
		return nil, fmt.Errorf("unsupported network type: %s", network)
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	if err = sc.Connect(host, uint16(p)); err != nil {
		return nil, err
	}
	return sc, nil
}

// Connect handshakes with the socks server and request the
// server to connect to the target host and port
func (sc *Client) Connect(host string, port uint16) error {
	if !sc.handshakeDone {
		if err := sc.handShake(); err != nil {
			return err
		}
	}

	if sc.connected {
		return fmt.Errorf("only one connection allowed")
	}

	buf := make([]byte, 512)

	l := 4 + len(host) + 1 + 2
	buf[0] = VersionByte
	buf[1] = cmdConnect
	buf[2] = 0x00
	buf[3] = addrTypeDomain
	buf[4] = byte(len(host))

	copy(buf[5:5+len(host)], host)

	binary.BigEndian.PutUint16(buf[l-2:l], port)

	if _, err := sc.Conn.Write(buf[:l]); err != nil {
		return err
	}

	if _, err := io.ReadAtLeast(sc.Conn, buf, 10); err != nil {
		return err
	}

	if buf[0] != VersionByte {
		return fmt.Errorf("error socks version %d", buf[0])
	}

	if buf[1] != 0x00 {
		return fmt.Errorf("server error code %d", buf[1])
	}

	sc.connected = true
	return nil
}

func (sc *Client) Read(b []byte) (int, error) {
	if !sc.connected {
		return 0, fmt.Errorf("call connect first")
	}
	return sc.Conn.Read(b)
}

func (sc *Client) Write(b []byte) (int, error) {
	if !sc.connected {
		return 0, fmt.Errorf("call connect first")
	}
	return sc.Conn.Write(b)
}

func (sc *Client) Close() error {
	if !sc.closed {
		sc.closed = true
		return sc.Conn.Close()
	}
	return nil
}
