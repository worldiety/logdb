```bash
perf stat -d -d -d -r 10 go/bin/reader -lz4=false -p=30 -mmap

 Performance counter stats for 'go/bin/reader -lz4=false -p=30 -mmap' (10 runs):

         81,558.64 msec task-clock                #    9.963 CPUs utilized            ( +-  1.87% )
            21,034      context-switches          #    0.258 K/sec                    ( +-  1.33% )
             1,690      cpu-migrations            #    0.021 K/sec                    ( +-  5.02% )
           387,908      page-faults               #    0.005 M/sec                    ( +-  0.06% )
   259,703,462,829      cycles                    #    3.184 GHz                      ( +-  1.92% )  (63.57%)
   309,524,814,929      instructions              #    1.19  insn per cycle           ( +-  0.03% )  (63.57%)
   <not supported>      branches                                                    
         6,823,482      branch-misses                                                 ( +-  2.98% )  (63.56%)
   160,482,453,757      L1-dcache-loads           # 1967.694 M/sec                    ( +-  0.28% )  (63.58%)
       513,806,569      L1-dcache-load-misses     #    0.32% of all L1-dcache hits    ( +-  6.24% )  (63.62%)
   <not supported>      LLC-loads                                                   
   <not supported>      LLC-load-misses                                             
    68,641,040,634      L1-icache-loads           #  841.616 M/sec                    ( +-  0.90% )  (63.65%)
        33,342,441      L1-icache-load-misses     #    0.05% of all L1-icache hits    ( +-  1.20% )  (63.70%)
        38,846,510      dTLB-loads                #    0.476 M/sec                    ( +-  0.44% )  (54.63%)
         6,370,239      dTLB-load-misses          #   16.40% of all dTLB cache hits   ( +-  0.29% )  (54.61%)
        12,251,463      iTLB-loads                #    0.150 M/sec                    ( +-  1.47% )  (54.57%)
           137,342      iTLB-load-misses          #    1.12% of all iTLB cache hits   ( +-  0.77% )  (54.51%)
   <not supported>      L1-dcache-prefetches                                        
   <not supported>      L1-dcache-prefetch-misses                                   

            8.1859 +- 0.0105 seconds time elapsed  ( +-  0.13% )


```