package server

import (
	"IMServer/internal/server/conf"
	"IMServer/internal/server/model"
)

var (
	tcpS *TCPServer
	udpS *UDPServer
)

var done = make(chan struct{}, 1)

func Run() {
	conf.L.Info("server run")

	go model.RunTicker(done)

	go runHTTP(conf.Conf.App.HttpAddr)

	udpS = NewUDPServer(conf.Conf.Server.UdpAddr)
	go udpS.ListenAndServer()

	tcpS = NewTCPServer(conf.Conf.Server.TcpAddr)
	tcpS.ListenAndServer()
}

func Close() {
	conf.L.Info("server close")
	tcpS.Stop()
	udpS.Stop()

	close(done)
}
