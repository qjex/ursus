package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"syscall"
	"ursus/control"
	"ursus/store"
)

type Config struct {
	ControlAddress string           `json:"control_address"`
	HttpAddress    string           `json:"http_address"`
	Mongo          store.ClientConf `json:"mongo"`
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
	if runtime.GOOS == "darwin" {
		rLimit.Cur = 12000
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	log.Printf("set cur limit: %d", rLimit.Cur)
}

func main() {
	setLimit()
	config := readConfig()
	store, err := store.NewMongoStore(&config.Mongo)
	if err != nil {
		log.Fatal("Couldn't create mongo connection: ", err)
	}
	server, err := control.NewServer(config.ControlAddress, 1, store)

	if err != nil {
		log.Fatal("Couldn't create control server: ", err)
	}
	server.Start()
	server.Stop()
	store.Close()
}
