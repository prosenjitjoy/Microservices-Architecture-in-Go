package model

import "main/rpc"

// MetadataToProto converts a Metadata struct into a generated proto counterpart.
func MetadataToProto(m *Metadata) *rpc.Metadata {
	return &rpc.Metadata{
		MovieId:     m.ID,
		Title:       m.Title,
		Description: m.Description,
		Director:    m.Director,
	}
}

// MetadataFromProto converts generated proto counterpart into a Metadata struct.
func MetadataFromProto(m *rpc.Metadata) *Metadata {
	return &Metadata{
		ID:          m.MovieId,
		Title:       m.Title,
		Description: m.Description,
		Director:    m.Director,
	}
}
