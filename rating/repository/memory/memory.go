package memory

import (
	"context"
	"main/rating/model"
	"main/rating/repository"
)

// Repository defines a rating repository.
type Repository struct {
	data map[model.RecordType]map[model.RecordID][]model.Rating
}

// New creates a new memory repository.
func New() *Repository {
	return &Repository{
		data: map[model.RecordType]map[model.RecordID][]model.Rating{},
	}
}

// Get retrieves all ratings for a given record.
func (r *Repository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	if _, ok := r.data[recordType]; !ok {
		return nil, repository.ErrNotFound
	}
	return r.data[recordType][recordID], nil
}

// Put adds a rating for a given record.
func (r *Repository) Put(ctx context.Context, recordID model.RecordID, recortType model.RecordType, rating *model.Rating) error {
	if _, ok := r.data[recortType]; !ok {
		r.data[recortType] = map[model.RecordID][]model.Rating{}
	}
	r.data[recortType][recordID] = append(r.data[recortType][recordID], *rating)
	return nil
}
