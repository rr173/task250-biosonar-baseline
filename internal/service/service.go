// Package service orchestrates the domain packages (geometry, echo, classify,
// segment, versioning) over the persistence store, enforcing cross-entity
// invariants such as batch immutability and sealed-batch protection.
package service

import (
	"task250-biosonar/internal/geometry"
	"task250-biosonar/internal/segment"
	"task250-biosonar/internal/store"
)

// Service is the application façade used by the HTTP layer.
type Service struct {
	store         *store.Store
	geoPitchLimit float64
	geoRollLimit  float64
	mergeCfg      segment.MergeConfig
}

// New constructs a Service over the given store with default tuning.
func New(st *store.Store) *Service {
	return &Service{
		store:         st,
		geoPitchLimit: geometry.DefaultPitchLimitDeg,
		geoRollLimit:  geometry.DefaultRollLimitDeg,
		mergeCfg:      segment.DefaultMergeConfig(),
	}
}

// Store exposes the underlying persistence handle for thin CRUD handlers.
func (svc *Service) Store() *store.Store { return svc.store }
