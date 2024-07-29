package models

import (
	"database/sql"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func GetConn() *gorm.DB {
    if db != nil {
        return db
    }

    sqlDB, err := sql.Open("pgx", os.Getenv("DB_URL"))
    if err != nil {
        log.Fatal("failed to connect to database")
    }
    db, err = gorm.Open(postgres.New(postgres.Config{
        Conn: sqlDB,
    }), &gorm.Config{})
    if err != nil {
        log.Fatal("failed to create connection")
    }
    return db
}

func CloseConn() {
    if db != nil {
        sqlDB, err := db.DB()
        if err != nil {
            log.Fatal("failed to close connection")
        }
        sqlDB.Close()
    }
}
