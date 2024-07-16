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

    sqlDB, err := sql.Open("pgx", "postgresql://neondb_owner:7L3BNrHmRubP@ep-flat-heart-a2q17yp9.eu-central-1.aws.neon.tech/neondb?sslmode=require")
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
