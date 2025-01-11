package timex

import (
	"testing"
	"time"

	"github.com/go-playground/assert"
)

func TestJST(t *testing.T) {
	jstLocation := JST()
	assert.Equal(t, "Asia/Tokyo", jstLocation.String())
}

func TestJSTTimeConversion(t *testing.T) {
	jstLocation := JST()

	// UTCの時間を作成
	utc := time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC)

	// UTC時間をJSTに変換
	got := utc.In(JST())

	// 期待する結果 (JSTの時間は9時間進んでいる)
	expected := time.Date(2024, 10, 2, 9, 0, 0, 0, jstLocation)

	// JSTへの変換が正しいかを確認
	assert.Equal(t, expected, got)
}
