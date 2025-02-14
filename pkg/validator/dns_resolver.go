package validator

import (
	"net"
	"time"
)

// DNSResolver interface for making DNS lookups configurable and mockable
type DNSResolver interface {
	LookupHost(domain string) ([]string, error)
	LookupMX(domain string) ([]*net.MX, error)
}

// DefaultResolver implements DNSResolver using net package
type DefaultResolver struct {
	timeout time.Duration
}

func (r *DefaultResolver) LookupHost(domain string) ([]string, error) {
	resultChan := make(chan []string, 1)
	errChan := make(chan error, 1)

	go func() {
		addrs, err := net.LookupHost(domain)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- addrs
	}()

	select {
	case addrs := <-resultChan:
		return addrs, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(r.timeout):
		return nil, net.ErrClosed
	}
}

func (r *DefaultResolver) LookupMX(domain string) ([]*net.MX, error) {
	resultChan := make(chan []*net.MX, 1)
	errChan := make(chan error, 1)

	go func() {
		mxs, err := net.LookupMX(domain)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- mxs
	}()

	select {
	case mxs := <-resultChan:
		return mxs, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(r.timeout):
		return nil, net.ErrClosed
	}
}
