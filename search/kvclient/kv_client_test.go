package kvclient

import (
	"testing"
	"time"
	"travel_ai_search/search/conf"
)

func TestKVClient(t *testing.T) {
	config := &conf.Config{
		RedisAddr: "127.0.0.1:9221",
	}
	client, err := InitClient(config)
	if err != nil {
		t.Error("init client err", err)
	}

	key := "test_key1"
	value := "aaa"
	err = client.Set(key, value, time.Hour)
	if err != nil {
		t.Error("set err", err)
	}

	value = "中文123"
	err = client.Set(key, value, time.Hour)
	if err != nil {
		t.Error("set err", err)
	}

	v, err := client.Get(key)
	if err != nil {
		t.Error("set err", err)
	}
	if v != value {
		t.Error(key, "'s value is not equal,src:", value, ",new:", v)
	}

	title := "北京一日"
	detail := "北京一日北京一日北京一日北京一日"
	err = client.HMSet("test1", "title", title, "detail", detail)
	if err != nil {
		t.Error("hset err", err)
	}
	values, err := client.HGetAll("test1")
	if err != nil {
		t.Error("hgetall err", err)
	}
	v, ok := values["title"]
	if !ok {
		t.Error("cann't find field :title")
	}
	if v != title {
		t.Error("field :title not equal,src:", title, ",new:", v)
	}
	v, ok = values["detail"]
	if !ok {
		t.Error("cann't find field :detail")
	}
	if v != detail {
		t.Error("field :detail not equal,src:", detail, ",new:", v)
	}
	client.Close()
}
