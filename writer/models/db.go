package models

import (
	"database/sql"
	"log"

    "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func GetConn() *gorm.DB {
    if db != nil {
        return db
    }

    sqlConn, err := sql.Open("pgx", "http://localhost:5432")
    if err != nil {
        log.Println("failed to connect to db")
        return nil
    }
    db, err = gorm.Open(postgres.New(postgres.Config{
        Conn: sqlConn,
    }), &gorm.Config{})
    if err != nil {
        log.Println("failed to connect to db")
        return nil
    }

    return db
}
