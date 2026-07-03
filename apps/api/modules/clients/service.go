package clients

import (
	"context"
	stderrors "errors"
	"strings"
	"time"

	"api/internal/errors"
	"api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm *gorm.DB
}

func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func (s *Service) Create(ctx context.Context, ownerID int64, req *CreateClientRequest) (*ClientResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.Invalid("name is required")
	}

	record := schemas.Client{
		Name:    req.Name,
		Email:   req.Email,
		Company: req.Company,
		Phone:   req.Phone,
		Notes:   req.Notes,
		OwnerID: ownerID,
	}
	if err := s.orm.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, errors.Internal("failed to create client", err)
	}

	return toClientResponse(&record), nil
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]ClientResponse, error) {
	var records []schemas.Client
	if err := s.orm.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list clients", err)
	}

	out := make([]ClientResponse, len(records))
	for i := range records {
		out[i] = *toClientResponse(&records[i])
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, ownerID int64, clientID int64) (*ClientResponse, error) {
	record, err := s.findOwned(ctx, ownerID, clientID)
	if err != nil {
		return nil, err
	}
	return toClientResponse(record), nil
}

func (s *Service) Update(ctx context.Context, ownerID int64, clientID int64, req *UpdateClientRequest) (*ClientResponse, error) {
	record, err := s.findOwned(ctx, ownerID, clientID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.Name) != "" {
		record.Name = req.Name
	}
	record.Email = req.Email
	record.Company = req.Company
	record.Phone = req.Phone
	record.Notes = req.Notes

	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to update client", err)
	}

	return toClientResponse(record), nil
}

func (s *Service) Delete(ctx context.Context, ownerID int64, clientID int64) error {
	record, err := s.findOwned(ctx, ownerID, clientID)
	if err != nil {
		return err
	}

	if err := s.orm.WithContext(ctx).Model(&schemas.Document{}).
		Where("client_id = ? AND owner_id = ?", clientID, ownerID).
		Update("client_id", nil).Error; err != nil {
		return errors.Internal("failed to unlink documents from client", err)
	}

	if err := s.orm.WithContext(ctx).Delete(record).Error; err != nil {
		return errors.Internal("failed to delete client", err)
	}
	return nil
}

func (s *Service) findOwned(ctx context.Context, ownerID int64, clientID int64) (*schemas.Client, error) {
	var record schemas.Client
	err := s.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", clientID, ownerID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("client not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to read client", err)
	}
	return &record, nil
}

func toClientResponse(record *schemas.Client) *ClientResponse {
	return &ClientResponse{
		ID:        record.ID,
		Name:      record.Name,
		Email:     record.Email,
		Company:   record.Company,
		Phone:     record.Phone,
		Notes:     record.Notes,
		OwnerID:   record.OwnerID,
		CreatedAt: record.CreatedAt.Format(time.RFC3339),
		UpdatedAt: record.UpdatedAt.Format(time.RFC3339),
	}
}
