package common

import "testing"

func TestSend(t *testing.T) {
	err := send("topic-test", "test")
	if err != nil {
		t.Errorf("send error %v", err)
	}
}
