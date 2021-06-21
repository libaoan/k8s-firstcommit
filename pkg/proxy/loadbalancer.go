package proxy

import (
	"net"
)

type LoadBalancer interface {
	// LoadBalance takes an incoming request and figures out where to route it to.
	// Determination is based on destination service (for example, 'mysql') as
	// well as the source making the connection.
	LoadBalance(service string, srcAddr net.Addr) (string, error)
}
