package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/repository"
	"github.com/ghstouch/one-go/pkg/logger"
)

// WebhookService defines webhook operations
type WebhookService interface {
	Create(webhook *model.Webhook) error
	GetByID(id string) (*model.Webhook, error)
	List() ([]model.Webhook, error)
	Update(webhook *model.Webhook) error
	Delete(id string) error
	ListLogs(webhookID string, limit int) ([]model.WebhookLog, error)
	Dispatch(event string, payload interface{})
}

type webhookService struct {
	repo repository.WebhookRepository
}

func NewWebhookService(repo repository.WebhookRepository) WebhookService {
	return &webhookService{repo: repo}
}

func (s *webhookService) Create(webhook *model.Webhook) error {
	if webhook.Name == "" {
		return fmt.Errorf("webhook name is required")
	}
	if webhook.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	return s.repo.Create(webhook)
}

func (s *webhookService) GetByID(id string) (*model.Webhook, error) {
	return s.repo.GetByID(id)
}

func (s *webhookService) List() ([]model.Webhook, error) {
	return s.repo.List()
}

func (s *webhookService) Update(webhook *model.Webhook) error {
	return s.repo.Update(webhook)
}

func (s *webhookService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *webhookService) ListLogs(webhookID string, limit int) ([]model.WebhookLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListLogs(webhookID, limit)
}

// Dispatch sends a webhook notification for the given event
func (s *webhookService) Dispatch(event string, payload interface{}) {
	go func() {
		hooks, err := s.repo.GetByEvent(event)
		if err != nil {
			logger.Warnf("Failed to get webhooks for event %s: %v", event, err)
			return
		}

		body, err := json.Marshal(map[string]interface{}{
			"event":     event,
			"timestamp": time.Now().Format(time.RFC3339),
			"data":      payload,
		})
		if err != nil {
			return
		}

		for _, hook := range hooks {
			go s.send(&hook, event, body)
		}
	}()
}

func (s *webhookService) send(hook *model.Webhook, event string, body []byte) {
	req, err := http.NewRequest("POST", hook.URL, bytes.NewReader(body))
	if err != nil {
		s.logDelivery(hook.ID, event, 0, false, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", event)

	// Sign with secret if configured
	if hook.Secret != "" {
		sig := hmacSign(body, hook.Secret)
		req.Header.Set("X-Webhook-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	now := time.Now()
	if err != nil {
		hook.FailCount++
		hook.LastStatus = 0
		hook.LastSentAt = &now
		s.repo.Update(hook)
		s.logDelivery(hook.ID, event, 0, false, err.Error())
		return
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	hook.LastStatus = resp.StatusCode
	hook.LastSentAt = &now
	if success {
		hook.FailCount = 0
	} else {
		hook.FailCount++
	}
	s.repo.Update(hook)
	s.logDelivery(hook.ID, event, resp.StatusCode, success, "")
}

func (s *webhookService) logDelivery(webhookID, event string, statusCode int, success bool, errMsg string) {
	s.repo.CreateLog(&model.WebhookLog{
		WebhookID:  webhookID,
		Event:      event,
		StatusCode: statusCode,
		Success:    success,
		Error:      errMsg,
	})
}

func hmacSign(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
