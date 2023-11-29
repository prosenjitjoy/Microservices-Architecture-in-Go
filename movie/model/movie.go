package model

import "main/metadata/model"

// MovieDetails includes movie metadata and its aggregated rating.
type MovieDetails struct {
	Rating   *float64       `json:"rating,omitempty"`
	Metadata model.Metadata `json:"metadata"`
}
