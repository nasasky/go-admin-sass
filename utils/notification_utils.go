package utils

import (
	models "nasa-go-admin/model"
	"nasa-go-admin/services" // 确保使用正确的导入路径
)

type NotificationUtils struct {
	NotificationService *services.NotificationService
}

func NewNotificationUtils() (*NotificationUtils, error) {
	notificationService, err := services.NewNotificationService()
	if err != nil {
		return nil, err
	}
	return &NotificationUtils{
		NotificationService: notificationService,
	}, nil
}

func (n *NotificationUtils) PublishNotification(notification models.Notification) error {
	return n.NotificationService.PublishNotification(notification)
}

func (n *NotificationUtils) ConsumeNotifications() {
	n.NotificationService.ConsumeNotifications()
}

func (n *NotificationUtils) ProcessNotification(notification models.Notification) {
	n.NotificationService.ProcessNotification(notification)
}
