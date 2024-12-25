package config

import (
	"Cobalt/system_struct"
	"github.com/BurntSushi/toml"
	"log"
)

func UseToml() system_struct.ConfigInfo {
	var c system_struct.ConfigInfo
	var path string = "./conf.toml"
	if _, err := toml.DecodeFile(path, &c); err != nil {
		log.Fatal(err)

	}
	//fmt.Println(c.DetectCycle * time.Second)
	return c
}
