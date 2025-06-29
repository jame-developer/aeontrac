package models

type WorkTimeRequest struct {
    Date    string `json:"date"`
    Start   string `json:"start"`
    Stop    string `json:"stop"`
    Comment string `json:"comment"`
}