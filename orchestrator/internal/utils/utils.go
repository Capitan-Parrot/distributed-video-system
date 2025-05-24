package utils

import (
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
)

// IsValidStatusTransition проверяет допустимость перехода между статусами
func IsValidStatusTransition(currentStatus, newStatus string) bool {
	transitions := map[string][]string{
		models.StatusInitStartup:          {models.StatusInStartupProcessing},
		models.StatusInStartupProcessing:  {models.StatusActive, models.StatusInactive}, // Может сразу перейти в inactive в случае ошибки
		models.StatusActive:               {models.StatusInitShutdown},
		models.StatusInitShutdown:         {models.StatusInShutdownProcessing},
		models.StatusInShutdownProcessing: {models.StatusInactive},
		models.StatusInactive:             {models.StatusInitStartup},
	}

	for _, allowedStatus := range transitions[currentStatus] {
		if allowedStatus == newStatus {
			return true
		}
	}
	return false
}
