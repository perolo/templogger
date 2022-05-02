package main

import (
	"fmt"
	"flag"
	"runtime"
	"os"
	"time"
	"log"
	"runtime/pprof"
	"github.com/perolo/ds18b20"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func sens() {
	fmt.Printf("Hello go! \n")
	
	sensors, err := ds18b20.Sensors()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Sensors: %s\n", sensors)
 	
	for i:= 1 ; i<10;i++ {
		//theTime := time.Now().Format("2006-01-02-15:04:05")
		for key, sensor := range sensors {
			//fmt.Printf("Time: %s\n", theTime)
			t, err := ds18b20.Temperature(sensor)
			if err == nil {
				fmt.Printf("sensor: %s key: %v temperature: %.2f°C \n", sensor, key, t)
			} else {
				fmt.Printf("sensor: %s key: %v err:%s \n", sensor, key, err)
			}
		}
		time.Sleep(time.Duration(time.Second))
	}

}

func main() {
	flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal("could not create CPU profile: ", err)
        }
        defer f.Close() // error handling omitted for example
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }
	
	sens()

    if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal("could not create memory profile: ", err)
        }
        defer f.Close() // error handling omitted for example
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            log.Fatal("could not write memory profile: ", err)
        }
    }

}
