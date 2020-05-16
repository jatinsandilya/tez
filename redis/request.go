package redis

// Request is the standard request type
// used by this service
type Request struct {
	Key     string                 `json:"key" example:"test:key"`
	Options map[string]interface{} `json:"options"`
	Value   interface{}            `json:"value"`
}
