package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"

	"github.com/Jigsaw-Code/outline-sdk/transport"
	"github.com/Jigsaw-Code/outline-sdk/x/config"
	"github.com/Jigsaw-Code/outline-sdk/x/httpproxy"
)

type runningProxy struct {
	server  *http.Server
	Address string
}

func (p *runningProxy) Close() {
	p.server.Close()
}

// newFilteredStreamDialer creates a direct [transport.StreamDialer] that blocks
// non public IPs to prevent access to localhost or the local network.
func newFilteredStreamDialer() transport.StreamDialer {
	var dialer net.Dialer
	dialer.Control = func(network, address string, c syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("failed to parse address: %w", err)
		}
		if ip := net.ParseIP(host); ip != nil {
			if !ip.IsGlobalUnicast() {
				return fmt.Errorf("addresses that are not global unicast are fobidden")
			}
			if ip.IsPrivate() {
				return fmt.Errorf("private addresses are forbidden")
			}
		}
		return nil
	}
	return &transport.TCPDialer{Dialer: dialer}
}

func runServer(address, transport string) (*runningProxy, error) {
	// TODO: block localhost, maybe local net.
	dialer, err := config.WrapStreamDialer(newFilteredStreamDialer(), transport)
	if err != nil {
		return nil, fmt.Errorf("could not create dialer: %w", err)
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("could not listen on address %v: %w", address, err)
	}

	server := http.Server{Handler: httpproxy.NewProxyHandler(dialer)}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Serve failed: %v\n", err)
		}
	}()
	return &runningProxy{server: &server, Address: listener.Addr().String()}, nil
}
