package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"net"
	"time"
)

var createStmt = `
	create table if not exists banners(
	    id integer primary key autoincrement ,
		ip varchar(32) not null,
		port varchar(4) not null,
		proto varchar(10) not null,
		added timestamp not null
);
`

var addStmt = `
	insert into banners(ip, port, proto, added) values (?, ?, ?, ?);
`

type Sqlite struct {
	db *sql.DB
}

func NewSqlite(path string) (*Sqlite, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the Sqlite db file")
	}
	return &Sqlite{db: db}, nil
}

func (s *Sqlite) prepare() error {
	_, err := s.db.Exec(createStmt)
	if err != nil {
		return errors.Wrap(err, "failed to create tables")
	}
	return nil
}

func (s *Sqlite) SaveBanner(ip net.IP, port uint16, proto string) error {
	_, err := s.db.Exec(addStmt, ip.String(), port, proto, time.Now().Second())
	if err != nil {
		return errors.Wrap(err, "failed to insert new data")
	}
	return nil
}
