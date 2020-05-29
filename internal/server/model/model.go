package model

import (
	"IMServer/internal/server/conf"
	"time"
)

const (
	PlayLook int16 = 0 + iota // 旁观 站起
	PlaySit                   // 坐下
	PlayAll                   // 全桌
)

func DataInfo() map[string]int32 {
	return map[string]int32{
		"conn_count":  CountConn(),
		"mid_count":   CountMid(),
		"user_count":  CountUser(),
		"table_count": CountTable(),
	}
}

func RunTicker(done chan struct{}) {
	var tk = time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-tk.C:
			var Data = struct {
				TableCnt    int
				TableDelCnt int
				UserCnt     int
				UserDelCnt  int
			}{}

			var now = time.Now()
			mapTable.Range(func(key, value interface{}) bool {
				Data.TableCnt++
				table := value.(*Table)
				if table.LastTime.Add(30 * time.Minute).Before(now) { // 30分钟没有更新
					var isNotEmpty bool
					table.Users.Range(func(key, value interface{}) bool {
						isNotEmpty = true
						return false
					})
					if !isNotEmpty {
						Data.TableDelCnt++
						DelTable(table)
					}
				}
				return true
			})

			mapUser.Range(func(key, value interface{}) bool {
				Data.UserCnt++
				user := value.(*User)
				if user.LastTime.Add(10 * time.Minute).Before(now) { // 10分钟没有更新
					Data.UserDelCnt++
					DelUser(user.Mid)
				}
				return true
			})
			conf.L.Sugar().Infof("clear data %+v", Data)

		case <-done:
			conf.L.Info("ticker stop")
			tk.Stop()
			return
		}
	}
}
