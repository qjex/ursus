package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"syscall"
	"ursus/control"
)

type MongoConf struct {
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}

type Config struct {
	ControlAddress string    `json:"control_address"`
	HttpAddress    string    `json:"http_address"`
	Mongo          MongoConf `json:"mongo"`
}

func readConfig() Config {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	c := Config{}
	err := decoder.Decode(&c)
	if err != nil {
		log.Fatal("Error reading config: ", err)
	}
	return c
}

func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	log.Printf("set cur limit: %d", rLimit.Cur)
}

func main() {
	setLimit()
	config := readConfig()
	server := control.HbRcv
}
