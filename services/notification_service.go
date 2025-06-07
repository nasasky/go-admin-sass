package services

import (
	"encoding/json"
	"log"
	models "nasa-go-admin/model" // 确保使用正确的导入路径

	"github.com/streadway/amqp"
)

type NotificationService struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func NewNotificationService() (*NotificationService, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"notification_queue", // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return nil, err
	}

	return &NotificationService{
		Conn:    conn,
		Channel: ch,
		Queue:   q,
	}, nil
}

func (s *NotificationService) ProcessNotification(notification models.Notification) {
	// 处理通知的逻辑
	log.Printf("Processing notification: %+v\n", notification)
	// 这里可以添加更多的通知处理逻辑，例如发送邮件、推送消息等
}

func (s *NotificationService) PublishNotification(notification models.Notification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	err = s.Channel.Publish(
		"",           // exchange
		s.Queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	return err
}

func (s *NotificationService) ConsumeNotifications() {
	msgs, err := s.Channel.Consume(
		s.Queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			var notification models.Notification
			if err := json.Unmarshal(d.Body, &notification); err != nil {
				log.Printf("Error decoding JSON: %v", err)
				continue
			}
			s.ProcessNotification(notification)
		}
	}()
}
