package watchdog

import (
	"context"
	"log"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/database"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
)

const watchInterval = 30 * time.Second

type Watchdog struct {
	db *database.Database
}

func New(db *database.Database) *Watchdog {
	return &Watchdog{
		db: db,
	}
}

func (w *Watchdog) Start(ctx context.Context) {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Watchdog stopped")
			return
		case <-ticker.C:
			w.checkScenarios(ctx)
		}
	}
}

func (w *Watchdog) checkScenarios(ctx context.Context) {
	scenarios, err := w.db.FindStuckScenarios(ctx, watchInterval)
	if err != nil {
		log.Printf("Failed to find stuck scenarios: %v", err)
		return
	}

	for _, scenario := range scenarios {
		log.Printf("Found stuck scenario %s, sending restart command", scenario.ID)

		if err := w.db.AddToOutbox(ctx, scenario.ID, models.CommandStart); err != nil {
			log.Printf("Failed to add restart command to outbox: %v", err)
			continue
		}

		if err := w.db.UpdateScenarioStatus(ctx, scenario.ID, models.StatusInitStartup); err != nil {
			log.Printf("Failed to update scenario status: %v", err)
		}
	}
}
