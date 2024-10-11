package psql

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5"
)

type DB struct {
	*sql.DB
}

func ConnectDB(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &DB{
		DB: db,
	}, nil
}

func (db *DB) PingDB() error {
	return db.DB.Ping()
}

func (db *DB) Close() error {
	return db.DB.Close()
}
