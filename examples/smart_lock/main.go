package main

import (
	"flag"
	smart_lock "github.com/itzloop/iot-vkube/examples/smart_lock/src"
)

func main() {
	controllerName := flag.String("cname", "lc1", "controller name")
	flag.Parse()
	smart_lock.RunServer(":5000", *controllerName)
}
