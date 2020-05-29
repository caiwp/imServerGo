package conf

import (
	"path/filepath"
	"time"

	"gogit.oa.com/March/gopkg/metric"

	"go.uber.org/zap"
	"gogit.oa.com/March/gopkg/logger"

	"github.com/BurntSushi/toml"

	"gogit.oa.com/March/gopkg/util"
)

var L *zap.Logger

var (
	Version string
	StartAt time.Time
)

var Conf = struct {
	App struct {
		Name     string `toml:"name"`
		Env      string `toml:"env"`
		HttpAddr string `toml:"http_addr"`
	}
	Log struct {
		Tag   string `toml:"tag"`
		Addr  string `toml:"addr"`
		Level int    `toml:"level"`
	}
	Server struct {
		TcpAddr string `toml:"tcp_addr"`
		UdpAddr string `toml:"udp_addr"`
	}
	Keys struct {
		Tcp0x888 string `toml:"tcp0x888"`
		Udp0x10e string `toml:"udp0x10e"`
		Udp0x10f string `toml:"udp0x10f"`
	}
}{}

func Init(filename string) {
	StartAt = time.Now()

	pth, err := filepath.Abs(filename)
	util.MustNil(err)

	_, err = toml.DecodeFile(pth, &Conf)
	util.MustNil(err)

	L = logger.NewLogger("udp", Conf.Log.Addr, Conf.Log.Tag, Conf.Log.Level, 0)
	L.Sugar().Infof("conf init file %s value %+v", pth, Conf)

	metric.InitMetricsWithCode(Conf.App.Name, []string{"startedCounter", "handledCounter", "handledHistogram"})
}
