package client

import (
	"IMServer/internal/client/conf"

	"gogit.oa.com/March/gopkg/protocol/bypack"
)

func buffer888() []byte {
	w := bypack.NewWriter(0x888)
	w.String(conf.Conf.Keys.Tcp0x888)
	w.End()
	return w.GetBuffer()
}
