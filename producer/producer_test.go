package producer

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/wlcn/yq-colly/common"
)

const total = 15000

func TestSendSync(t *testing.T) {
	p := NewSyncProducer()
	defer CloseSync(p)
	err := SendSync(p, common.Topic, map[string]string{
		"age":  "18",
		"name": "中国",
	})
	if err != nil {
		t.Errorf("send error %v", err)
	}
}

func TestSendSyncPressure(t *testing.T) {
	p := NewSyncProducer()
	defer CloseSync(p)
	start := time.Now()
	for i := 0; i < total; i++ {
		err := SendSync(p, common.Topic, map[string]interface{}{
			"age":  strconv.Itoa(i),
			"name": "中国",
		})
		if err != nil {
			t.Errorf("send error %v", err)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("Sync num is %v, elapsed  is %v", total, elapsed)
}

func TestSendAsync(t *testing.T) {
	p := NewAsyncProducer()
	defer CloseAsync(p)
	err := SendAsync(p, common.Topic, map[string]string{
		"age":  "18",
		"name": "中国",
	})
	if err != nil {
		t.Errorf("send error %v", err)
	}
}

func TestSendAsyncPressure(t *testing.T) {
	p := NewAsyncProducer()
	defer CloseAsync(p)
	start := time.Now()
	for i := 0; i < total; i++ {
		err := SendAsync(p, common.Topic, map[string]interface{}{
			"age":  strconv.Itoa(i),
			"name": "中国",
		})
		if err != nil {
			t.Errorf("send error %v", err)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("Sync num is %v, elapsed  is %v", total, elapsed)
}
