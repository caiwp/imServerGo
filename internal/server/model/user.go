package model

import (
	"IMServer/internal/server/conf"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

type User struct {
	Conn     net.Conn
	Mid      int32
	Source   int16 // 0 移动 1 PC
	Play     int16 // 0 旁观 1 在玩
	Table    *Table
	Friends  []int32
	LastTime time.Time
}

func NewUser(conn net.Conn, mid int32) *User {
	user := &User{
		Conn:     conn,
		Mid:      mid,
		LastTime: time.Now(),
	}
	AddUser(user)
	return user
}

func (u *User) LogField() zap.Field {
	return zap.String("mid", fmt.Sprintf("%d[%s]", u.Mid, u.Conn.RemoteAddr().String()))
}

func (u *User) Send(data []byte) {
	if u.Conn == nil {
		return
	}

	if err := u.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		conf.L.Warn(err.Error(), u.LogField())
		return
	}

	if _, err := u.Conn.Write(data); err != nil {
		conf.L.Warn(err.Error(), u.LogField())
		return
	}
}

func (u *User) EnterTable(tid int32) {
	if tid == 0 {
		return
	}

	// 原有桌子
	if u.Table != nil && u.Table.Tid == tid {
		return
	}

	// 旧桌子清除
	if u.Table != nil {
		u.Table.DelUser(u.Mid)
	}

	table := GetTable(tid)
	table.AddUser(u)
	u.Table = table
	u.Play = PlayLook
}

func (u *User) ExitTable() {
	if u.Table == nil {
		return
	}

	u.Table.DelUser(u.Mid)
	u.Table = nil
	u.Play = PlayLook
}

func (u *User) UpdatePlay(play int16) {
	if play != PlaySit {
		play = PlayLook
	}
	u.Play = play
}

func (u *User) Heartbeat() {
	now := time.Now()
	u.LastTime = now
	if u.Table != nil {
		u.Table.LastTime = now
	}
}
