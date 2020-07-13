## node specification

|Property|Value|
|-----------|-----|
|cpu|0|
|vendor-id|GenuineIntel|
|model-name|Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz|
|family|6|
|stepping|9|
|MHz|2900|
|cores (physical)|4|
|cores (logical)|8|
|cache-size|256|
|clocksPerSec|100|
|flags|[fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clfsh ds acpi mmx fxsr sse sse2 ss htt tm pbe sse3 pclmulqdq dtes64 mon dscpl vmx smx est tm2 ssse3 fma cx16 tpr pdcm sse4.1 sse4.2 x2apic movbe popcnt aes pcid xsave osxsave seglim64 tsctmr avx1.0 rdrand f16c rdwrfsgs tsc_thread_offset sgx bmi1 hle avx2 smep bmi2 erms invpcid rtm fpu_csds mpx rdseed adx smap clfsopt ipt mdclear tsxfa ibrs stibp l1df ssbd syscall xd 1gbpage em64t lahf lzcnt prefetchw rdtscp tsci]|


echo off > /sys/devices/system/cpu/smt/control

root@ams1-m2:~# go/bin/reader -lz4=false -p=28 -mmap
database file is '/tmp/sensor.logdb'
mmap=true compression=false
db size is 46016722673 (43884 MiB)
names: 3
objects: 2000000000
last transaction: 703
found 703 records
-------
100%
processed 2000000000 elements in 3.063506162s
processing duration per element was 1ns
processed 652846736.46 elements per second (653 million elements per second)

the min value is {SensorId:0 Timestamp:0 Temperature:127}
the max value is {SensorId:0 Timestamp:0 Temperature:-128}
values with 0: 7845084


Performance counter stats for 'go/bin/reader -lz4=false -p=28 -mmap' (10 runs):

         78,816.06 msec task-clock                #   17.302 CPUs utilized            ( +-  1.39% )
            11,076      context-switches          #    0.141 K/sec                    ( +-  2.75% )
             3,860      cpu-migrations            #    0.049 K/sec                    ( +-  1.63% )
           735,780      page-faults               #    0.009 M/sec                    ( +-  0.03% )
   205,120,311,393      cycles                    #    2.603 GHz                      ( +-  1.67% )  (30.58%)
   564,668,200,422      instructions              #    2.75  insn per cycle           ( +-  0.05% )  (38.30%)
    77,833,122,450      branches                  #  987.529 M/sec                    ( +-  0.05% )  (38.39%)
        14,455,185      branch-misses             #    0.02% of all branches          ( +-  0.52% )  (38.49%)
   224,163,122,557      L1-dcache-loads           # 2844.130 M/sec                    ( +-  0.03% )  (38.61%)
       897,215,052      L1-dcache-load-misses     #    0.40% of all L1-dcache hits    ( +-  1.27% )  (38.67%)
        19,750,606      LLC-loads                 #    0.251 M/sec                    ( +- 14.14% )  (30.91%)
        11,812,898      LLC-load-misses           #   59.81% of all LL-cache hits     ( +- 11.88% )  (30.87%)
   <not supported>      L1-icache-loads                                             
        37,470,122      L1-icache-load-misses                                         ( +-  0.98% )  (30.81%)
   225,501,589,808      dTLB-loads                # 2861.112 M/sec                    ( +-  0.02% )  (30.76%)
         4,364,283      dTLB-load-misses          #    0.00% of all dTLB cache hits   ( +-  0.27% )  (30.70%)
           511,365      iTLB-loads                #    0.006 M/sec                    ( +-  6.77% )  (30.65%)
         1,212,811      iTLB-load-misses          #  237.17% of all iTLB cache hits   ( +-  3.72% )  (30.58%)
   <not supported>      L1-dcache-prefetches                                        
   <not supported>      L1-dcache-prefetch-misses                                   

             4.555 +- 0.633 seconds time elapsed  ( +- 13.90% )

1: 31795906.36
2: 61714357.16
3: 25722887.17
