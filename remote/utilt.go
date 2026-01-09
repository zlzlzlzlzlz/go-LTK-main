package remote

import (
	"io"
)

func errHandler(err error) {
	if err != nil {
		panic(err)
	}
}

type readBuf struct {
	prt   []byte
	index int
}

func (b *readBuf) Read(p []byte) (n int, err error) {
	if b.index == len(b.prt) {
		err = io.EOF
		return
	}
	copy(p, b.prt[b.index:])
	n = min(len(p), len(b.prt)-b.index)
	b.index += n
	return
}

func (b *readBuf) ReadByte() (out byte, err error) {
	if b.index == len(b.prt) {
		err = io.EOF
		return
	}
	out = b.prt[b.index]
	b.index++
	return
}

func (b *readBuf) load(p []byte) {
	b.prt = p
	b.index = 0
}

type writeBuf struct {
	prt []byte
}

func (b *writeBuf) Write(p []byte) (n int, err error) {
	b.prt = append(b.prt, p...)
	n = len(p)
	return
}

func (b *writeBuf) get() (p []byte) {
	p = b.prt
	b.prt = nil
	return
}
