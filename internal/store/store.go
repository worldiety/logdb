package store

import "os"

// Store handles a flat file to store no-sql column data. It reserves maxHeaderSize at the beginning, so
// that later changes by appending are fine - up to the imposed limits.
type Store struct {
	header Header
	file   os.File
}

// OpenStore creates or updates an existing file. If the file looks corrupted
func OpenStore(fname string)(*Store,error){
	os.Open()
}
