package agent

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
)

type ControllerListBody struct {
	Controllers []ControllerBody
}

type DeviceListBody struct {
	ControllerBody
	Devices []DeviceBody `json:"devices,omitempty"`
}

type ControllerBody struct {
	Name      string `json:"name,omitempty"`
	Readiness bool   `json:"readiness,omitempty"`
}

type DeviceBody struct {
	Name      string `json:"name,omitempty"`
	Readiness bool   `json:"readiness,omitempty"`
}

func doGetRequest(url string, response interface{}) error {
	if response == nil {
		return errors.New("response is nil")
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("request failed: %s", string(b))
	}

	// marshal body
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, response)
}
