package models

type Metadata struct {
    ID uint `gorm:"primaryKey,autoIncrement"`
    Name string
    Metadata string
}
