package types

type Controller struct {
	Host      string            `json:"host,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	Name      string            `json:"name"`
	Readiness bool              `json:"readiness"`
	Devices   []Device          `json:"devices"`
}
