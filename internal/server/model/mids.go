package model

import (
	"net"
	"sync"
)

var mapMid sync.Map // conn mid

func AddMid(conn net.Conn, mid int32) {
	if conn == nil {
		return
	}
	mapMid.Store(conn, mid)
}

func DelMid(conn net.Conn) {
	if conn == nil {
		return
	}
	mapMid.Delete(conn)
}

func GetMid(conn net.Conn) (int32, bool) {
	v, ok := mapMid.Load(conn)
	if !ok {
		return 0, false
	}
	return v.(int32), true
}

func CountMid() (cnt int32) {
	mapMid.Range(func(key, value interface{}) bool {
		cnt++
		return true
	})
	return cnt
}
