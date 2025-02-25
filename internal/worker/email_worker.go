package worker

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"sync"
)

// EmailJob holds all info required to send an email
type EmailJob struct {
	To       string
	Subject  string
	BodyHTML string // We'll send HTML
}

// EmailWorker is a simple worker that processes EmailJob from a channel
type EmailWorker struct {
	smtpHost string
	smtpPort string
	auth     smtp.Auth
	from     string

	jobs chan EmailJob
	wg   sync.WaitGroup
}

// NewEmailWorker constructs an EmailWorker
// Example: host=smtp.gmail.com, port=587, from=user@gmail.com, user + password from config
func NewEmailWorker(host, port, from, username, password string) *EmailWorker {
	w := &EmailWorker{
		smtpHost: host,
		smtpPort: port,
		from:     from,
		// Use PlainAuth for demonstration. Some providers might require OAuth2 or other methods.
		auth: smtp.PlainAuth("", username, password, host),
		jobs: make(chan EmailJob, 100), // capacity 100
	}
	w.startWorker()
	return w
}

// Start a single goroutine worker
func (w *EmailWorker) startWorker() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for job := range w.jobs {
			if err := w.sendEmail(job); err != nil {
				log.Printf("Failed to send email to %s: %v", job.To, err)
			} else {
				log.Printf("Successfully sent email to %s", job.To)
			}
		}
	}()
}

// Enqueue an email job
func (w *EmailWorker) Enqueue(job EmailJob) {
	w.jobs <- job
}

// Close gracefully
func (w *EmailWorker) Close() {
	close(w.jobs)
	w.wg.Wait()
}

func (w *EmailWorker) sendEmail(job EmailJob) error {
	// Construct the raw MIME message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", w.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", job.To))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", job.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(job.BodyHTML)

	addr := fmt.Sprintf("%s:%s", w.smtpHost, w.smtpPort)
	return smtp.SendMail(addr, w.auth, w.from, []string{job.To}, msg.Bytes())
}
