```bash
root@ams1-c2:~# perf stat -d -d -d -r 10 go/bin/reader -lz4=false -p=32 -mmap

 Performance counter stats for 'go/bin/reader -lz4=false -p=32 -mmap' (10 runs):

         83,451.36 msec task-clock                #   16.382 CPUs utilized            ( +-  1.63% )
            21,310      context-switches          #    0.255 K/sec                    ( +-  2.46% )
             2,253      cpu-migrations            #    0.027 K/sec                    ( +-  6.14% )
           388,994      page-faults               #    0.005 M/sec                    ( +-  0.07% )
   265,180,007,789      cycles                    #    3.178 GHz                      ( +-  1.72% )  (63.59%)
   309,444,239,209      instructions              #    1.17  insn per cycle           ( +-  0.03% )  (63.58%)
   <not supported>      branches                                                    
         7,220,838      branch-misses                                                 ( +-  7.13% )  (63.57%)
   160,959,818,605      L1-dcache-loads           # 1928.786 M/sec                    ( +-  0.24% )  (63.59%)
       563,240,044      L1-dcache-load-misses     #    0.35% of all L1-dcache hits    ( +-  6.27% )  (63.62%)
   <not supported>      LLC-loads                                                   
   <not supported>      LLC-load-misses                                             
    69,553,409,830      L1-icache-loads           #  833.460 M/sec                    ( +-  0.93% )  (63.64%)
        33,142,124      L1-icache-load-misses     #    0.05% of all L1-icache hits    ( +-  1.74% )  (63.66%)
        38,859,471      dTLB-loads                #    0.466 M/sec                    ( +-  0.41% )  (54.61%)
         6,347,954      dTLB-load-misses          #   16.34% of all dTLB cache hits   ( +-  0.17% )  (54.62%)
        12,480,335      iTLB-loads                #    0.150 M/sec                    ( +-  1.19% )  (54.58%)
           133,114      iTLB-load-misses          #    1.07% of all iTLB cache hits   ( +-  0.90% )  (54.53%)
   <not supported>      L1-dcache-prefetches                                        
   <not supported>      L1-dcache-prefetch-misses                                   

             5.094 +- 0.639 seconds time elapsed  ( +- 12.54% )

```