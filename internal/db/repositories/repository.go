package repositories

import "github.com/go-pg/pg/v10"

type repository struct {
	db *pg.DB
}
