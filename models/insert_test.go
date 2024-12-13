package models

import (
	"Cobalt/dao"
	"testing"
)

func TestRedis(t *testing.T) {
	conn := dao.ConnRedis()
	RetrieveAndProcessData(conn, "208.67.222.222", "192.168.1.1")
}
