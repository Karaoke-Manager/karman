package db

import (
	"github.com/psanford/memfs"
	"gorm.io/gorm"
)

var db *gorm.DB

func Set(v *gorm.DB) {
	db = v
}

func Get() *gorm.DB {
	return db
}

var FS = memfs.New()
