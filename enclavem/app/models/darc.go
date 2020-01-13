package models

// DarcPostResponse is sent back when using the POST /darcs endpoint
type DarcPostResponse struct {
	DarcID string `json:"darcID"`
}
