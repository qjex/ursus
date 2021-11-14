package storage

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var dbPath = "/tmp/ursus.db"

func TestSqlite_SaveBanner(t *testing.T) {
	s := prep(t)
	err := s.prepare()
	require.NoError(t, err)
	type args struct {
		ip    net.IP
		port  uint16
		proto string
	}
	tests := []args{
		{
			ip:    net.ParseIP("129.192.123.42"),
			port:  0,
			proto: "socks5",
		},
		{
			ip:    net.ParseIP("129.192.123.42"),
			port:  0,
			proto: "socks5",
		},
		{
			ip:    net.ParseIP("129.192.123.42"),
			port:  0,
			proto: "socks4",
		},
		{
			ip:    net.ParseIP("129.192.123.42"),
			port:  2222,
			proto: "socks4",
		},
		{
			ip:    net.ParseIP("129.192.123.1"),
			port:  2222,
			proto: "socks4",
		},
		{
			ip:    net.ParseIP("129.192.123.111"),
			port:  65000,
			proto: "http",
		},
	}
	for _, tt := range tests {
		t.Run("save banner", func(t *testing.T) {
			if err := s.SaveBanner(tt.ip, tt.port, tt.proto); err != nil {
				t.Errorf("SaveBanner() error")
			}
		})
	}
}

func TestSqlite_ConcurrentReadOnly(t *testing.T) {
	s := prep(t)
	err := s.prepare()
	require.NoError(t, err)
	type args struct {
		ip    net.IP
		port  uint16
		proto string
	}
	tests := []args{
		{
			ip:    net.ParseIP("129.192.123.42"),
			port:  0,
			proto: "socks4",
		},
		{
			ip:    net.ParseIP("129.192.123.42"),
			port:  2222,
			proto: "socks4",
		},
	}
	for _, tt := range tests {
		err := s.SaveBanner(tt.ip, tt.port, tt.proto)
		require.NoError(t, err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	defer db.Close()
	require.NoError(t, err)
	rows, err := db.Query("select count(*) as cnt from banners")
	require.NoError(t, err)

	assert.Equal(t, true, rows.Next())
	var cnt int
	err = rows.Scan(&cnt)
	require.NoError(t, err)
	assert.Equal(t, 2, cnt)
}

func TestSqlite_prepare(t *testing.T) {
	s := prep(t)

	err := s.prepare()
	require.NoError(t, err)

	_, err = s.db.Exec("select * from banners;")
}

func prep(t *testing.T) *Sqlite {
	_ = os.Remove(dbPath)
	s, err := NewSqlite(dbPath)
	require.NoError(t, err)
	return s
}
