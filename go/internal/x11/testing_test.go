//go:build x11

package x11

import (
	"io"
	"testing"
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

// MockX11Client is a mock implementation of x11Client for testing.
type MockX11Client struct {
	conn     io.ReadWriteCloser
	sequence uint16
}

func (m *MockX11Client) Read(p []byte) (n int, err error) {
	return m.conn.Read(p)
}

func (m *MockX11Client) Write(p []byte) (n int, err error) {
	return m.conn.Write(p)
}

func (m *MockX11Client) Close() error {
	return m.conn.Close()
}
