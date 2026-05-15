package webhooks

import (
	"fmt"
	"time"

	"api/schemas"
)

func BuildDocumentEvent(eventType string, doc *schemas.Document, domain string) EventPayload {
	return EventPayload{
		EventID:    newEventID(),
		EventType:  eventType,
		OccurredAt: time.Now().UTC(),
		Document:   buildDocumentDTO(doc, domain),
	}
}

func BuildSignerEvent(eventType string, doc *schemas.Document, signer *schemas.Signer, domain string) EventPayload {
	payload := BuildDocumentEvent(eventType, doc, domain)
	dto := buildSignerDTO(signer, domain)
	payload.Signer = &dto
	return payload
}

func buildDocumentDTO(doc *schemas.Document, domain string) EventDocument {
	url := ""
	if domain != "" && doc.ID != 0 {
		url = fmt.Sprintf("%s/documents/%d", domain, doc.ID)
	}
	return EventDocument{
		ID:         doc.ID,
		Name:       doc.Name,
		Status:     doc.Status,
		FileName:   doc.FileName,
		URL:        url,
		Sequential: doc.Sequential,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}
}

func buildSignerDTO(signer *schemas.Signer, domain string) EventSigner {
	signingURL := ""
	if domain != "" && signer.Token != "" {
		signingURL = fmt.Sprintf("%s/share/%s", domain, signer.Token)
	}
	return EventSigner{
		ID:         signer.ID,
		Name:       signer.Name,
		Email:      signer.Email,
		Role:       signer.Role,
		Status:     signer.Status,
		OrderNum:   signer.OrderNum,
		SigningURL: signingURL,
		SignedAt:   signer.SignedAt,
		ViewedAt:   signer.ViewedAt,
	}
}
