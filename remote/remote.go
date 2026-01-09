package remote

import (
	"net"
	"strconv"
)

// 最大包长度
const maxLenght = 1<<16 - 2

const Port uint16 = 9999

type conn struct {
	con         net.Conn
	readCh      chan []byte
	writeBuf    []byte
	closeSignal chan struct{}
	onError     func(error)
	blackHold   bool
}

func NewConn(con net.Conn) *conn {
	conn := &conn{
		con:         con,
		readCh:      make(chan []byte, 1),
		closeSignal: make(chan struct{}),
	}
	go conn.listen()
	return conn
}

func (c *conn) listen() {
	var readBuf [1024]byte
	totalLen := 0
	var buf []byte
	for {
		select {
		case <-c.closeSignal:
			return
		default:
		}
		n, err := c.con.Read(readBuf[:])
		if err != nil {
			select {
			case <-c.closeSignal:
				return
			default:
				if c.onError != nil {
					c.onError(err)
				} else {
					panic(err)
					//errHandler(err)
				}
			}
		}
		buf = append(buf, readBuf[:n]...)
		for {
			if len(buf) < 2 {
				break
			}
			if totalLen == 0 {
				totalLen = int(buf[0]) + int(buf[1])<<8
			}
			if totalLen <= len(buf) {
				c.readCh <- buf[2:totalLen]
				tmpBuf := make([]byte, 0)
				tmpBuf = append(tmpBuf, buf[totalLen:]...)
				buf = tmpBuf
				totalLen = 0
			} else {
				break
			}
		}

	}
}

func (c *conn) read() []byte {
	return <-c.readCh
}

func (c *conn) Write(p []byte) (n int, err error) {
	if c.writeBuf == nil {
		c.writeBuf = make([]byte, 2)
	}
	c.writeBuf = append(c.writeBuf, p...)
	n = len(p)
	return
}

func (c *conn) send() {
	lenght := len(c.writeBuf)
	if lenght-2 > maxLenght {
		panic("过长的数据包")
	}
	c.writeBuf[0] = byte(lenght)
	c.writeBuf[1] = byte(lenght >> 8)
	if c.blackHold {
		c.writeBuf = nil
		return
	}
	_, err := c.con.Write(c.writeBuf)
	if err != nil {
		panic(err)
	}
	//errHandler(err)
	c.writeBuf = nil
}

func (c *conn) Close() {
	select {
	case <-c.closeSignal:
		return
	default:
	}
	close(c.closeSignal)
	c.con.Close()
}

func (c *conn) Switch2BlackHold() {
	c.blackHold = true
}

func (c *conn) SetOnErr(fn func(error)) {
	c.onError = fn
}

type Hub struct {
	cons        chan net.Conn
	closeSignal chan struct{}
	lr          net.Listener
}

func NewHub(maxCons int) *Hub {
	h := &Hub{
		cons:        make(chan net.Conn, maxCons),
		closeSignal: make(chan struct{}),
	}
	go func() {
		lr, err := net.Listen("tcp", ":"+strconv.Itoa(int(Port)))
		h.lr = lr
		errHandler(err)
		for range maxCons {
			c, err := lr.Accept()
			if err != nil {
				select {
				case <-h.closeSignal:
					return
				default:
					panic(err)
				}
			}
			h.cons <- c
		}
		lr.Close()
		close(h.cons)
	}()
	return h
}

func (h *Hub) GetConn() net.Conn {
	return <-h.cons
}

func (h *Hub) GetConnChan() <-chan net.Conn {
	return h.cons
}

func (h *Hub) Close() {
	close(h.closeSignal)
	h.lr.Close()
}

func DialServer(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr+":"+strconv.Itoa(int(Port)))
}

// func (c *conn) write(inf []byte) {
// 	lenght := len(inf)
// 	if lenght > maxLenght {
// 		panic("过长的数据包")
// 	}
// 	lenght += 2
// 	header := make([]byte, 2)
// 	header[0] = byte(lenght)
// 	header[1] = byte(lenght >> 8)
// 	_, err := c.con.Write(append(header, inf...))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
