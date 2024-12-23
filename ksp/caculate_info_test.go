package test_info

import (
	"fmt"
	"testing"
)

func TestCaculateInfo(t *testing.T) {
	r := CaculateInfo()
	err := r.Run(":8080")
	if err != nil {
		fmt.Println("err:", err)
	}
}
func TestCaculateInfo_1(t *testing.T) {
	r := CaculateInfo_1()
	err := r.Run(":8080")
	if err != nil {
		fmt.Println("err:", err)
	}
}
func TestCaculateInfo_2(t *testing.T) {
	r := CaculateInfo_2()
	err := r.Run(":8080")
	if err != nil {
		fmt.Println("err:", err)
	}
}
func TestPrintInfo(t *testing.T) {
	PrintInfo("topology_info_2")
}
