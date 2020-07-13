package main

import (
	"flag"
	"fmt"
	"github.com/jaypipes/ghw"
	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/worldiety/logdb"
	"github.com/worldiety/logdb/benchmark"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type processResult struct {
	histogram [256]uint64
	total     uint64
	duration  time.Duration
}

func (p processResult) rowsPerSec() float64 {
	return float64(p.total) / p.duration.Seconds()
}

func main() {
	dir := os.TempDir()
	tmpFile := filepath.Join(dir, "sensor.logdb")
	fmt.Printf("database file is '%s'\n", tmpFile)

	concurrency := flag.Int("p", 2, "amount of go routines")
	usemmap := flag.Bool("mmap", false, "mmap the entire file instead of pread")
	comp := flag.Bool("lz4", false, "lz4 compression on")
	doBench := flag.Bool("bench", false, "perform a benchmark and loop for each thread count and generate report")
	flag.Parse()

	if !*doBench {
		if _, err := scanTable(tmpFile, *concurrency, *usemmap, *comp); err != nil {
			panic(err)
		}
		return
	}

	if err := doBenchmark(tmpFile, *concurrency, *usemmap, *comp); err != nil {
		panic(err)
	}
}

func doBenchmark(tmpFile string, maxRoutines int, useMMap, useComp bool) error {
	const MAX_BEST_OF = 3

	fmt.Println("performing benchmark, please wait...")
	md := &strings.Builder{}

	runsPerThread := make([]processResult, (maxRoutines+1)*2)

	for gCount := 1; gCount <= maxRoutines*2; gCount++ {
		var bestRun *processResult

		for i := 0; i < MAX_BEST_OF; i++ {
			res, err := scanTable(tmpFile, gCount, useMMap, useComp)
			if err != nil {
				return err
			}
			if bestRun == nil || bestRun.duration.Nanoseconds() > res.duration.Nanoseconds() {
				bestRun = res
			}
		}

		runsPerThread[gCount] = *bestRun
	}

	md.WriteString(fmt.Sprintf("# benchmark results\n"))
	// print timing
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Scan speed, in million rows per second"
	p.X.Label.Text = "number of goroutines"
	p.Y.Label.Text = "rps"

	pts := make(plotter.XYs, 0)
	for gCount := 1; gCount <= maxRoutines*2; gCount++ {
		pts = append(pts, plotter.XY{
			X: float64(gCount),
			Y: runsPerThread[gCount].rowsPerSec() / 1_000_000,
		})
	}

	err = plotutil.AddLinePoints(p,
		"rows per second", pts)
	if err != nil {
		panic(err)
	}
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "benchmark.svg"); err != nil {
		panic(err)
	}

	md.WriteString(fmt.Sprintf("![rps-plot](./benchmark.svg)\n\n"))

	md.WriteString(fmt.Sprintf("Table: scan speed depending on spawned goroutines\n"))
	md.WriteString(fmt.Sprintf("|goroutines|rows per second|duration|total rows\n"))
	md.WriteString(fmt.Sprintf("|-----------|-----|----|----|\n"))
	for gCount := 1; gCount <= maxRoutines*2; gCount++ {
		md.WriteString(fmt.Sprintf("|%d|%.2f|%v|%d\n", gCount, runsPerThread[gCount].rowsPerSec(), runsPerThread[gCount].duration, runsPerThread[gCount].total))
	}
	md.WriteString("\n")

	md.WriteString(fmt.Sprintf("## node specification\n\n"))
	hname, _ := os.Hostname()
	md.WriteString(fmt.Sprintf("The benchmark has been run on '%s', which has the following specification:\n\n", hname))

	mem, err := ghw.Memory()
	if err != nil {
		mem = &ghw.MemoryInfo{
			TotalPhysicalBytes: int64(memory.TotalMemory()),
		}
	}
	md.WriteString(fmt.Sprintf("Table: memory specification\n"))
	md.WriteString(fmt.Sprintf("|Property|Value|\n"))
	md.WriteString(fmt.Sprintf("|-----------|-----|\n"))
	md.WriteString(fmt.Sprintf(json2MarkdownTable(toJson(mem))))

	md.WriteString(fmt.Sprintf("Table: cpu specification\n"))
	md.WriteString(fmt.Sprintf("|Property|Value|\n"))
	md.WriteString(fmt.Sprintf("|-----------|-----|\n"))
	cpuinfo, err := ghw.CPU()
	if err != nil {
		stat, err := cpu.Info()
		if err != nil {
			return err
		}
		for _, sStat := range stat {
			md.WriteString(fmt.Sprintf(json2MarkdownTable(toJson(sStat))))
		}
	} else {
		md.WriteString(fmt.Sprintf(json2MarkdownTable(toJson(cpuinfo))))
	}

	md.WriteString(fmt.Sprintf("Table: storage specification\n"))

	blockInfo, err := ghw.Block()
	if err != nil {
		parts, _ := disk.Partitions(false)
		for _, part := range parts {
			md.WriteString(fmt.Sprintf("|Property|Value|\n"))
			md.WriteString(fmt.Sprintf("|-----------|-----|\n"))
			md.WriteString(fmt.Sprintf(json2MarkdownTable(toJson(part))))
			md.WriteString(fmt.Sprintf("\n"))
			stat, _ := disk.Usage(part.Mountpoint)
			if stat != nil {
				fmt.Println(json2MarkdownTable(toJson(stat)))
			}
			fmt.Println()
		}
	} else {
		for _, disk := range blockInfo.Disks {
			md.WriteString(fmt.Sprintf("|Property|Value|\n"))
			md.WriteString(fmt.Sprintf("|-----------|-----|\n"))
			md.WriteString(fmt.Sprintf(json2MarkdownTable(toJson(disk))))
			md.WriteString(fmt.Sprintf("\n"))
		}
	}

	// print histogram
	md.WriteString(fmt.Sprintf("## histogram data of test set\n\n"))
	md.WriteString(fmt.Sprintf("The task was to calculate a histogram from the test set. In total, %d rows have been scanned.\n", runsPerThread[1].total))
	md.WriteString(fmt.Sprintf("|Temperature|Count|\n"))
	md.WriteString(fmt.Sprintf("|-----------|-----|\n"))
	for k, v := range runsPerThread[1].histogram {
		md.WriteString(fmt.Sprintf("|%d|%d|\n", int8(k), v))
	}

	ioutil.WriteFile("benchmark.md", []byte(md.String()), os.ModePerm)

	return nil
}

