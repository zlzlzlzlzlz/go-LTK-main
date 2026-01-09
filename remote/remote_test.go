package remote

import (
	"math/rand"
	"net"
	"strconv"
	"testing"
)

func TestCoon(t *testing.T) {
	const testCount = 30000
	data := [][]byte{}
	for i := 0; i < testCount; i++ {
		data = append(data, genRandBytes())
	}
	lr, err := net.Listen("tcp", ":"+strconv.Itoa(int(Port)))
	errHandler(err)
	go func() {
		conn, err := lr.Accept()
		con := NewConn(conn)
		errHandler(err)
		for i := 0; i < testCount; i++ {
			con.Write(data[i])
			con.send()
		}
	}()
	conn, err := DialServer("")
	errHandler(err)
	con := NewConn(conn)
	for i := 0; i < testCount; i++ {
		inf := con.read()
		if !cmpBytes(inf, data[i]) {
			t.Error("err")
		}
	}
}

func genRandBytes() []byte {
	lenght := rand.Intn(200)
	out := make([]byte, lenght)
	for i := 0; i < lenght; i++ {
		out[i] = byte(rand.Intn(256))
	}
	return out
}

func cmpBytes(b1 []byte, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}
