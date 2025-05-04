// internal/db/connect

package db

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // <-- tu blank-import sterownika
)

func MustConnect(databaseURL string) *sqlx.DB {
	var db *sqlx.DB
	var err error
	for i := 1; i <= 10; i++ {
		db, err = sqlx.Connect("postgres", databaseURL)
		if err == nil {
			return db
		}
		log.Printf("DB not ready (%d/10), retrying in 1s: %v", i, err)
		time.Sleep(1 * time.Second)
	}
	log.Fatalf("Could not connect to Postgres after retries: %v", err)
	return nil
}
