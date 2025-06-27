package handler

import (
	"encoding/json"
	"log"
	"os"

	"github.com/streadway/amqp"
)

// EmailMessage is the payload for email jobs
type EmailMessage struct {
	To      string   `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	Files   []string `json:"files,omitempty"`
}

func StartEmailConsumer() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/" // fallback for dev
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	q, err := ch.QueueDeclare(
		"email.queue", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}
	log.Println("Email consumer started...")
	for d := range msgs {
		var msg EmailMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("Invalid email message: %v", err)
			continue
		}
		if err := sendEmail(msg); err != nil {
			log.Printf("Failed to send email: %v", err)
		}
	}
}

func sendEmail(msg EmailMessage) error {
	// SMTP credentials from your handler
	smtpHost := "smtp.zoho.com"
	smtpPort := "587"
	smtpUser := "e_book_aitu@zohomail.com"
	smtpPass := "gakon2006"
	from := smtpUser
	to := []string{msg.To}
	// Compose email (simple, no attachments for now)
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = msg.To
	headers["Subject"] = msg.Subject
	body := ""
	for k, v := range headers {
		body += k + ": " + v + "\r\n"
	}
	body += "\r\n" + msg.Body
	return smtpSend(smtpHost, smtpPort, smtpUser, smtpPass, from, to, []byte(body))
}

func smtpSend(host, port, user, pass, from string, to []string, msg []byte) error {
	return nil // TODO: implement using net/smtp (see your handler)
}
