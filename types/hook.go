package types

import (
	"fmt"
	"net/http"
)

type Hook uint8

const (
	HookUnknown Hook = iota
	HookGetController
	HookListDevices
	HookCreateDevice
	HookGetDevice
	HookUpdateDevice
	HookDeleteDevice
)

var HookNames = map[uint8]string{
	0: "HookUnknown",
	1: "HookGetController",
	2: "HookListDevices",
	3: "HookCreateDevice",
	4: "HookGetDevice",
	5: "HookUpdateDevice",
	6: "HookDeleteDevice",
}

var HookValues = map[string]uint8{
	"HookUnknown":       0,
	"HookGetController": 1,
	"HookListDevices":   2,
	"HookCreateDevice":  3,
	"HookGetDevice":     4,
	"HookUpdateDevice":  5,
	"HookDeleteDevice":  6,
}

func (x Hook) String() string {
	return HookNames[uint8(x)]
}

func (x Hook) UrlAndMethod(base, controllerName, deviceName string) (string, string) {
	switch x {
	case HookGetController:
		return fmt.Sprintf("%s/controllers/%s", base, controllerName), http.MethodGet
	case HookListDevices:
		return fmt.Sprintf("%s/controllers/%s/devices", base, controllerName), http.MethodGet
	case HookCreateDevice:
		return fmt.Sprintf("%s/controllers/%s", base, controllerName), http.MethodPost
	case HookGetDevice:
		return fmt.Sprintf("%s/controllers/%s/devices/%s", base, controllerName, deviceName), http.MethodGet
	case HookUpdateDevice:
		return fmt.Sprintf("%s/controllers/%s/devices/%s", base, controllerName, deviceName), http.MethodPatch
	case HookDeleteDevice:
		return fmt.Sprintf("%s/controllers/%s/devices/%s", base, controllerName, deviceName), http.MethodDelete
	}

	return "", ""
}

func (x Hook) Method() string

func ParseHook(hook string) Hook {
	hookVal, ok := HookValues[hook]
	if !ok {
		return HookUnknown
	}

	return Hook(hookVal)
}
