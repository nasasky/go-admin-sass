package model

type Notification struct {
	NotificationID string `json:"notification_id"`
	UserID         string `json:"user_id"`
	Message        string `json:"message"`
	Timestamp      string `json:"timestamp"`
}
