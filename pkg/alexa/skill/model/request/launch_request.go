package request

type LaunchRequest struct {
	BaseRequest
	Task struct {
		Name    string                 `json:"name"`
		Version string                 `json:"version"`
		Input   map[string]interface{} `json:"input"`
	} `json:"task"`
}
