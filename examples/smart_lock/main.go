package main

import smart_lock "github.com/itzloop/iot-vkube/examples/smart_lock/src"

func main() {
	smart_lock.RunServer(":5000", "lc1")
}