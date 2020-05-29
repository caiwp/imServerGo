package model

import (
	"net"
	"sync"
	"time"
)

var mapConn sync.Map

func AddConn(conn net.Conn) {
	mapConn.Store(conn, time.Now())
}

// 断开连接会调
func DelConn(conn net.Conn) {
	DelMid(conn)

	mapConn.Delete(conn)
	_ = conn.Close()
}

func CountConn() (cnt int32) {
	mapConn.Range(func(key, value interface{}) bool {
		cnt++
		return true
	})
	return cnt
}

func Heartbeat(done chan struct{}) {
	var tk = time.NewTicker(10 * time.Second)
	for {
		select {
		case <-tk.C:
			var now = time.Now()
			mapConn.Range(func(key, value interface{}) bool {
				conn := key.(net.Conn)
				tm := value.(time.Time)

				if tm.Add(60 * time.Second).Before(now) {
					DelConn(conn)
					return true
				}

				user, ok := GetUserWithConn(conn)
				if !ok || user.Mid == 0 || user.Conn == nil {
					DelConn(conn)
					return true
				}

				_, err := conn.Write(BufferHeartbeat)
				if err != nil {
					DelConn(conn)
					return true
				}

				return true
			})

		case <-done:
			tk.Stop()
			return
		}
	}
}
