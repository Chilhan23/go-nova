package ai

type ContextMessage struct {
	Role    string `json:"role"`    // user / model
	Content string `json:"content"`
}

type AIRequest struct {
	Prompt    string           `json:"prompt"`
	History   []ContextMessage `json:"history"`
	ImagePath string           `json:"image_path,omitempty"`
}

type AIResponse struct {
	Text string `json:"text"`
}
