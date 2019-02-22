package storage

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("Data Not Found")

type Storage struct {
	pgURL  string
	db     *sql.DB
	errLog *log.Logger
}

func Open(pgURL string, errLog *log.Logger) (*Storage, error) {
	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		errLog.Println(err)
		return nil, err
	}

	return &Storage{
		pgURL:  pgURL,
		db:     db,
		errLog: errLog,
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
