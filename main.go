package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/magiconair/properties"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yryz/ds18b20"
)

var db *sql.DB
var cfg Config

type Reading struct {
	Id          int
	Sensor      int
	Temperature float64
}
type Config struct {
	DbFile   string `properties:"dbfile"`
	Interval int    `properties:"interval"`
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func initdb() {
	var err error
	db, err = sql.Open("sqlite3", cfg.DbFile)
	Check(err)
	statement, prepError := db.Prepare("CREATE TABLE IF NOT EXISTS reading (Id INTEGER PRIMARY KEY AUTOINCREMENT, Sensor INTEGER, Temperature NUMERIC, Datetime DATETIME)")
	Check(prepError)
	statement.Exec()
}

func saveToDatabase(Sensor int, Temperature float64) {

	statement, err := db.Prepare("INSERT INTO reading (Sensor, Temperature, Datetime) VALUES (?,?,CURRENT_TIMESTAMP)")
	Check(err)

	_, err = statement.Exec(Sensor, Temperature)
	Check(err)
}

func main() {

	propPtr := flag.String("prop", "templogger.properties", "a string")
	flag.Parse()

	p := properties.MustLoadFile(*propPtr, properties.ISO_8859_1)

	if err := p.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	initdb()
	calib := make(map[string]float64)
	sensors, err := ds18b20.Sensors()
	if err != nil {
		panic(err)
	}

	calib["28-1b61221e64ff"] = -0.12
	calib["28-7167221e64ff"] = -0.06
	calib["28-a4a8211e64ff"] = 0.06
	calib["28-039d231e64ff"] = 0.06
	calib["28-41b7231e64ff"] = 0.00
	calib["28-6d8f231e64ff"] = 0.00

	fmt.Printf("sensor IDs: %v\n", sensors)

	for {
		theTime := time.Now().Format("2006-01-02-15:04:05")
		for key, sensor := range sensors {
			fmt.Printf("Time: %s\n", theTime)
			t, err := ds18b20.Temperature(sensor)
			t = t - calib[sensor]
			if err == nil {
				fmt.Printf("sensor: %s key: %v temperature: %.2f°C \n", sensor, key, t)
				saveToDatabase(key, t)
			}
		}
		time.Sleep(time.Duration(cfg.Interval) * time.Second)
	}
}
