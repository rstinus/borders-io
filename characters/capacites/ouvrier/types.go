package ouvrier

import "time"

const (
	demolitionCooldown = 4 * time.Second
)

type DemolitionState struct {
	Active        bool
	Targets       [2]int // obstacle indices
	LastTick      time.Time
	NextAvailable time.Time
}
