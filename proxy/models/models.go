package models

type File struct {
    ID uint `gorm:"primaryKey,autoIncrement"`
    Name string
    Metadata string
}
