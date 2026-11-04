package main

import (
	"errors"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

	if err = db.AutoMigrate(&ds.User{}, &ds.Dosage_drugs{}, &ds.Like{}); err != nil {
		panic("cant migrate db")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	if err != nil {
		panic("cant hash password")
	}

	// Тестовый пользователь test / test123. Если он остался с ЛР3 с паролем в открытом
	// виде — заменяем пароль на bcrypt-хеш, иначе вход не сработает.
	var u ds.User
	err = db.Where("login = ?", "test").First(&u).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		db.Create(&ds.User{Login: "test", Password: string(hash)})
	case err == nil:
		if !strings.HasPrefix(u.Password, "$2") {
			db.Model(&u).Update("password", string(hash))
		}
	default:
		panic(err)
	}
}
