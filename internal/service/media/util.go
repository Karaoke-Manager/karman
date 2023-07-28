package media

import "io"

type countReader struct {
	r io.Reader
	c *int64
}

func countBytes(r io.Reader, count *int64) io.Reader {
	return countReader{r, count}
}

func (c countReader) Read(p []byte) (n int, err error) {
	n, err = c.r.Read(p)
	*c.c += int64(n)
	return
}
