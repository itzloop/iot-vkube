package main

import (
	"flag"
	"html/template"
	"log"
	"os"
	"path"
)

func main() {
	p := flag.String("f", "/home/loop/p/iot-vkube/examples/sample.yaml.gotmpl", "manifest template path")
	deviceName := flag.String("d", "device_name", "device name")
	controllerName := flag.String("c", "controller_name", "controller name")
	controllerAddr := flag.String("caddr", "localhost:5000", "controller address")
	max := flag.Uint("max", 5, "max pods")

	flag.Parse()

	t := template.New(path.Base(*p))
	t.Funcs(template.FuncMap{
		"Iterate": func(count uint) []uint {
			var i uint
			var Items []uint
			for i = 0; i < count; i++ {
				Items = append(Items, i)
			}
			return Items
		},
	})
	template.Must(t.ParseFiles(*p))
	if err := t.Execute(os.Stdout, struct {
		DeviceName        string
		ControllerName    string
		ControllerAddress string
		Max               uint
	}{
		DeviceName:        *deviceName,
		ControllerName:    *controllerName,
		ControllerAddress: *controllerAddr,
		Max:               *max,
	}); err != nil {
		log.Fatal(err)
	}
}
