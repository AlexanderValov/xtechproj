package postgres

import (
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

type Postgres struct {
	DB *sqlx.DB
}

func NewPostgresDB(url string) (*Postgres, error) {
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	p := &Postgres{
		DB: db,
	}
	return p, nil
}
