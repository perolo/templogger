# templogger
Simple temperature logger that reads temperatures from ds18b20 sensors and writes to a database

## How to use
Reads a properties file templogger.properties, override with --prop filename.properties

* dbfile - string: Name and path to sqlite file database
* interval - string: Interval between reads in s
* sensornames - string: Names of sensors, comma separated (not used - just sanity check)
* expectedsensors - string: Device names , comma separated (not used - just sanity check)
* sensorcalibration - float: Sensor calibration values, comma separated

## Build
`
go build templogger.go
`
## Start
`
nohup ./templogger &
`

## Using TempServer
Provides a way to retrieve data through rest API: 
https://github.com/perolo/tempserver

Gui graphs very rudimetal - work remains...

## Reading sensors
Using https://github.com/yryz/ds18b20





