package models

type Instance struct {
    ID uint `gorm:"primaryKey,autoIncrement"`
    Path string
    Healthy bool
}
