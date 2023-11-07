package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

type Cambio struct {
	gorm.Model // Inclui campos como ID, CreatedAt, UpdatedAt, DeletedAt
	Code       string
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
}

var db *gorm.DB

func init() {
	// Conecta ao banco de dados SQLite
	var err error
	// db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db, err = gorm.Open(sqlite.Open("cambio.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Cambio{})
}
