package server

import (
	"IMServer/internal/server/conf"
	"fmt"
	"net"
	"reflect"
	"time"

	"go.uber.org/zap"

	"gogit.oa.com/March/gopkg/metric"

	"gogit.oa.com/March/gopkg/util"

	"gogit.oa.com/March/gopkg/protocol/bypack"
)

type UDPPackage struct {
	addr   net.Addr
	reader *bypack.Reader
}

type UDPServer struct {
	uc   *net.UDPConn
	pc   chan UDPPackage
	done chan struct{}
}

func (s *UDPServer) ListenAndServer() {
	conf.L.Info("udp listen and server")

	go s.Proc()

	var buff = make([]byte, bypack.MaxBufferLen)
	for {
		n, addr, err := s.uc.ReadFrom(buff)
		if err != nil {
			break
		}
		if n < int(bypack.HeaderSize) {
			continue
		}

		h, err := bypack.NewHeader(buff[:bypack.HeaderSize])
		if err != nil {
			conf.L.Error(err.Error())
			continue
		}

		r := bypack.NewReader(h.GetCmd(), buff[bypack.HeaderSize:int(bypack.HeaderSize)+int(h.GetSize())])
		r.RawBuffer = buff[:int(bypack.HeaderSize)+int(h.GetSize())]

		select {
		case s.pc <- UDPPackage{
			addr:   addr,
			reader: r,
		}:
		case <-time.After(30 * time.Second):
			conf.L.Warn("udp channel full!!!")
		}
	}
}

func (s *UDPServer) Proc() {
	for {
		select {
		case <-s.done:
			return
		case p := <-s.pc:
			go s.Transport(p)
		}
	}
}

func (s *UDPServer) Stop() {
	conf.L.Info("udp server stop")
	_ = s.uc.Close()
	close(s.done)
}

func (s *UDPServer) Transport(p UDPPackage) {
	defer func() {
		if err := recover(); err != nil {
			conf.L.Error(util.CatchPanic(err).Error())
		}
	}()

	worker := NewWorker(p.addr, p.reader)
	method := fmt.Sprintf("UDP0x%x", p.reader.GetCmd())
	v := reflect.ValueOf(worker).MethodByName(method)
	if v.String() == "<invalid Value>" {
		conf.L.Sugar().Warnf("worker not found method %s", method)
		return
	}

	conf.L.Debug(method, zap.String("conn", p.addr.String()))

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

func NewUDPServer(addr string) *UDPServer {
	a, err := net.ResolveUDPAddr("udp", addr)
	util.MustNil(err)
	uc, err := net.ListenUDP("udp", a)
	util.MustNil(err)

	return &UDPServer{
		uc:   uc,
		pc:   make(chan UDPPackage, 1024),
		done: make(chan struct{}, 1),
	}
}
