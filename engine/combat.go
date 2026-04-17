// Fichier : engine/combat_new.go
package engine

import (
	"Jeu/models"
	"math"
	"math/rand"
	"time"
)

const (
	CollisionThreshold = 5.0
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// CheckSelfCollision vérifie si le joueur touche sa propre traînée (suicide)
func CheckSelfCollision(player *models.Player) bool {
	if player == nil || !player.IsAlive || len(player.Trail) < 5 {
		return false
	}

	headX := player.X + 7.5
	headY := player.Y + 7.5

	// Vérifier contre tous les points de la traînée sauf les 2 derniers (pour éviter faux positifs)
	for i := 0; i < len(player.Trail)-5; i++ {
		trailPoint := player.Trail[i]
		distance := math.Sqrt(
			math.Pow(headX-trailPoint.X, 2) +
				math.Pow(headY-trailPoint.Y, 2),
		)

		if distance <= CollisionThreshold {
			return true
		}
	}

	return false
}

func CheckTrailCollisionWithOthers(attacker *models.Player, allPlayers map[string]*models.Player) *models.Player {
	if attacker == nil || !attacker.IsAlive {
		return nil
	}

	attackerCenterX := attacker.X + 7.5
	attackerCenterY := attacker.Y + 7.5

	for _, victim := range allPlayers {
		if victim == nil || victim.ID == attacker.ID || !victim.IsAlive || len(victim.Trail) == 0 {
			continue
		}

		for _, trailPoint := range victim.Trail {
			distance := math.Sqrt(
				math.Pow(attackerCenterX-trailPoint.X, 2) +
					math.Pow(attackerCenterY-trailPoint.Y, 2),
			)

			if distance <= CollisionThreshold {
				return victim
			}
		}
	}

	return nil
}

// KillPlayer marque un joueur comme mort et nettoie ses traces
func KillPlayer(player *models.Player, reason string) {
	if player == nil {
		return
	}

	player.IsAlive = false
	player.DeathReason = reason
	player.Trail = []models.Point{}
	player.CurrentDirection = ""

	// Nettoyer son territoire de la grille
	for i := 0; i < len(Grid); i++ {
		if Grid[i] == int16(player.PIdx) {
			Grid[i] = 0
		}
	}
}

// RemoveDeadPlayers supprime les joueurs morts du map de joueurs
func RemoveDeadPlayers(players map[string]*models.Player) {
	for id, player := range players {
		if player != nil && !player.IsAlive {
			delete(players, id)
		}
	}
}

// FindValidSpawnPosition trouve une position de spawn aléatoire
// avec vérification de distance minimale par rapport aux autres joueurs
func FindValidSpawnPosition(existingPlayers map[string]*models.Player) (int, int) {
	const (
		minDistanceCells = 20 // Distance minimale en cellules (300 pixels)
		maxAttempts      = 50 // Nombre maximum de tentatives
		margin           = 10 // Marge par rapport aux bords
	)

	// Dimensions de la grille en cellules
	maxX := GRID_WIDTH - (2 * margin)
	maxY := GRID_HEIGHT - (2 * margin)

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Position aléatoire
		spawnX := margin + rand.Intn(maxX)
		spawnY := margin + rand.Intn(maxY)

		// Vérifier la distance avec tous les joueurs existants
		validSpawn := true

		// NOUVEAU : Vérifier si le point de spawn est sur un obstacle
		if CheckObstacleCollision(float64(spawnX*SCALE)+5.0, float64(spawnY*SCALE)+5.0) {
			validSpawn = false // Recommence une tentative
		}

		for _, player := range existingPlayers {
			if player == nil || !player.IsAlive {
				continue
			}

			// Position du joueur existant en cellules
			playerCellX := int(player.X / SCALE)
			playerCellY := int(player.Y / SCALE)

			// Calculer la distance
			dx := spawnX - playerCellX
			dy := spawnY - playerCellY
			distance := math.Sqrt(float64(dx*dx + dy*dy))

			// Si trop proche, essayer une autre position
			if distance < minDistanceCells {
				validSpawn = false
				break
			}
		}

		if validSpawn {
			return spawnX, spawnY
		}
	}

	// Si aucune position valide trouvée après maxAttempts,
	// utiliser une position de secours aux coins
	corners := []struct{ x, y int }{
		{margin, margin},
		{maxX, margin},
		{margin, maxY},
		{maxX, maxY},
		{GRID_WIDTH / 2, GRID_HEIGHT / 2},
	}

	// Choisir le coin le plus éloigné des autres joueurs
	bestCorner := corners[0]
	bestDistance := 0.0

	for _, corner := range corners {
		minDist := math.MaxFloat64
		for _, player := range existingPlayers {
			if player == nil || !player.IsAlive {
				continue
			}
			playerCellX := int(player.X / SCALE)
			playerCellY := int(player.Y / SCALE)
			dx := corner.x - playerCellX
			dy := corner.y - playerCellY
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist < minDist {
				minDist = dist
			}
		}
		if minDist > bestDistance {
			bestDistance = minDist
			bestCorner = corner
		}
	}

	return bestCorner.x, bestCorner.y
}

func IsInsidePolygon(x, y int, trail []models.Point) bool {
	// Conversion des coordonnées de la cellule en coordonnées pixels pour la comparaison
	px := float64(x*SCALE) + 5.0
	py := float64(y*SCALE) + 5.0

	isInside := false
	j := len(trail) - 1

	for i := 0; i < len(trail); i++ {
		// Algorithme de Ray Casting (Jordan Curve Theorem)
		if ((trail[i].Y > py) != (trail[j].Y > py)) &&
			(px < (trail[j].X-trail[i].X)*(py-trail[i].Y)/(trail[j].Y-trail[i].Y)+trail[i].X) {
			isInside = !isInside
		}
		j = i
	}
	return isInside
}
