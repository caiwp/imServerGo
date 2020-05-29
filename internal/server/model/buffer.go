package model

import "gogit.oa.com/March/gopkg/protocol/bypack"

var (
	BufferLoginSuccess []byte
	BufferMonitorTest  []byte
	BufferHeartbeat    []byte
)

func init() {
	setBufferLoginSuccess()
	setBufferMonitorTest()
	setBufferHeartbeat()
}

func setBufferLoginSuccess() {
	w := bypack.NewWriter(0x201)
	w.End()
	BufferLoginSuccess = w.GetBuffer()
}

func setBufferMonitorTest() {
	w := bypack.NewWriter(0x881)
	w.End()
	BufferMonitorTest = w.GetBuffer()
}

func setBufferHeartbeat() {
	w := bypack.NewWriter(0x1)
	w.End()
	BufferHeartbeat = w.GetBuffer()
}

func GetBufferAdmin(data string) []byte {
	return getStringBuffer(0x888, data)
}

func GetBufferOnlineNum() []byte {
	w := bypack.NewWriter(0x109)
	w.Int(CountMid())
	w.End()
	return w.GetBuffer()
}

func GetBufferJS(data string) []byte {
	return getStringBuffer(0x10E, data)
}

func GetBuffer887(data string) []byte {
	return getStringBuffer(0x887, data)
}

func GetBuffer886(data string) []byte {
	return getStringBuffer(0x886, data)
}

func getStringBuffer(cmd uint16, data string) []byte {
	w := bypack.NewWriter(cmd)
	w.String(data)
	w.End()
	return w.GetBuffer()
}
