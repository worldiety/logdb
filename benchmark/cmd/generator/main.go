package main

import (
	"flag"
	"fmt"
	"github.com/worldiety/logdb"
	"github.com/worldiety/logdb/benchmark"
	"math/rand"
	"os"
	"path/filepath"
)

func main() {
	compress := flag.Bool("lz4", true, "lz4 compression")
	flag.Parse()

	dir := os.TempDir()
	tmpFile := filepath.Join(dir, "sensor.logdb")
	fmt.Printf("database file is '%s'\n", tmpFile)

	if err := create(tmpFile, *compress); err != nil {
		panic(err)
	}
}

func create(fname string, compress bool) error {
	_ = os.Remove(fname)
	db, err := logdb.Open(fname, false, compress)
	if err != nil {
		return err
	}
	defer db.Close()

	colSensorId := db.PutName("SensorId")
	colTimestamp := db.PutName("Timestamp")
	colTemperature := db.PutName("Temperature")

	const largestSensorId = 1_000_000
	const pointsToGenerate = 1_000_000_000
	timestamp := uint32(1594204360)

	random := rand.New(rand.NewSource(1234))
	point := benchmark.TemperaturePoint{}

	progress := benchmark.NewProgress(pointsToGenerate, 100_000)

	for i := 0; i < pointsToGenerate; i++ {
		progress.Next()

		point.SensorId = uint32(random.Int31n(largestSensorId))
		point.Timestamp = timestamp
		point.Temperature = int8(rand.Intn(255))
		timestamp++

		err := db.Add(func(obj *logdb.Object) error {
			obj.AddUint32(colSensorId, point.SensorId)
			obj.AddUint32(colTimestamp, point.Timestamp)
			obj.AddInt8(colTemperature, point.Temperature)
			return nil
		})

		if err != nil {
			return err
		}
	}

	progress.PrintStatistics()

	return nil
}
