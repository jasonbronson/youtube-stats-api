package main

import (
	"log"
	"time"

	"go.etcd.io/bbolt"
)


var (
	DB *bbolt.DB
)

func init(){

	db, err := bbolt.Open("./database.db", 0666, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	stats := db.Stats()
	log.Println("Stats:", stats)
	DB = db
	//defer DB.Close()

}