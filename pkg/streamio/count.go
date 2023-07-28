package streamio

import "io"

// CountBytes returns an io.Reader that counts all bytes read and updates the counter appropriately.
// counter must not be nil.
func CountBytes(r io.Reader, counter *int64) io.Reader {
	return countReader{r, counter}
}

// countReader implements the reader for CountBytes.
type countReader struct {
	io.Reader
	counter *int64
}

// Read implements io.Reader.
// Reading operations are passed to the underlying reader.
func (c countReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	*c.counter += int64(n)
	return n, err
}
