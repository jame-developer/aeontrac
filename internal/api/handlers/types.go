package handlers

// StartTrackingRequest defines the structure for time-related requests.
type StartTrackingRequest struct {
	Comment string `json:"time,omitempty"`
}
