//go:build x11

package x11

import (
	"io"
	"net"
	"testing"
	"time"
)

type testLogger struct {
	t *testing.T
}

func (l *testLogger) Errorf(format string, args ...interface{}) {
	l.t.Errorf(format, args...)
}

func (l *testLogger) Infof(format string, args ...interface{}) {
	l.t.Logf(format, args...)
}

func (l *testLogger) Printf(format string, args ...interface{}) {
	l.t.Logf(format, args...)
}

type testConn struct {
	r io.Reader
	w io.Writer
}

func (c *testConn) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *testConn) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

func (c *testConn) Close() error {
	return nil
}

func (c *testConn) LocalAddr() net.Addr {
	return nil
}

func (c *testConn) RemoteAddr() net.Addr {
	return nil
}

func (c *testConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *testConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *testConn) SetWriteDeadline(t time.Time) error {
	return nil
}
