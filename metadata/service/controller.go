package service

import (
	"context"
	"errors"
	"main/metadata/model"
	"main/metadata/repository"
)

// ErrNotFound is returned when a requested record is not found.
var ErrNotFound = errors.New("not found")

type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
	Put(ctx context.Context, id string, metadata *model.Metadata) error
}

// MetadataService defines a metadata service controller.
type MetadataService struct {
	repo metadataRepository
}

// New creates a metadata service controller.
func New(repo metadataRepository) *MetadataService {
	return &MetadataService{
		repo: repo,
	}
}

func (c *MetadataService) GetMetadata(ctx context.Context, id string) (*model.Metadata, error) {
	res, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return res, err
}

func (c *MetadataService) PutMetadata(ctx context.Context, id string, metadata *model.Metadata) error {
	return c.repo.Put(ctx, id, metadata)
}
