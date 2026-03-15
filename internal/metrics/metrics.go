package metrics

import "sync/atomic"

var (
	LinksCreated  atomic.Int64
	ClicksTotal   atomic.Int64
	WSConnections atomic.Int64
)
