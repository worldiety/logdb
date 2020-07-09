package main

import (
	"flag"
	"fmt"
	"github.com/worldiety/logdb"
	"github.com/worldiety/logdb/benchmark"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync/atomic"
)

func main() {
	dir := os.TempDir()
	tmpFile := filepath.Join(dir, "sensor.logdb")
	fmt.Printf("database file is '%s'\n", tmpFile)

	concurrency := flag.Int("p", 1, "amount of go routines")
	flag.Parse()

	if err := scanTable(tmpFile, *concurrency); err != nil {
		panic(err)
	}
}

func scanTable(fname string, concurrency int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	debug.SetGCPercent(-1)
	defer debug.SetGCPercent(100)

	db, err := logdb.Open(fname)
	if err != nil {
		return err
	}
	defer db.Close()

	colSensorId := uint16(db.IndexByName("SensorId"))
	colTimestamp := uint16(db.IndexByName("Timestamp"))
	colTemperature := uint16(db.IndexByName("Temperature"))

	max := benchmark.TemperaturePoint{
		Temperature: -128,
	}

	min := benchmark.TemperaturePoint{
		Temperature: 127,
	}

	countForZero := int64(0)

	point := benchmark.TemperaturePoint{}

	progress := benchmark.NewProgress(db.ObjectCount(), 100_000)

	err = db.ForEachP(concurrency, func(id uint64, obj *logdb.Object) error {
		obj.FieldReaderReset()
		for obj.FieldReaderNext() {
			name := obj.FieldReaderName()
			if name == colSensorId {
				point.SensorId = obj.FieldReader().ReadUint32()
			} else if name == colTimestamp {
				point.Timestamp = obj.FieldReader().ReadUint32()
			} else if name == colTemperature {
				point.Temperature = obj.FieldReader().ReadInt8()
			}
		}
		/*obj.WithFields(func(name uint16, kind ioutil.Type, f *logdb.FieldReader) {
			if name == colSensorId {
				point.SensorId = f.ReadUint32()
				return
			}

			if name == colTimestamp {
				point.Timestamp = f.ReadUint32()
				return
			}

			if name == colTemperature {
				point.Temperature = f.ReadInt8()
				return
			}
		})*/

		if max.Temperature < point.Temperature {
			max = point
		}

		if min.Temperature > point.Temperature {
			min = point
		}

		if point.Temperature == 0 {
			atomic.AddInt64(&countForZero, 1)
		}

		progress.Next()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to forEach: %w", err)
	}

	progress.PrintStatistics()

	fmt.Printf("the min value is %+v\n", min)
	fmt.Printf("the max value is %+v\n", max)
	fmt.Printf("values with 0: %d\n", countForZero)

	/**
	the min value is {SensorId:826995 Timestamp:1594204572 Temperature:-128}
	the max value is {SensorId:509454 Timestamp:1594205061 Temperature:127}
	values with 0: 3921916
	*/

	return nil
}
