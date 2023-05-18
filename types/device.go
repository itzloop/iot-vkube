package types

type Device struct {
	Meta            map[string]interface{} `json:"meta,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Ready           bool                   `json:"ready,omitempty"`
	RegisteredHooks []Hook                 `json:"registered_hooks"`
}
