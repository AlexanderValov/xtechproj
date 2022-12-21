package repository

import (
	"XTechProject/internal/models"
	"XTechProject/pkg/postgres"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository struct {
	driver *postgres.Postgres
}

func New(driver *postgres.Postgres) *Repository {
	r := &Repository{driver: driver}
	r.CreateTablesIfTheyNotExist()
	return r
}

type Repositorier interface {
	CreateBTCRecord(model *models.BTC) error
	UpdateLastRecordForBTC() error
	GetLastBTC() (*models.BTC, error)
	GetAllBTC(limit, offset int, orderBy string) ([]models.BTC, error)
	UpdateFiatForLastBTC(model *models.BTC) error

	GetLastFiat() (*models.Fiat, error)
	GetAllFiat(limit, offset int, orderBy string) ([]models.Fiat, error)
	CreateFiatRecord(model *models.Fiat) error
	SetAllRecordsFiatLatestFalse() error
	GetLastDateForFiat() (*time.Time, error)
}

func (r *Repository) CreateTablesIfTheyNotExist() {
	r.driver.DB.Exec(`CREATE TABLE if not exists fiat
	(
		id         bigserial              	primary key,
		currencies jsonb                    not null,
		usd_rub    decimal(8, 4)            not null,
		created_at timestamp with time zone not null,
		latest     boolean                  not null
	);`)
	r.driver.DB.Exec(`CREATE TABLE if not exists bitcoin
	(
		id                bigserial                primary key,
		created_at 		  timestamp with time zone not null,
		in_usdt           decimal(10, 1)           not null,
		in_rub            decimal(12, 5)           not null,
		latest            boolean                  not null,
		btc_to_fiat       jsonb                    
	);`)
}

func (r *Repository) CreateBTCRecord(model *models.BTC) error {
	query := `
	INSERT INTO bitcoin (in_usdt, created_at, latest, in_rub, btc_to_fiat) 
	VALUES (:in_usdt, :created_at, :latest, :in_rub, :btc_to_fiat)`
	_, err := r.driver.DB.NamedExec(query, model)
	return err
}

func (r *Repository) UpdateFiatForLastBTC(model *models.BTC) error {
	query := `UPDATE bitcoin SET in_rub=:in_rub, btc_to_fiat=:btc_to_fiat WHERE latest = true`
	_, err := r.driver.DB.NamedExec(query, model)
	return err
}

func (r *Repository) UpdateLastRecordForBTC() error {
	query := `UPDATE bitcoin SET latest = false WHERE latest = true`
	_, err := r.driver.DB.Exec(query)
	return err
}

func (r *Repository) CreateFiatRecord(model *models.Fiat) error {
	query := `
	INSERT INTO fiat (currencies, latest, usd_rub, created_at)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`
	_, err := r.driver.DB.Exec(query, model.Currencies, model.Latest, model.USDRUB)
	return err
}

func (r *Repository) SetAllRecordsFiatLatestFalse() error {
	query := `UPDATE fiat SET latest = false  WHERE latest = true`
	_, err := r.driver.DB.Exec(query)
	return err
}

func (r *Repository) GetLastDateForFiat() (*time.Time, error) {
	var date time.Time
	query := `SELECT created_at FROM fiat WHERE latest = true`
	err := r.driver.DB.Get(&date, query)
	if err != nil {
		// OK if there is no date
		if errors.Is(sql.ErrNoRows, err) {
			return nil, nil
		}
	}
	return &date, err
}

func (r *Repository) GetLastBTC() (*models.BTC, error) {
	query := `SELECT * FROM bitcoin WHERE latest = true`
	var btc models.BTC
	err := r.driver.DB.Get(&btc, query)
	return &btc, err
}

func (r *Repository) GetLastFiat() (*models.Fiat, error) {
	query := `SELECT * FROM fiat WHERE latest = true`
	var fiat models.Fiat
	err := r.driver.DB.Get(&fiat, query)
	return &fiat, err
}

func (r *Repository) GetAllBTC(limit, offset int, orderBy string) ([]models.BTC, error) {
	var btc []models.BTC
	var err error
	var query string
	if limit == 0 && offset == 0 {
		query = fmt.Sprintf("SELECT * FROM bitcoin %s;", orderBy)
		err = r.driver.DB.Select(&btc, query)
	} else if limit != 0 && offset == 0 {
		query = fmt.Sprintf("SELECT * FROM bitcoin %s LIMIT $1;", orderBy)
		err = r.driver.DB.Select(&btc, query, limit)
	} else if limit == 0 && offset != 0 {
		query = fmt.Sprintf("SELECT * FROM bitcoin %s OFFSET $1;", orderBy)
		err = r.driver.DB.Select(&btc, query, offset)
	} else {
		query = fmt.Sprintf("SELECT * FROM bitcoin %s LIMIT $1 OFFSET $2;", orderBy)
		err = r.driver.DB.Select(&btc, query, limit, offset)
	}
	return btc, err
}

func (r *Repository) GetAllFiat(limit, offset int, orderBy string) ([]models.Fiat, error) {
	var fiat []models.Fiat
	var err error
	var query string
	if limit == 0 && offset == 0 {
		query = fmt.Sprintf("SELECT * FROM fiat %s;", orderBy)
		err = r.driver.DB.Select(&fiat, query)
	} else if limit != 0 && offset == 0 {
		query = fmt.Sprintf("SELECT * FROM fiat %s LIMIT $1;", orderBy)
		err = r.driver.DB.Select(&fiat, query, limit)
	} else if limit == 0 && offset != 0 {
		query = fmt.Sprintf("SELECT * FROM fiat %s OFFSET $1;", orderBy)
		err = r.driver.DB.Select(&fiat, query, offset)
	} else {
		query = fmt.Sprintf("SELECT * FROM fiat %s LIMIT $1 OFFSET $2;", orderBy)
		err = r.driver.DB.Select(&fiat, query, limit, offset)
	}
	return fiat, err
}
