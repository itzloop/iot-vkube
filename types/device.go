package types

type Device struct {
	Meta      map[string]string `json:"meta,omitempty"`
	Name      string            `json:"name"`
	Readiness bool              `json:"readiness"`
}
