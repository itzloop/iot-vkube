package types

type Device struct {
	Meta  map[string]string `json:"meta,omitempty"`
	Name  string            `json:"name,omitempty"`
	Ready bool              `json:"ready,omitempty"`
}
