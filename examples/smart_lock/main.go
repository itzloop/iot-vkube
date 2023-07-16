package main

import (
	"flag"
	smart_lock "github.com/itzloop/iot-vkube/examples/smart_lock/src"
)

func main() {
	addr := flag.String("addr", ":5000", "server bind address")
	controllersCount := flag.Int("controllers", 1, "Controllers count")
	devicesPerController := flag.Int("devices", 10, "Devices per controller")
	flag.Parse()

	smart_lock.RunServer(*addr, *controllersCount, *devicesPerController)
}
