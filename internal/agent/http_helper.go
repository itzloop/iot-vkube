package agent

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
)

type ControllerBody struct {
	Name      string
	Readiness bool
	Devices   []DeviceBody
}

type DeviceBody struct {
	Name      string
	Readiness bool
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
