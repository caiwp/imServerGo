package server

import (
	"IMServer/internal/server/conf"
	"IMServer/internal/server/model"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"

	"go.uber.org/zap"

	"gogit.oa.com/March/gopkg/metric"

	"gogit.oa.com/March/gopkg/util"

	"gogit.oa.com/March/gopkg/protocol/bypack"
)

type TCPPackage struct {
	conn   net.Conn
	reader *bypack.Reader
}

type TCPServer struct {
	lis  net.Listener
	pc   chan TCPPackage
	done chan struct{}
}

func (s *TCPServer) ListenAndServer() {
	conf.L.Info("tcp listen and server")
	go s.Proc()
	go model.Heartbeat(s.done)

	for {
		conn, err := s.lis.Accept()
		if err != nil {
			conf.L.Warn(err.Error())
			return
		}

		model.AddConn(conn)
		go func() {
			s.handleConn(conn)
			model.DelConn(conn)
		}()
	}
}

func (s *TCPServer) handleConn(conn net.Conn) {
	conf.L.Info("handle conn", zap.String("conn", conn.RemoteAddr().String()))
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(60 * time.Second)); err != nil {
		conf.L.Warn(err.Error())
		return
	}

	for {
		var hb = make([]byte, bypack.HeaderSize)
		n, err := io.ReadFull(conn, hb)
		if err != nil {
			return
		}

		header, err := bypack.NewHeader(hb[:n])
		if err != nil {
			conf.L.Warn(err.Error())
			return
		}

		buff := make([]byte, header.GetSize())
		_, err = io.ReadFull(conn, buff)
		if err != nil {
			conf.L.Error(err.Error())
			return
		}

		reader := bypack.NewReader(header.GetCmd(), buff)
		reader.RawBuffer = append(hb, buff...)

		select {
		case s.pc <- TCPPackage{
			conn:   conn,
			reader: reader,
		}:
			if err = conn.SetDeadline(time.Now().Add(60 * time.Second)); err != nil {
				conf.L.Warn(err.Error())
				return
			}

		case <-time.After(30 * time.Second):
			conf.L.Sugar().Errorf("conn %s channel full!!!", conn.RemoteAddr().String())
		}
	}
}

func (s *TCPServer) Proc() {
	for {
		select {
		case p := <-s.pc:
			go s.Transport(p)

		case <-s.done:
			conf.L.Info("tcp server done")
			return
		}
	}
}

func (s *TCPServer) Stop() {
	conf.L.Info("tcp server stop")
	_ = s.lis.Close()
	close(s.done)
}

func (s *TCPServer) Transport(p TCPPackage) {
	defer func() {
		if err := recover(); err != nil {
			conf.L.Error(util.CatchPanic(err).Error())
		}
	}()

	model.AddConn(p.conn)

	worker := NewWorkerWithConn(p.conn, p.reader)
	method := fmt.Sprintf("TCP0x%x", p.reader.GetCmd())
	v := reflect.ValueOf(worker).MethodByName(method)
	if v.String() == "<invalid Value>" {
		conf.L.Sugar().Warnf("worker not found method %s", method)
		return
	}

	if p.reader.GetCmd() != 0x2 {
		conf.L.Debug(method, zap.String("conn", p.conn.RemoteAddr().String()))
	}

	reporter := metric.NewReporter(method)
	res := v.Call(nil)
	if len(res) > 0 {
		code, ok := res[0].Interface().(metric.Code)
		if !ok {
			return
		}
		reporter.HandledWithCode(code)
	}
}

func NewTCPServer(addr string) *TCPServer {
	lis, err := net.Listen("tcp", addr)
	util.MustNil(err)

	return &TCPServer{
		lis:  lis,
		pc:   make(chan TCPPackage, 1024),
		done: make(chan struct{}, 1),
	}
}