func scanTable(fname string, concurrency int, mmap, compression bool) (*processResult, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	debug.SetGCPercent(-1)
	defer debug.SetGCPercent(100)

	db, err := logdb.Open(fname, mmap, compression)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	colSensorId := uint16(db.IndexByName("SensorId"))
	colTimestamp := uint16(db.IndexByName("Timestamp"))
	colTemperature := uint16(db.IndexByName("Temperature"))

	progress := benchmark.NewProgress(db.ObjectCount(), 100_000)

	debug.SetGCPercent(-1)

	threadLocals := make([]*processResult, concurrency)
	for i := range threadLocals {
		threadLocals[i] = new(processResult) // this 25% faster (30m/40m than value type for 2 threads, probably to close in cache line?)
	}

	start := time.Now()
	err = db.ForEachP(concurrency, func(gid int, id uint64, obj *logdb.Object) error {
		var point benchmark.TemperaturePoint
		threadLocal := threadLocals[gid]

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

		threadLocal.histogram[int(uint8(point.Temperature))]++
		threadLocal.total++

		return nil
	})

	stop := time.Now()
	progress.Done()

	if err != nil {
		return nil, fmt.Errorf("failed to forEach: %w", err)
	}

	overallResult := processResult{}
	overallResult.duration = stop.Sub(start)

	for _, threadLocal := range threadLocals {
		overallResult.total += threadLocal.total
		for i, p := range threadLocal.histogram {
			overallResult.histogram[i] += p
		}
	}

	//progress.PrintStatistics()

	fmt.Printf("values with 0: %d\n", overallResult.histogram[0])
	if overallResult.total == 1_000_000_000 {
		if overallResult.histogram[0] != 3921916 {
			return nil, fmt.Errorf("result validation failed")
		}
		fmt.Printf("result validation passed\n")
	}

	/**
	the min value is {SensorId:826995 Timestamp:1594204572 Temperature:-128}
	the max value is {SensorId:509454 Timestamp:1594205061 Temperature:127}
	values with 0: 3921916
	*/

	return &overallResult, nil
}
