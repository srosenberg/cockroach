package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

const bufferSize = 16 << 10

type PartitionableConn struct {
	net.Conn

	clientConn net.Conn
	serverConn net.Conn

	mu struct {
		syncutil.Mutex

		err error

		c2sPartitioned bool
		s2cPartitioned bool

		c2sBuffer buf
		s2cBuffer buf

		c2sWaiter *sync.Cond
		s2cWaiter *sync.Cond
	}
}

type buf struct {
	*syncutil.Mutex

	data     []byte
	capacity int
	closed   bool

	closedErr error
	name      string

	readerWait *sync.Cond

	capacityWait *sync.Cond
}

func makeBuf(name string, capacity int, mu *syncutil.Mutex) buf {
	__antithesis_instrumentation__.Notify(645572)
	return buf{
		Mutex:        mu,
		name:         name,
		capacity:     capacity,
		readerWait:   sync.NewCond(mu),
		capacityWait: sync.NewCond(mu),
	}
}

func (b *buf) Write(data []byte) (int, error) {
	__antithesis_instrumentation__.Notify(645573)
	b.Lock()
	defer b.Unlock()
	for b.capacity == len(b.data) && func() bool {
		__antithesis_instrumentation__.Notify(645577)
		return !b.closed == true
	}() == true {
		__antithesis_instrumentation__.Notify(645578)

		b.capacityWait.Wait()
	}
	__antithesis_instrumentation__.Notify(645574)
	if b.closed {
		__antithesis_instrumentation__.Notify(645579)
		return 0, b.closedErr
	} else {
		__antithesis_instrumentation__.Notify(645580)
	}
	__antithesis_instrumentation__.Notify(645575)
	available := b.capacity - len(b.data)
	toCopy := available
	if len(data) < available {
		__antithesis_instrumentation__.Notify(645581)
		toCopy = len(data)
	} else {
		__antithesis_instrumentation__.Notify(645582)
	}
	__antithesis_instrumentation__.Notify(645576)
	b.data = append(b.data, data[:toCopy]...)
	b.wakeReaderLocked()
	return toCopy, nil
}

var errEAgain = errors.New("try read again")

func (b *buf) readLocked(size int) ([]byte, error) {
	__antithesis_instrumentation__.Notify(645583)
	if len(b.data) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(645587)
		return !b.closed == true
	}() == true {
		__antithesis_instrumentation__.Notify(645588)
		b.readerWait.Wait()

		return nil, errEAgain
	} else {
		__antithesis_instrumentation__.Notify(645589)
	}
	__antithesis_instrumentation__.Notify(645584)
	if b.closed && func() bool {
		__antithesis_instrumentation__.Notify(645590)
		return len(b.data) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(645591)
		return nil, b.closedErr
	} else {
		__antithesis_instrumentation__.Notify(645592)
	}
	__antithesis_instrumentation__.Notify(645585)
	var ret []byte
	if len(b.data) < size {
		__antithesis_instrumentation__.Notify(645593)
		ret = b.data
		b.data = nil
	} else {
		__antithesis_instrumentation__.Notify(645594)
		ret = b.data[:size]
		b.data = b.data[size:]
	}
	__antithesis_instrumentation__.Notify(645586)
	b.capacityWait.Signal()
	return ret, nil
}

func (b *buf) Close(err error) {
	__antithesis_instrumentation__.Notify(645595)
	b.Lock()
	defer b.Unlock()
	b.closed = true
	b.closedErr = err
	b.readerWait.Signal()
	b.capacityWait.Signal()
}

func (b *buf) wakeReaderLocked() {
	__antithesis_instrumentation__.Notify(645596)
	b.readerWait.Signal()
}

