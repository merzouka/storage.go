package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

const (
    host = "localhost"
    port = 5432
    user = "docker"
    dbname = "metadata"
)

func GetConn() *gorm.DB {
    password := os.Getenv("DB_PASSWORD")
    if db != nil {
        return db
    }
    connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable", user, password, host, port, dbname)
    sqlConn, err := sql.Open("pgx", connStr)
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
    log.Println("established connection to database")

    return db
}
