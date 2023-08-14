package streamio

import (
	"errors"
	"fmt"
	"io"
)

var (
	// ErrSeekOutOfRange indicates that a seek operation could not be fulfilled because the buffer was not big enough.
	ErrSeekOutOfRange = errors.New("out of range")
)

// bufReader implements a buffered reader supporting partial seek operations.
type bufReader struct {
	r          io.Reader // underlying reader
	buf        []byte    // seek buffer
	offset     int64     // number of bytes read from r
	seekOffset int       // number of bytes >= 0 we have seeked backwards into buf
}

// NewBufferedReadSeeker creates a new io.ReadSeeker that reads from r.
// Seek operations are implemented by buffering read operations from r.
// Seeking is only supported up to size bytes back.
// Advancing the reader will first return up to size bytes from the internal buffer
// before the next read operation is passed on to r.
func NewBufferedReadSeeker(r io.Reader, size int) io.ReadSeeker {
	if rs, ok := r.(io.ReadSeeker); ok {
		return rs
	}
	return &bufReader{
		r:   r,
		buf: make([]byte, size),
	}
}

// Offset returns the current offset of r from the beginning of the file.
// This may be less than the number of bytes read from the underlying reader if a Seek call has moved the reader backwards.
func (r *bufReader) Offset() int64 {
	return r.offset - int64(r.seekOffset)
}

// Read implements io.Reader.
// Read operations are passed on to the underlying reader after the current seek buffer has been fully consumed.
func (r *bufReader) Read(p []byte) (n int, err error) {
	if r.seekOffset > 0 {
		n = copy(p, r.buf[len(r.buf)-r.seekOffset:])
		r.seekOffset -= n
		return
	}
	n, err = r.r.Read(p)
	r.offset += int64(n)
	if n >= len(r.buf) {
		copy(r.buf, p[n-len(r.buf):n])
	} else {
		copy(r.buf, r.buf[n:])
		copy(r.buf[len(r.buf)-n:], p)
	}
	return
}

// Seek implements io.Seeker.
// Seek operations are only supported within the buffer size of r.
func (r *bufReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return r.Offset(), fmt.Errorf("streamio: Seek(%d, %d): seek before start: %w", offset, whence, ErrSeekOutOfRange)
		}
		if offset < r.offset-int64(len(r.buf)) {
			return r.Offset(), fmt.Errorf("streamio: Seek(%d, %d): buffer exceeded (%d bytes read): %w", offset, whence, r.offset, ErrSeekOutOfRange)
		}
		if offset <= r.offset {
			r.seekOffset = int(r.offset - offset)
			return r.Offset(), nil
		}
		r.seekOffset = 0
		_, err := io.CopyN(io.Discard, r, offset-r.offset)
		return r.offset, err
	case io.SeekCurrent:
		if offset >= 0 {
			_, err := io.CopyN(io.Discard, r, offset)
			return r.Offset(), err
		}
		if r.seekOffset-int(offset) > len(r.buf) {
			return r.Offset(), fmt.Errorf("streamio: Seek(%d, %d): offset %d exceeds buffer: %w", offset, whence, r.seekOffset+int(offset), ErrSeekOutOfRange)
		}
		r.seekOffset -= int(offset)
		return r.Offset(), nil
	case io.SeekEnd:
		if offset > 0 {
			offset = 0
		}
		if -int(offset) > len(r.buf) {
			return r.Offset(), fmt.Errorf("streamio: Seek(%d, %d): buffer exceeded: %w", offset, whence, ErrSeekOutOfRange)
		}
		_, err := io.Copy(io.Discard, r)
		r.seekOffset = -int(offset)
		return r.Offset(), err
	default:
		panic(fmt.Sprintf("invalid whence: %d", whence))
	}
}
