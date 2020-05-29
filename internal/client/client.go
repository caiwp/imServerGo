package client

import (
	"IMServer/internal/client/conf"
	"io"
	"net"

	"gogit.oa.com/March/gopkg/protocol/bypack"

	"gogit.oa.com/March/gopkg/util"
)

func Run() {
	reader := SendAndRecv(buffer888())
	conf.L.Info(reader.String())
}

func Send(data []byte) {
	_, err := Conn().Write(data)
	util.MustNil(err)
}

func SendAndRecv(data []byte) *bypack.Reader {
	conn := Conn()
	_, err := conn.Write(data)
	util.MustNil(err)

	hb := make([]byte, bypack.HeaderSize)
	_, err = io.ReadFull(conn, hb)
	util.MustNil(err)

	header, err := bypack.NewHeader(hb)
	util.MustNil(err)

	body := make([]byte, header.GetSize())
	n, err := io.ReadFull(conn, body)
	util.MustNil(err)

	return bypack.NewReader(header.GetCmd(), body[:n])
}

func Conn() net.Conn {
	conn, err := net.Dial("tcp", conf.Conf.Client.Addr)
	util.MustNil(err)
	return conn
}
