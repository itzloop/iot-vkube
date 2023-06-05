package agent

import (
	"bytes"
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

func doPostRequest(url string, request, response interface{}) error {
	if response == nil {
		return errors.New("response is nil")
	}

	if request == nil {
		return errors.New("request is nil")
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
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
