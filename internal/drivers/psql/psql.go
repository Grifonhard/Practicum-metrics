package psql

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5"
)

type DB struct {
	*sql.DB
}

func ConnectDB(host, user, password, dbname string) (*DB, error) {
	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, `video`, `XXXXXXXX`, `video`)
	db, err := sql.Open("pgx", ps)
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
