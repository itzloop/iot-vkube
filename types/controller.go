package types

type Controller struct {
	Host    string            `json:"host,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
	Name    string            `json:"name,omitempty"`
	Ready   bool              `json:"ready,omitempty"`
	Devices []Device          `json:"devices"`
}
