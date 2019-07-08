// Package socket implements a simple server that copies
// line buffered content from a socket to stdout.
package socket

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

// OptionsFunc is a function passed to new for setting options on a new server.
type OptionsFunc func(*Server) error

// Server copies data from a socket to stdout
type Server struct {
	tcpAddr  string
	unixAddr string
	auxAddr  string
	stopChan chan struct{}
	listener net.Listener
}

// New creates a new server
func New(options ...OptionsFunc) (*Server, error) {
	s := &Server{
		auxAddr:  ":9090",
		stopChan: make(chan struct{}, 1),
	}

	for _, f := range options {
		if err := f(s); err != nil {
			return nil, errors.Wrap(err, "options func failed")
		}
	}

	return s, nil
}

// SetTCPAddress will set the TCP address to accept line oriented data.
// Only TCP or Unix address may be set. If both are set, Unix takes precedence.
func SetTCPAddress(addr string) func(*Server) error {
	return func(s *Server) error {
		s.tcpAddr = addr
		return nil
	}
}

// SetUnixAddress will set the Unix address to accept line oriented data.
// Only TCP or Unix address may be set. If both are set, Unix takes precedence.
func SetUnixAddress(addr string) func(*Server) error {
	return func(s *Server) error {
		s.unixAddr = addr
		return nil
	}
}

// SetAuxAddress will set the address for the metrics listener.
func SetAuxAddress(addr string) func(*Server) error {
	return func(s *Server) error {
		s.auxAddr = addr
		return nil
	}
}

// Run will start the listener. This will run until
// a SIGINT or SIGTERM are received.
func (s *Server) Run() error {
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	var srv net.Listener
	var err error

	if s.unixAddr != "" {
		srv, err = unixListener(s.unixAddr)
		if err != nil {
			return errors.Wrap(err, "unix listen failed")
		}
	} else {
		srv, err = tcpListener(s.tcpAddr)
		if err != nil {
			return errors.Wrap(err, "tcp listen failed")
		}
	}

	s.listener = srv

	http.HandleFunc("/healthz", s.healthz)
	http.Handle("/metrics", promhttp.Handler())

	// channel for shutdown
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	var g errgroup.Group

	// default mux handles aux
	auxSrv := &http.Server{
		Addr: s.auxAddr,
	}

	g.Go(func() error {
		return auxSrv.ListenAndServe()
	})

	g.Go(func() error {
		return s.copyData()
	})

	g.Go(func() error {
		// wait for shutdown signal. stop aux listener first as health checks will fail.
		<-stopChan
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		auxSrv.Shutdown(ctx)
		s.Stop()
		return nil
	})

	if err := g.Wait(); err != http.ErrServerClosed {
		return errors.Wrap(err, "failed to run server")
	}

	return nil
}

// Stop will stop the server
func (s *Server) Stop() {
	s.stopChan <- struct{}{}
	_ = s.listener.Close()
}

var healthzOK = []byte("ok\n")

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.Write(healthzOK)
}

func tcpListener(addr string) (net.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "tcp listen failed")
	}
	return l, nil
}

func unixListener(addr string) (net.Listener, error) {
	a := net.UnixAddr{
		Name: addr,
		Net:  "unix",
	}

	l, err := net.ListenUnix("unix", &a)

	if err != nil {
		return nil, errors.Wrap(err, "unix listen failed")
	}

	l.SetUnlinkOnClose(true)

	// TODO: allow overriding perms?
	if err := os.Chmod(addr, 0777); err != nil {
		return nil, errors.Wrap(err, "failed to set pemissions on unix socket file")
	}

	return l, nil
}

func errorLog(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func (s *Server) copyData() error {
	// XXX: do we ever want to increase channel size?

	buffer := make(chan []byte, 1)

	go stdoutWriter(buffer)

	for {
		select {
		case <-s.stopChan:
			return nil
		default:
		}
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return nil
			default:
				errorLog("accept failed: %s\n", err)
			}
			continue
		}
		connectionCounter.Inc()
		go handleConnection(conn, buffer)
	}
}

// probably need a stop on this, but it exists on signal
func stdoutWriter(buffer chan []byte) {
	for {
		_, err := os.Stdout.Write(<-buffer)
		if err != nil {
			errorLog("failed to write to stdout: %s\n", err)
		}
		linesWritten.Inc()
	}
}

func handleConnection(conn net.Conn, buffer chan []byte) {
	connectionGuage.Inc()
	defer connectionGuage.Dec()

	r := bufio.NewReader(conn)
	for {
		data, err := r.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				errorLog("read failed: %s", err)
			}
			_ = conn.Close()
			return
		}
		linesRead.Inc()
		buffer <- data
	}
}
