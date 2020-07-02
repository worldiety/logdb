# dragster / dragsterdb
*dragster* is a database to turn some *big data* problems for servers into 
*small data* problems on your laptop. It is not a database, which you want to
use for your daily shopping. However, it is a database which
you can use in special edge cases, in which general solutions perform
badly.

## technique
The name *dragster* is not without reason: it is optimized to read or write
rows sequentially at maximum speed. Everything else is either inefficient
or not even possible. We did our best to keep the hot paths for writing
and reading without any dynamic memory allocations. We have chosen
a very primitive design, so that this is even possible with a GC based language
like Go. We only use unsafe to optimize a few string-based scenarios.
In our use cases we are mostly limited by the bus speed of the host.

## history
We had to answer a research question for a dataset of wind power data.
It consisted of approximately 350 millions rows with more than 700 columns
in a normal MySQL database on a huge server. To read and process the data from
server to server using conventional sql techniques (java + hibernate),
the extrapolation of the required processing time, was already measured in weeks - 
for a single run. 

So we decided to find out the technical limits and took a dump of the database
onto an USB-C driven flashdrive. The first disappointing experiments
showed us, that we were not able to consume rows on *localhost* faster, than mysql
returns them. After some profiling, we found out that the amount of GCs in
the sql-abstraction layer within the Go-SDK was the reason. Sadly, there is
no real solution for this, but we found the optimized [pubnative mysql driver](https://github.com/pubnative/mysqldriver-go)
which allowed us to saturate the mysql process at 100%, consuming more than 40.000 
rows per second (without actual processing them and there is still some
GC in the driver itself). From here, we started to create a custom storage
format and engine, for which we researched e.g. how MongoDB handles its schemaless
storage in [BSON](https://en.wikipedia.org/wiki/BSON) and combined it with ideas 
from [Cap'n Proto](https://capnproto.org/). 

At the end, we were able to read the entire dataset from MySQL, convert
and write it into *dragster* with a constant speed of more than 20.000 rows
per second (with more than 700 columns, just mind you). So, we are done
in 3,5 hours, not days or weeks! There is still GC in, so we cannot saturate 
the local MySQL server here, but it is already *acceptably* fast. 
Processing is even faster and limited by the USB-C bus speed, which peaks
out at around 700-800 MB per second. We were able to create an in-memory
index by visiting each object to read every single timestamp - for 350 million 
objects in less than 3 minutes!

## format and restrictions
*Dragster* never stores field names together with entries, instead it
always works with indices of 2 bytes, so you can have at most 2^16 = 65.536 
distinct field names. It is also not possible to nest them, so you have to
map them yourself. However, there is a reserved 16MiB section in the header, 
to store an index table, for your convenience. A single row or *object*, how
we call it, has a theoretical limit of 16MiB, but the current implementation
limits that to 64K, including some meta data. Each field data is prefixed at
least with a type byte and depending on that with some length bytes. This
allows us to jump through all fields, without doing much parsing work. Our
data set also contained a lot of small numbers, which came all in as float64,
so we introduced an automatic detection of the smallest available data type
to represent the number in a (mostly) lossless format. E.g. a float64, which
has a delta of less than 10^-9 to the next natural number is stored as an 
integer. For the integer itself, we check how many bytes are needed (8/16/24/32/40/48/56 or 64)
and use the smallest possible one. We also have a similar check
to use a float32 instead of a float64. This way, we have a kind of a *semantic compression*
which is even magnitudes faster than compression algorithms like LZ4, even though
our prefix-style is so primitive and still wasteful.

## downsides of the design
It is not intended to be able to delete any entries. It is also not really possible
to update entries, though one may at least update any data field, as long as
the required byte width will not change. Inserting and reading concurrently
may be possible but will probably performs bad, especially for random access
in the current implementation, which is optimized for sequential single user/thread 
processing.