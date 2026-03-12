package execer

import "os"

func HomeDir() (string, error) {
	return os.UserHomeDir()
}

// tailBuffer keeps the last `max` bytes written to it.
// Used to capture the tail of stderr for error reporting.
type tailBuffer struct {
	buf []byte
	max int
}

func newTailBuffer(max int) *tailBuffer {
	return &tailBuffer{max: max}
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	if len(b.buf) > b.max {
		b.buf = b.buf[len(b.buf)-b.max:]
	}
	return len(p), nil
}

func (b *tailBuffer) String() string {
	return string(b.buf)
}
