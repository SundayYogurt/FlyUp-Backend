package handlers

import (
	"encoding/json"
	"log"

	"github.com/SundayYogurt/FlyUp-Backend/mail-svc/internal/dto"
	"github.com/SundayYogurt/FlyUp-Backend/mail-svc/internal/services"
)

type MailHandler struct {
	MailService *services.MailService
}

func NewMailHandler(ms *services.MailService) *MailHandler {
	return &MailHandler{MailService: ms}
}

func (h *MailHandler) HandleMessage(message string) error {
	var event dto.VerifyEmailEvent

	if err := json.Unmarshal([]byte(message), &event); err != nil {
		log.Printf("invalid event payload: %s\n", message)
		return err
	}

	log.Printf("Verify email event received: user_id=%d email=%s",
		event.UserID, event.Email)

	log.Println("[MAIL] sending...")
	err := h.MailService.SendVerifyEmail(event.Email, event.Token)
	log.Println("[MAIL] send finished, err =", err)
	return err
}
