package qlog

import (
	"testing"
	"time"
)

func TestQlog(t *testing.T) {

	Error("测试2", "12333333333333333333333333")
	Error("测试", "456")
	for {
		if len(LogChannel) == 0 {
			time.Sleep(3 * time.Second)
			return
		}
	}
}
