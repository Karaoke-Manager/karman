package db

import (
	"gorm.io/gorm"
)

var db *gorm.DB

func Set(v *gorm.DB) {
	db = v
}

func Get() *gorm.DB {
	return db
}
