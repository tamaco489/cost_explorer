package timex

import "time"

var jst *time.Location

func init() {
	var err error
	jst, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
}

// JST: jstタイムゾーンを返す。
func JST() *time.Location {
	return jst
}