package model

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type Table struct {
	Tid      int32
	Users    sync.Map // mid *User
	LastTime time.Time
}

func (t *Table) AddUser(user *User) {
	if user == nil {
		return
	}

	t.Users.Store(user.Mid, user)
}

func (t *Table) DelUser(mid int32) {
	t.Users.Delete(mid)
}

func (t *Table) LogField() zap.Field {
	return zap.Int32("tid", t.Tid)
}

func (t *Table) GetUsersMidGroupByPlay(typ int16) map[int16][]int32 {
	var res = make(map[int16][]int32)
	t.Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		if user.Conn == nil {
			return true
		}

		if typ == 3 || (typ == 1 && user.Play == PlaySit) || (typ == 2 && user.Play == PlayLook) {
			if len(res[user.Play]) == 0 {
				res[user.Play] = make([]int32, 0)
			}
			res[user.Play] = append(res[user.Play], user.Mid)
		}
		return true
	})
	return res
}

func NewTable(tid int32) *Table {
	table := &Table{
		Tid:      tid,
		LastTime: time.Now(),
	}
	AddTable(table)
	return table
}
