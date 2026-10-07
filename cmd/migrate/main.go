package main

import (
    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "dosage_drugs/internal/app/ds"
    "dosage_drugs/internal/app/dsn"
)

func main() {
    _ = godotenv.Load()
    db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    err = db.AutoMigrate(&ds.User{}, &ds.Drug{}, &ds.Like{})
    if err != nil {
        panic("cant migrate db")
    }

    var count int64
    db.Model(&ds.User{}).Where("login = ?", "test").Count(&count)
    if count == 0 {
        db.Create(&ds.User{Login: "test", Password: "test123", IsModerator: false})
    }
}
