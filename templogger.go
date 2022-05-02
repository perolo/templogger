package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"
	"strings"
	"strconv"
	"github.com/magiconair/properties"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yryz/ds18b20"
)

var db *sql.DB
var cfg Config

type Config struct {
	DbFile   string `properties:"dbfile"`
	Interval int    `properties:"interval"`
	SensorNames string `properties:"sensornames"`
	ExpectedSensor string `properties:"expectedsensors"`
	SensorCalibration string `properties:"sensorcalibration"`
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
	_, err = statement.Exec()
	Check(err)
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
	
	sensorNames := strings.Split(cfg.SensorNames, ",")
	expectedSensors := strings.Split(cfg.ExpectedSensor, ",")
	sensorCalibration:= strings.Split(cfg.SensorCalibration, ",")
	calib := make(map[string]float64)

	initdb()

	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	fmt.Println("database is reachable")

	sensors, err := ds18b20.Sensors()
	if err != nil {
		panic(err)
	}

	if len(expectedSensors)!= len(sensorCalibration) {
		fmt.Printf("Expected equal number sensors and calibration, setting calibration to 0")
		for _, c := range sensors{
			calib[c] = 0	
		}
	} else {
		for k, c := range sensors{
			if c != expectedSensors[k] {
				fmt.Printf("Expected sensor:%s, found sensor: %s ", expectedSensors[k], c)
			}
		}
		for k, c := range sensorCalibration{
			x, err := strconv.ParseFloat(c, 64)
			if err != nil {
				panic(err)
			}
			calib[expectedSensors[k]] = x	
		} 	
	}

	fmt.Printf("sensor IDs: %v\n", sensorNames)
	fmt.Printf("sensor IDs: %v\n", sensorCalibration)
	fmt.Printf("sensor IDs: %v\n", sensors)
	
 	/*i:= 1 ; i<10;i++*/ 
	for{
		//theTime := time.Now().Format("2006-01-02-15:04:05")
		for key, sensor := range sensors {
			//fmt.Printf("Time: %s\n", theTime)
			t, err := ds18b20.Temperature(sensor)
			t = t - calib[sensor]
			if err == nil {
				//fmt.Printf("sensor: %s key: %v temperature: %.2f°C \n", sensor, key, t)
				saveToDatabase(key, t)
			}
		}
		time.Sleep(time.Duration(cfg.Interval) * time.Second)
	}
}
