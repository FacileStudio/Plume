package documents

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// clientFor returns the id of the owner's client with this email, creating the
// client when there is none. The match ignores case, so a caller sending the
// address in another case does not split one customer in two.
func (s *Service) clientFor(ctx context.Context, ownerID int64, c *SignatureClient) (int64, error) {
	var client schemas.Client
	err := s.orm.WithContext(ctx).
		Where("owner_id = ? AND lower(email) = lower(?)", ownerID, c.Email).
		First(&client).Error
	if err == nil {
		return client.ID, nil
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.Internal("failed to look up the client", err)
	}
	client = schemas.Client{Name: c.Name, Email: c.Email, Company: c.Company, OwnerID: ownerID}
	if err := s.orm.WithContext(ctx).Create(&client).Error; err != nil {
		return 0, errors.Internal("failed to create the client", err)
	}
	return client.ID, nil
}