func NewPartitionableConn(serverConn net.Conn) *PartitionableConn {
	__antithesis_instrumentation__.Notify(645597)
	clientEnd, clientConn := net.Pipe()
	c := &PartitionableConn{
		Conn:       clientEnd,
		clientConn: clientConn,
		serverConn: serverConn,
	}
	c.mu.c2sWaiter = sync.NewCond(&c.mu.Mutex)
	c.mu.s2cWaiter = sync.NewCond(&c.mu.Mutex)
	c.mu.c2sBuffer = makeBuf("c2sBuf", bufferSize, &c.mu.Mutex)
	c.mu.s2cBuffer = makeBuf("s2cBuf", bufferSize, &c.mu.Mutex)

	go func() {
		__antithesis_instrumentation__.Notify(645600)
		err := c.copy(
			c.clientConn,
			c.serverConn,
			&c.mu.c2sBuffer,
			func() {
				__antithesis_instrumentation__.Notify(645603)
				for c.mu.c2sPartitioned {
					__antithesis_instrumentation__.Notify(645604)
					c.mu.c2sWaiter.Wait()
				}
			})
		__antithesis_instrumentation__.Notify(645601)
		c.mu.Lock()
		c.mu.err = err
		c.mu.Unlock()
		if err := c.clientConn.Close(); err != nil {
			__antithesis_instrumentation__.Notify(645605)
			log.Errorf(context.TODO(), "unexpected error closing internal pipe: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(645606)
		}
		__antithesis_instrumentation__.Notify(645602)
		if err := c.serverConn.Close(); err != nil {
			__antithesis_instrumentation__.Notify(645607)
			log.Errorf(context.TODO(), "error closing server conn: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(645608)
		}
	}()
	__antithesis_instrumentation__.Notify(645598)

	go func() {
		__antithesis_instrumentation__.Notify(645609)
		err := c.copy(
			c.serverConn,
			c.clientConn,
			&c.mu.s2cBuffer,
			func() {
				__antithesis_instrumentation__.Notify(645612)
				for c.mu.s2cPartitioned {
					__antithesis_instrumentation__.Notify(645613)
					c.mu.s2cWaiter.Wait()
				}
			})
		__antithesis_instrumentation__.Notify(645610)
		c.mu.Lock()
		c.mu.err = err
		c.mu.Unlock()
		if err := c.clientConn.Close(); err != nil {
			__antithesis_instrumentation__.Notify(645614)
			log.Fatalf(context.TODO(), "unexpected error closing internal pipe: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(645615)
		}
		__antithesis_instrumentation__.Notify(645611)
		if err := c.serverConn.Close(); err != nil {
			__antithesis_instrumentation__.Notify(645616)
			log.Errorf(context.TODO(), "error closing server conn: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(645617)
		}
	}()
	__antithesis_instrumentation__.Notify(645599)

	return c
}

func (c *PartitionableConn) Finish() {
	__antithesis_instrumentation__.Notify(645618)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mu.c2sPartitioned = false
	c.mu.c2sWaiter.Signal()
	c.mu.s2cPartitioned = false
	c.mu.s2cWaiter.Signal()
}

func (c *PartitionableConn) PartitionC2S() {
	__antithesis_instrumentation__.Notify(645619)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mu.c2sPartitioned {
		__antithesis_instrumentation__.Notify(645621)
		panic("already partitioned")
	} else {
		__antithesis_instrumentation__.Notify(645622)
	}
	__antithesis_instrumentation__.Notify(645620)
	c.mu.c2sPartitioned = true
	c.mu.c2sBuffer.wakeReaderLocked()
}

func (c *PartitionableConn) UnpartitionC2S() {
	__antithesis_instrumentation__.Notify(645623)
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.mu.c2sPartitioned {
		__antithesis_instrumentation__.Notify(645625)
		panic("not partitioned")
	} else {
		__antithesis_instrumentation__.Notify(645626)
	}
	__antithesis_instrumentation__.Notify(645624)
	c.mu.c2sPartitioned = false
	c.mu.c2sWaiter.Signal()
}

func (c *PartitionableConn) PartitionS2C() {
	__antithesis_instrumentation__.Notify(645627)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mu.s2cPartitioned {
		__antithesis_instrumentation__.Notify(645629)
		panic("already partitioned")
	} else {
		__antithesis_instrumentation__.Notify(645630)
	}
	__antithesis_instrumentation__.Notify(645628)
	c.mu.s2cPartitioned = true
	c.mu.s2cBuffer.wakeReaderLocked()
}

func (c *PartitionableConn) UnpartitionS2C() {
	__antithesis_instrumentation__.Notify(645631)
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.mu.s2cPartitioned {
		__antithesis_instrumentation__.Notify(645633)
		panic("not partitioned")
	} else {
		__antithesis_instrumentation__.Notify(645634)
	}
	__antithesis_instrumentation__.Notify(645632)
	c.mu.s2cPartitioned = false
	c.mu.s2cWaiter.Signal()
}

func (c *PartitionableConn) Read(b []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(645635)
	c.mu.Lock()
	err = c.mu.err
	c.mu.Unlock()
	if err != nil {
		__antithesis_instrumentation__.Notify(645637)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(645638)
	}
	__antithesis_instrumentation__.Notify(645636)

	return c.Conn.Read(b)
}

func (c *PartitionableConn) Write(b []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(645639)
	c.mu.Lock()
	err = c.mu.err
	c.mu.Unlock()
	if err != nil {
		__antithesis_instrumentation__.Notify(645641)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(645642)
	}
	__antithesis_instrumentation__.Notify(645640)

	return c.Conn.Write(b)
}

func (b *buf) readFrom(src io.Reader) error {
	__antithesis_instrumentation__.Notify(645643)
	data := make([]byte, 1024)
	for {
		__antithesis_instrumentation__.Notify(645644)
		nr, err := src.Read(data)
		if err != nil {
			__antithesis_instrumentation__.Notify(645646)
			return err
		} else {
			__antithesis_instrumentation__.Notify(645647)
		}
		__antithesis_instrumentation__.Notify(645645)
		toSend := data[:nr]
		for {
			__antithesis_instrumentation__.Notify(645648)
			nw, ew := b.Write(toSend)
			if ew != nil {
				__antithesis_instrumentation__.Notify(645651)
				return ew
			} else {
				__antithesis_instrumentation__.Notify(645652)
			}
			__antithesis_instrumentation__.Notify(645649)
			if nw == len(toSend) {
				__antithesis_instrumentation__.Notify(645653)
				break
			} else {
				__antithesis_instrumentation__.Notify(645654)
			}
			__antithesis_instrumentation__.Notify(645650)
			toSend = toSend[nw:]
		}
	}
}

func (c *PartitionableConn) copyFromBuffer(
	src *buf, dst net.Conn, waitForNoPartitionLocked func(),
) error {
	__antithesis_instrumentation__.Notify(645655)
	for {
		__antithesis_instrumentation__.Notify(645656)

		src.Mutex.Lock()
		waitForNoPartitionLocked()
		data, err := src.readLocked(1024 * 1024)
		src.Mutex.Unlock()

		if len(data) > 0 {
			__antithesis_instrumentation__.Notify(645658)
			nw, ew := dst.Write(data)
			if ew != nil {
				__antithesis_instrumentation__.Notify(645660)
				err = ew
			} else {
				__antithesis_instrumentation__.Notify(645661)
			}
			__antithesis_instrumentation__.Notify(645659)
			if len(data) != nw {
				__antithesis_instrumentation__.Notify(645662)
				err = io.ErrShortWrite
			} else {
				__antithesis_instrumentation__.Notify(645663)
			}
		} else {
			__antithesis_instrumentation__.Notify(645664)
			if err == nil {
				__antithesis_instrumentation__.Notify(645665)
				err = io.EOF
			} else {
				__antithesis_instrumentation__.Notify(645666)
				if errors.Is(err, errEAgain) {
					__antithesis_instrumentation__.Notify(645667)
					continue
				} else {
					__antithesis_instrumentation__.Notify(645668)
				}
			}
		}
		__antithesis_instrumentation__.Notify(645657)
		if err != nil {
			__antithesis_instrumentation__.Notify(645669)
			return err
		} else {
			__antithesis_instrumentation__.Notify(645670)
		}
	}
}

func (c *PartitionableConn) copy(
	src net.Conn, dst net.Conn, buf *buf, waitForNoPartitionLocked func(),
) error {
	__antithesis_instrumentation__.Notify(645671)
	tasks := make(chan error)
	go func() {
		__antithesis_instrumentation__.Notify(645675)
		err := buf.readFrom(src)
		buf.Close(err)
		tasks <- err
	}()
	__antithesis_instrumentation__.Notify(645672)
	go func() {
		__antithesis_instrumentation__.Notify(645676)
		err := c.copyFromBuffer(buf, dst, waitForNoPartitionLocked)
		buf.Close(err)
		tasks <- err
	}()
	__antithesis_instrumentation__.Notify(645673)
	err := <-tasks
	err2 := <-tasks
	if err == nil {
		__antithesis_instrumentation__.Notify(645677)
		err = err2
	} else {
		__antithesis_instrumentation__.Notify(645678)
	}
	__antithesis_instrumentation__.Notify(645674)
	return err
}
