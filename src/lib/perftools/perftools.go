package perftools

import (
	"net"
	"time"

	"extremeWorkload.com/daytrader/lib"
	audit "extremeWorkload.com/daytrader/lib/audit"
)

// PerfConn wraps net.Conn interface with a timing logging
type PerfConn struct {
	innerConn   net.Conn
	auditClient *audit.AuditClient
	AcceptTime  uint64
	ReadTime    uint64
	WriteTime   uint64
	CloseTime   uint64
}

// NewPerfConn create PerfConn fron net.Conn
func NewPerfConn(innerConn net.Conn) *PerfConn {
	return &PerfConn{
		innerConn:  innerConn,
		AcceptTime: lib.GetUnixTimestamp(),
	}
}

// SetAuditClient sets client to use later for logging
func (pconn *PerfConn) SetAuditClient(auditClient *audit.AuditClient) {
	pconn.auditClient = auditClient
}

func (pconn *PerfConn) Read(b []byte) (n int, err error) {
	pconn.ReadTime = lib.GetUnixTimestamp() - pconn.AcceptTime
	return pconn.innerConn.Read(b)
}

func (pconn *PerfConn) Write(b []byte) (n int, err error) {
	pconn.WriteTime = lib.GetUnixTimestamp() - pconn.AcceptTime
	return pconn.innerConn.Write(b)
}

// Close adds timing
func (pconn *PerfConn) Close() error {
	pconn.CloseTime = lib.GetUnixTimestamp() - pconn.AcceptTime
	if pconn.auditClient != nil && lib.PerfLoggingEnabled {
		var performanceInfo = audit.PerformanceMetricInfo{
			AcceptTimestamp: pconn.AcceptTime,
			ReadTimestamp:   pconn.ReadTime,
			WriteTimestamp:  pconn.WriteTime,
			CloseTimestamp:  pconn.CloseTime,
		}
		pconn.auditClient.LogPerformanceMetric(performanceInfo)
	}
	return pconn.innerConn.Close()
}

// LocalAddr same as net.LocalAddr
func (pconn *PerfConn) LocalAddr() net.Addr {
	return pconn.innerConn.LocalAddr()
}

// RemoteAddr same as net.LocalAddr
func (pconn *PerfConn) RemoteAddr() net.Addr {
	return pconn.innerConn.RemoteAddr()
}

// SetDeadline same as net.LocalAddr
func (pconn *PerfConn) SetDeadline(t time.Time) error {
	return pconn.innerConn.SetDeadline(t)
}

// SetReadDeadline same as net.LocalAddr
func (pconn *PerfConn) SetReadDeadline(t time.Time) error {
	return pconn.innerConn.SetReadDeadline(t)
}

// SetWriteDeadline same as net.LocalAddr
func (pconn *PerfConn) SetWriteDeadline(t time.Time) error {
	return pconn.innerConn.SetWriteDeadline(t)
}
