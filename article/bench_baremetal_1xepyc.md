echo off > /sys/devices/system/cpu/smt/control

go/bin/reader -lz4=false -p=25 -mmap
database file is '/tmp/sensor.logdb'
mmap=true compression=false
db size is 23016717057 (21950 MiB)
names: 3
objects: 1000000000
last transaction: 352
found 352 records
-------
100%
processed 1000000000 elements in 1.596613832s
processing duration per element was 1ns
processed 626325527.16 elements per second (626 million elements per second)

the min value is {SensorId:0 Timestamp:0 Temperature:127}
the max value is {SensorId:0 Timestamp:0 Temperature:-128}


perf stat -d -d -d -r 10 go/bin/reader -lz4=false -p=25 -mmap

 Performance counter stats for 'go/bin/reader -lz4=false -p=25 -mmap' (10 runs):

         31,268.42 msec task-clock                #   17.880 CPUs utilized            ( +-  0.05% )
             6,486      context-switches          #    0.207 K/sec                    ( +-  3.79% )
             1,198      cpu-migrations            #    0.038 K/sec                    ( +-  3.90% )
           379,886      page-faults               #    0.012 M/sec                    ( +-  0.02% )
   104,242,982,245      cycles                    #    3.334 GHz                      ( +-  0.04% )  (32.73%)
       254,210,499      stalled-cycles-frontend   #    0.24% frontend cycles idle     ( +-  5.43% )  (32.90%)
    65,741,874,153      stalled-cycles-backend    #   63.07% backend cycles idle      ( +-  0.10% )  (33.18%)
   280,399,455,767      instructions              #    2.69  insn per cycle         
                                                  #    0.23  stalled cycles per insn  ( +-  0.07% )  (33.48%)
    38,715,027,157      branches                  # 1238.151 M/sec                    ( +-  0.07% )  (33.79%)
         6,791,474      branch-misses             #    0.02% of all branches          ( +-  0.57% )  (33.95%)
   163,419,959,502      L1-dcache-loads           # 5226.358 M/sec                    ( +-  0.10% )  (33.93%)
        89,152,279      L1-dcache-load-misses     #    0.05% of all L1-dcache hits    ( +-  1.33% )  (33.80%)
   <not supported>      LLC-loads                                                   
   <not supported>      LLC-load-misses                                             
     1,535,224,499      L1-icache-loads           #   49.098 M/sec                    ( +-  0.31% )  (33.65%)
         2,810,650      L1-icache-load-misses     #    0.18% of all L1-icache hits    ( +-  0.27% )  (33.49%)
         9,209,062      dTLB-loads                #    0.295 M/sec                    ( +-  1.42% )  (33.32%)
         5,753,549      dTLB-load-misses          #   62.48% of all dTLB cache hits   ( +-  1.89% )  (33.17%)
               461      iTLB-loads                #    0.015 K/sec                    ( +-  7.16% )  (33.01%)
             1,188      iTLB-load-misses          #  257.84% of all iTLB cache hits   ( +- 14.75% )  (32.87%)
       401,677,593      L1-dcache-prefetches      #   12.846 M/sec                    ( +-  0.09% )  (32.72%)
   <not supported>      L1-dcache-prefetch-misses                                   

            1.7488 +- 0.0130 seconds time elapsed  ( +-  0.74% )
            

1:   32366962.52
2:   64212172.63
3:   33406853.36
4:   38683059.94
5:  143046281.92
6:  181394095.65
7:  199536533.66
8:  229788300.59
9:  227708720.90
10: 280750055.14
11: 284946676.33
12: 315084572.58
13: 321479160.49
14: 325806629.81
15: 321043431.11
16: 340895091.78
17: 299849267.30
18: 326150896.14
19: 316354590.88
20: 324026235.74
21: 301448213.56
22: 582884365.58
23: 502317237.64
24: 297609497.01
25: 551395945.54
26: 305906749.48
27: 642000910.37
28: 374583274.89
29: 481704119.58
30: 284638320.98
31: 329570077.64
32: 403514577.10