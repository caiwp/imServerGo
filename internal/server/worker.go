package server

import (
	"IMServer/internal/server/conf"
	"IMServer/internal/server/model"
	"encoding/json"
	"net"
	"time"

	"gogit.oa.com/March/gopkg/metric"

	"gogit.oa.com/March/gopkg/util"

	"gogit.oa.com/March/gopkg/protocol/bypack"
)

type Worker struct {
	conn   net.Conn
	addr   net.Addr
	reader *bypack.Reader
	user   *model.User
}

func NewWorker(addr net.Addr, reader *bypack.Reader) *Worker {
	return &Worker{
		addr:   addr,
		reader: reader,
	}
}

func NewWorkerWithConn(conn net.Conn, reader *bypack.Reader) *Worker {
	return &Worker{
		conn:   conn,
		reader: reader,
		addr:   conn.RemoteAddr(),
	}
}

func (w *Worker) IsLogin() (ok bool) {
	w.user, ok = model.GetUserWithConn(w.conn)
	return ok
}

func (w *Worker) IsLocal() bool {
	if w.addr == nil {
		return false
	}
	return util.IsLocal(w.addr)
}

func (w *Worker) send(data []byte) {
	if w.conn == nil {
		return
	}

	for i := 0; i < 3; i++ {
		_, err := w.conn.Write(data)
		if err != nil {
			conf.L.Warn(err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		return
	}
}

// 接收客户端发心跳 10s
func (w *Worker) TCP0x2() metric.Code {
	if !w.IsLogin() {
		return metric.CodeFailure
	}
	w.user.Heartbeat()
	return metric.CodeSuccess
}

// 监控测试
func (w *Worker) TCP0x881() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}

	w.send(model.BufferMonitorTest)
	return metric.CodeSuccess
}

// 重启 FIXME 不实现
func (w *Worker) TCP0x882() metric.Code {
	return metric.CodeFailure
}

// 系统信息
func (w *Worker) TCP0x888() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}

	key := w.reader.String()
	if key != conf.Conf.Keys.Tcp0x888 {
		conf.L.Sugar().Errorf("invalid key %s", key)
		_ = w.conn.Close()
		return metric.CodeFailure
	}

	var data = map[string]interface{}{
		"tableInfo": model.DataInfo(),
		"version":   conf.Version,
		"startAt":   conf.StartAt.Format("2006-01-02 15:04:05"),
	}
	b, err := json.Marshal(data)
	if err != nil {
		conf.L.Warn(err.Error())
		return metric.CodeFailure
	}

	w.send(model.GetBufferAdmin(string(b)))

	return metric.CodeSuccess
}

// 登录
func (w *Worker) TCP0x101() metric.Code {
	var mid = w.reader.Int()
	conf.L.Sugar().Infof("conn %s mid %d login", w.conn.RemoteAddr().String(), mid)
	if mid <= 0 {
		conf.L.Sugar().Warnf("invalid mid %s login", mid)
		return metric.CodeFailure
	}

	midTmp, ok := model.GetMid(w.conn)
	if ok && mid != midTmp {
		model.DelMid(w.conn)
		return metric.CodeFailure
	}

	model.AddMid(w.conn, mid)

	source := w.reader.Short()
	friends := getIdsByReader(w.reader, 1000) // 客户端其实不会传了

	user, ok := model.GetUser(mid)
	if !ok {
		user = model.NewUser(w.conn, mid)
	}

	user.Conn = w.conn
	user.Source = source
	user.Friends = friends

	w.send(model.BufferLoginSuccess)
	conf.L.Sugar().Infof("user %+v login success", user)
	return metric.CodeSuccess
}

// 退出
// 原有小喇叭没有处理，但有接收请求
// 加上退出桌子操作
func (w *Worker) TCP0x102() metric.Code {
	return metric.CodeSuccess
}

// 对单用户
func (w *Worker) TCP0x103() metric.Code {
	if !w.IsLogin() {
		return metric.CodeFailure
	}
	return w.x103()
}

// 对单用户
// FIXME 客户端貌似有问题
func (w *Worker) UDP0x103() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}
	return w.x103()
}

func (w *Worker) x103() metric.Code {
	w.reader.Int()
	toMid := w.reader.Int()
	user, ok := model.GetUser(toMid)
	if !ok {
		conf.L.Sugar().Warnf("user %d not found", toMid)
		return metric.CodeFailure
	}

	user.Send(w.reader.RawBuffer)

	return metric.CodeSuccess
}

// 广播所有人
func (w *Worker) TCP0x104() metric.Code {
	if !w.IsLogin() {
		return metric.CodeFailure
	}
	return w.x104(false)
}

// 广播所有人
func (w *Worker) UDP0x104() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}
	return w.x104(true)
}

func (w *Worker) x104(isUdp bool) metric.Code {
	buff := w.reader.RawBuffer
	var source int16 = 2

	if isUdp {
		source = w.reader.Short()
		str := w.reader.String()

		w := bypack.NewWriter(0x104)
		w.String(str)
		w.End()
		buff = w.GetBuffer()
	}

	users := model.GetUsersWithSource(source)
	for _, v := range users {
		v.Send(buff)
	}

	return metric.CodeSuccess
}

