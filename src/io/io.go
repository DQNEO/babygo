package io

type Writer interface {
	Write(p []byte) (n int, err error)
}

type WriteCloser interface {
	Write(p []byte) (n int, err error)
	Close() error
}
