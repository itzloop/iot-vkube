package main

import (
	"flag"
	smart_lock "github.com/itzloop/iot-vkube/examples/smart_lock/src"
)

func main() {
	addr := flag.String("addr", ":5000", "server bind address")
	flag.Parse()
	smart_lock.RunServer(*addr)
}
