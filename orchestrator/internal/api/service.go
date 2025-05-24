package api

import (
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/database"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/s3"
)

type Handlers struct {
	db *database.Database
	s3 *s3.Client
}

func NewHandlers(db *database.Database, s3 *s3.Client) *Handlers {
	return &Handlers{db: db, s3: s3}
}
