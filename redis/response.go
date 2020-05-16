package redis

// Response is the general response type
// returned by this service
type Response struct {
	Status  string      `json:"status" example:"ok" `
	Code    int         `json:"code" example:"200" `
	Message string      `json:"message" example:"Key unavailable" `
	Payload interface{} `json:"payload,omitempty" example:"{[1,2]}" `
}
