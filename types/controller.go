package types

type Controller struct {
	Host            string                 `json:"host,omitempty"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Ready           bool                   `json:"ready,omitempty"`
	RegisteredHooks []Hook                 `json:"registered_hooks,omitempty"`
	// TODO do we need list of devices here?
}
