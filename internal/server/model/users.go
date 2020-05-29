package model

import (
	"net"
	"sync"
	"time"
)

var mapUser sync.Map // mid *User

func AddUser(user *User) {
	if user == nil {
		return
	}
	mapUser.Store(user.Mid, user)
}

func DelUser(mid int32) {
	mapUser.Delete(mid)
}

func GetUser(mid int32) (*User, bool) {
	v, ok := mapUser.Load(mid)
	if !ok {
		return nil, false
	}

	user := v.(*User)
	user.LastTime = time.Now()
	return user, true
}

func GetUserWithConn(conn net.Conn) (*User, bool) {
	mid, ok := GetMid(conn)
	if !ok {
		return nil, false
	}
	return GetUser(mid)
}

func GetUsersWithSource(source int16) []*User {
	var sl = make([]*User, 0)
	mapUser.Range(func(key, value interface{}) bool {
		user := value.(*User)
		if user.Conn == nil {
			return true
		}

		if source == 2 || source == user.Source {
			sl = append(sl, user)
		}
		return true
	})
	return sl
}

func CountUser() (cnt int32) {
	mapUser.Range(func(key, value interface{}) bool {
		cnt++
		return true
	})
	return cnt
}
