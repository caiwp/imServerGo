package model

import (
	"sync"
	"time"
)

var mapTable sync.Map // tid *Table

func AddTable(table *Table) {
	if table == nil {
		return
	}
	mapTable.Store(table.Tid, table)
}

func DelTable(table *Table) {
	table.Users.Range(func(key, value interface{}) bool {
		user := value.(*User)
		user.ExitTable()
		return true
	})
	mapTable.Delete(table.Tid)
}

func GetTable(tid int32) *Table {
	v, ok := mapTable.Load(tid)
	if !ok {
		return NewTable(tid)
	}

	table := v.(*Table)
	table.LastTime = time.Now()
	return table
}

func CountTable() (cnt int32) {
	mapTable.Range(func(key, value interface{}) bool {
		cnt++
		return true
	})
	return cnt
}