// 广播所有好友
// NOTE 好友在登录时传过来
func (w *Worker) TCP0x105() metric.Code {
	if !w.IsLogin() {
		return metric.CodeFailure
	}

	w.reader.Int()
	if len(w.user.Friends) == 0 {
		return metric.CodeSuccess
	}

	for _, v := range w.user.Friends {
		u, ok := model.GetUser(v)
		if !ok {
			continue
		}

		u.Send(w.reader.RawBuffer)
	}

	return metric.CodeSuccess
}

// 进入房间
func (w *Worker) TCP0x106() metric.Code {
	if !w.IsLogin() {
		return metric.CodeFailure
	}

	w.reader.Int()
	tid := w.reader.Int()
	w.user.EnterTable(tid)

	return metric.CodeSuccess
}

// 退出房间
func (w *Worker) TCP0x107() metric.Code {
	if !w.IsLogin() {
		return metric.CodeFailure
	}

	w.reader.Int()
	w.user.ExitTable()

	return metric.CodeSuccess
}

// 设置坐下状态 0 站起 1 坐下
// FIXME 客户端的逻辑有误
func (w *Worker) TCP0x108() metric.Code {
	if !w.IsLogin() {
		return metric.CodeFailure
	}

	play := w.reader.Short()
	w.user.UpdatePlay(play)

	return metric.CodeSuccess
}

// 在线人数
func (w *Worker) TCP0x109() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}

	w.send(model.GetBufferOnlineNum())

	return metric.CodeSuccess
}

// 从 PHP 发出加密单播 推 JS
func (w *Worker) UDP0x10e() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}

	mid := w.reader.Int()
	key := w.reader.String()
	if key != conf.Conf.Keys.Udp0x10e {
		conf.L.Sugar().Errorf("invalid key %s", key)
		return metric.CodeFailure
	}

	data := w.reader.String()
	user, ok := model.GetUser(mid)
	if !ok {
		conf.L.Sugar().Warnf("user %d not found data %s", mid, data)
		return metric.CodeFailure
	}
	user.Send(model.GetBufferJS(data))

	return metric.CodeSuccess
}

// 从 PHP 发出全桌广播 推 JS
func (w *Worker) UDP0x10f() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}

	tid := w.reader.Int()
	key := w.reader.String()
	if tid == 0 || key != conf.Conf.Keys.Udp0x10f {
		conf.L.Sugar().Errorf("invalid key %s", key)
		return metric.CodeFailure
	}

	play := w.reader.Short()
	data := w.reader.String()

	table := model.GetTable(tid)
	var i int
	table.Users.Range(func(key, value interface{}) bool {
		user := value.(*model.User)
		if play == model.PlayAll || play == user.Play {
			user.Send(model.GetBufferJS(data))
		}

		i++
		if i > 200 {
			return false
		}
		return true
	})

	return metric.CodeSuccess
}

// 批量获取 ID 在线状态 支持一次最多1000个 ID
func (w *Worker) TCP0x110() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}

	writer := bypack.NewWriter(0x110)

	mids := getIdsByReader(w.reader, 1000)
	for _, v := range mids {
		writer.Int(v)

		var status byte = 0 // 0 离线 1 大厅 2 旁观 3 在玩
		user, ok := model.GetUser(v)
		if !ok || user.Conn == nil {
			writer.Byte(status)
			continue
		}

		if user.Table == nil {
			status = 1
		} else {
			status = 2
			if user.Play == model.PlaySit {
				status = 3
			}
		}

		writer.Byte(status)
	}

	writer.End()
	w.send(writer.GetBuffer())

	return metric.CodeSuccess
}

// 获取单 tid 里玩家状态
func (w *Worker) TCP0x887() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}

	typ := w.reader.Short() // 1 坐下在玩 2 旁观 3 所有
	tid := w.reader.Int()

	table := model.GetTable(tid)
	data := table.GetUsersMidGroupByPlay(typ)
	if len(data) == 0 {
		return metric.CodeSuccess
	}

	b, err := json.Marshal(data)
	if err != nil {
		conf.L.Warn(err.Error())
		return metric.CodeFailure
	}

	w.send(model.GetBuffer887(string(b)))

	return metric.CodeSuccess
}

// 获取多 tid 里玩家状态
func (w *Worker) TCP0x886() metric.Code {
	if !w.IsLocal() {
		return metric.CodeFailure
	}

	typ := w.reader.Short() // 1 坐下在玩 2 旁观 3 所有
	tids := getIdsByReader(w.reader, 1000)

	var res = make(map[int32]map[int16][]int32)
	for _, v := range tids {
		table := model.GetTable(v)
		data := table.GetUsersMidGroupByPlay(typ)
		if len(data) == 0 {
			continue
		}

		res[v] = data
	}

	b, err := json.Marshal(res)
	if err != nil {
		conf.L.Warn(err.Error())
		return metric.CodeFailure
	}

	w.send(model.GetBuffer886(string(b)))

	return metric.CodeSuccess
}
