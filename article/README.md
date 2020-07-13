# How worldiety scaled to 450 million rows a second
*by Torben Schinke, 2020, July 10*

There are many highly scalable big databases out there, like Scylla which
is capable of scanning more than [1 billion rows](https://www.scylladb.com/2019/12/12/how-scylla-scaled-to-one-billion-rows-a-second/)
a second on a cluster consisting of 83 nodes. This allows them to read
12 millions rows per second per server.

Not bad, but this still means, that a dataset of 5 billion rows needs 7 minutes to
process. We thought, we can do better:

