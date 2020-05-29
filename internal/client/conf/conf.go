package conf

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
	"gogit.oa.com/March/gopkg/logger"
	"gogit.oa.com/March/gopkg/util"
)

var L *zap.Logger

var Conf = struct {
	App struct {
		Env string
	}
	Log struct {
		Tag   string
		Addr  string
		Level int
	}
	Client struct {
		Addr string
	}
	Keys struct{
		Tcp0x888 string `toml:"tcp0x888"`
		Udp0x10e string `toml:"udp0x10e"`
		Udp0x10f string `toml:"udp0x10f"`
	}
}{}

func Init(filename string) {
	pth, err := filepath.Abs(filename)
	util.MustNil(err)

	_, err = toml.DecodeFile(pth, &Conf)
	util.MustNil(err)

	L = logger.NewLogger("udp", Conf.Log.Addr, Conf.Log.Tag, Conf.Log.Level, 1)
	L.Sugar().Infof("conf init file %s value %+v", pth, Conf)
}
