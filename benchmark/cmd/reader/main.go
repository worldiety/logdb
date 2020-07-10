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
)

func main() {
	dir := os.TempDir()
	tmpFile := filepath.Join(dir, "sensor.logdb")
	fmt.Printf("database file is '%s'\n", tmpFile)

	concurrency := flag.Int("p", 1, "amount of go routines")
	usemmap := flag.Bool("mmap", false, "mmap the entire file instead of pread")
	flag.Parse()

	if err := scanTable(tmpFile, *concurrency, *usemmap); err != nil {
		panic(err)
	}
}

func scanTable(fname string, concurrency int, mmap bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	debug.SetGCPercent(-1)
	defer debug.SetGCPercent(100)

	db, err := logdb.Open(fname, mmap)
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

	progress := benchmark.NewProgress(db.ObjectCount(), 100_000)

	debug.SetGCPercent(-1)
	debug.SetMaxThreads(32)
	/*
		err = db.ForEach( func(id uint64, obj *logdb.Object) error {
			point := benchmark.TemperaturePoint{}
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


			if max.Temperature < point.Temperature {
				max = point
			}

			if min.Temperature > point.Temperature {
				min = point
			}

			if point.Temperature == 0 {
				//atomic.AddInt64(&countForZero, 1)
				countForZero++
			}
			return nil
		}) */ //23m

	type perThreadRes struct {
		zeroCount       int64
		point, min, max benchmark.TemperaturePoint
		pad             [20]byte // fill entire 64byte cache line per thread? seems to have no influence, doesn't work like this?
	}

	const cur = 8 //2 threads scales good, beyond not

	threadLocals := make([]*perThreadRes, cur)
	for i := range threadLocals {
		threadLocals[i] = new(perThreadRes) // this 25% faster (30m/40m than value type for 2 threads, probably to close in cache line?)
	}

	err = db.ForEachP(cur, func(gid int, id uint64, obj *logdb.Object) error {
		threadLocal := threadLocals[gid]

		obj.FieldReaderReset()
		for obj.FieldReaderNext() {
			name := obj.FieldReaderName()
			if name == colSensorId {
				threadLocal.point.SensorId = obj.FieldReader().ReadUint32()
			} else if name == colTimestamp {
				threadLocal.point.Timestamp = obj.FieldReader().ReadUint32()
			} else if name == colTemperature {
				threadLocal.point.Temperature = obj.FieldReader().ReadInt8()
			}
		}

		if threadLocal.max.Temperature < threadLocal.point.Temperature {
			threadLocal.max = threadLocal.point
		}

		if threadLocal.min.Temperature > threadLocal.point.Temperature {
			threadLocal.min = threadLocal.point
		}

		if threadLocal.point.Temperature == 0 {
			//atomic.AddInt64(&countForZero, 1)
			threadLocals[gid].zeroCount++
		}
		return nil
	})

	progress.Done()

	if err != nil {
		return fmt.Errorf("failed to forEach: %w", err)
	}

	for _, threadLocal := range threadLocals {
		countForZero += threadLocal.zeroCount
	}

	//progress.PrintStatistics()

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
