package server

import "gogit.oa.com/March/gopkg/protocol/bypack"

func getIdsByReader(reader *bypack.Reader, cnt int) (ids []int32) {
	defer func() {
		recover()
	}()

	for i := 0; i < cnt; i++ {
		ids = append(ids, reader.Int())
	}
	return ids
}
