package common

import (
	"reflect"
	"testing"
)

func TestItem(t *testing.T) {
	i := Item{}
	// 零值测试
	// fmt.Println(reflect.TypeOf(i.Title))
	// fmt.Println(reflect.TypeOf(i.Description))
	if reflect.String != reflect.TypeOf(i.Title).Kind() {
		t.Error("title type is not expected")
	}
	if reflect.Map != reflect.TypeOf(i.Description).Kind() {
		t.Error("Description type is not expected")
	}
	if i.Title != "" {
		t.Error("title is not expected")
	}
	if i.Description != nil {
		t.Error("Description is not expected")
	}

}
