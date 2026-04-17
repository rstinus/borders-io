package ouvrier

import (
	"Jeu/engine"
	"Jeu/models"
	"math"
	"time"
)

var states = map[string]*DemolitionState{}

func ensureState(playerID string) *DemolitionState {
	s, ok := states[playerID]
	if !ok {
		s = &DemolitionState{
			Targets: [2]int{-1, -1},
		}
		states[playerID] = s
	}
	return s
}

func ExecuteInstantAttack(p *models.Player) bool {
	now := time.Now()
	s := ensureState(p.ID)

	// Vérifier le cooldown
	if now.Before(s.NextAvailable) {
		return false
	}

	pCenterX := p.X + 7.5
	pCenterY := p.Y + 7.5
	gridX, gridY := int(pCenterX/engine.SCALE), int(pCenterY/engine.SCALE)
	found := false

	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			nx, ny := gridX+dx, gridY+dy

			// Vérification des limites de la grille
			if nx < 0 || nx >= engine.GRID_WIDTH || ny < 0 || ny >= engine.GRID_HEIGHT {
				continue
			}

			idx := ny*engine.GRID_WIDTH + nx

			// Si la case contient un obstacle
			if engine.ObstaclesGrid[idx] > 0 {
				// Centre de la case d'obstacle (10x10 pixels)
				obsCenterX := float64(nx*engine.SCALE) + 5.0
				obsCenterY := float64(ny*engine.SCALE) + 5.0

				// Calcul de la distance (Formule utilisée pour les boosts)
				dist := math.Sqrt(math.Pow(pCenterX-obsCenterX, 2) + math.Pow(pCenterY-obsCenterY, 2))

				// Si la case est dans le rayon de 50 pixels
				if dist < 50.0 {
					// On supprime l'obstacle entier via la fonction créée précédemment
					engine.RemoveObstacleAt(idx)
					found = true
				}
			}
		}
	}

	if found {
		// Appliquer le cooldown uniquement si un mur a été touché
		s.NextAvailable = now.Add(demolitionCooldown)
		return true
	}

	return false
}
