package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"
	"strings"
	"strconv"
	"runtime"
	"github.com/magiconair/properties"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yryz/ds18b20"
	"net/http"
	_"net/http/pprof"
	"github.com/pkg/profile"	
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
	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	fmt.Printf("%s: Database is reachable\n", time.Now().Format("2006-01-02-15:04:05"))

}

func closedb() {
	err := db.Close()
	Check(err)
}


func saveToDatabase(Sensor int, Temperature float64) {

	statement, err := db.Prepare("INSERT INTO reading (Sensor, Temperature, Datetime) VALUES (?,?,CURRENT_TIMESTAMP)")
	Check(err)

	_, err = statement.Exec(Sensor, Temperature)
	Check(err)
}

func main() {
	defer profile.Start(profile.MemProfile).Stop()

	go func() {
			http.ListenAndServe(":8080", nil)
	}()

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

	dbinitialized := false
	initTime := time.Now()

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
		if !dbinitialized {
			initdb()
			dbinitialized = true
			initTime = time.Now()
		}
		loopStart := time.Now()
		for key, sensor := range sensors {
			t, err := ds18b20.Temperature(sensor)
			t = t - calib[sensor]
			if err == nil {
				//fmt.Printf("sensor: %s key: %v temperature: %.2f°C \n", sensor, key, t)
				saveToDatabase(key, t)
			}
		}
		// TODO Workaround for Memory leak in database - find root cause
		if time.Since(initTime)> time.Duration(24*time.Hour) {
			closedb()
			time.Sleep(time.Duration(time.Second))
			runtime.GC()
			dbinitialized = false
		}
		loopTime := time.Since(loopStart)
		time.Sleep(time.Duration(cfg.Interval) * time.Second - loopTime)  //TODO Should be possible with better time drift control
	}
}
