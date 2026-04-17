// Fichier : player/player.go
package player

import (
	"Jeu/engine"
	"Jeu/models"
	"math"
)

const (
	PlayerSize = 15.0
)

// CanChangeDirection vérifie si le changement de direction est autorisé (pas de demi-tour 180°)
func CanChangeDirection(current, new string) bool {
	if current == "" {
		return true // Premier mouvement, toujours autorisé
	}

	// Interdire les demi-tours 180°
	if (current == "up" && new == "down") ||
		(current == "down" && new == "up") ||
		(current == "left" && new == "right") ||
		(current == "right" && new == "left") {
		return false
	}

	return true
}

// UpdateDirection change la direction du joueur si autorisé.
// Retourne true si la direction a réellement changé.
func UpdateDirection(p *models.Player, newDir string) bool {
	if !p.IsAlive {
		return false
	}

	if p.CurrentDirection == newDir {
		return false
	}

	if CanChangeDirection(p.CurrentDirection, newDir) {
		p.CurrentDirection = newDir
		return true
	}

	return false
}

// Fichier : player/player.go

func MovePlayerInCurrentDirection(p *models.Player) {
	if !p.IsAlive || p.CurrentDirection == "" {
		return
	}

	wasOutside := p.IsOutside
	stats := models.CharacterStats[p.Character]

	// 1. Gestion de la vitesse (Boost)
	if p.SpeedMultiplier <= 0 {
		p.SpeedMultiplier = 1.0
	}
	currentSpeed := stats.BaseSpeed * p.SpeedMultiplier

	// 2. Calcul de la future position (Stockée temporairement pour vérification)
	futureX, futureY := p.X, p.Y

	switch p.CurrentDirection {
	case "up":
		futureY -= currentSpeed
	case "down":
		futureY += currentSpeed
	case "left":
		futureX -= currentSpeed
	case "right":
		futureX += currentSpeed
	}

	// 3. Vérification des limites de la carte
	if futureX < 0 {
		futureX = 0
	}
	if futureX > engine.WIDTH-PlayerSize {
		futureX = engine.WIDTH - PlayerSize
	}
	if futureY < 0 {
		futureY = 0
	}
	if futureY > engine.LENGTH-PlayerSize {
		futureY = engine.LENGTH - PlayerSize
	}

	// 4. GESTION DES OBSTACLES (Vérification avant déplacement réel)
	futureCenterX := futureX + 7.5
	futureCenterY := futureY + 7.5

	if engine.CheckObstacleCollision(futureCenterX, futureCenterY) {
		// Le joueur percute un mur : il meurt immédiatement
		engine.KillPlayer(p, "collision_mur")
		return
	}

	// 5. Application du déplacement réel
	p.X = futureX
	p.Y = futureY

	// 6. Gestion du Timer du Boost
	if p.BoostTimer > 0 {
		p.BoostTimer--
		if p.BoostTimer <= 0 {
			p.SpeedMultiplier = 1.0
			p.BoostTimer = 0
		}
	}

	// 7. Logique de traînée et de capture
	cX, cY := p.X+7.5, p.Y+7.5
	isNowOutside := engine.CheckIfOutside(cX, cY, p.PIdx)

	if isNowOutside {
		// On devient "Outside" ou on continue d'allonger la traînée
		p.IsOutside = true

		if len(p.Trail) > 0 {
			lastPoint := p.Trail[len(p.Trail)-1]
			dist := math.Sqrt(math.Pow(cX-lastPoint.X, 2) + math.Pow(cY-lastPoint.Y, 2))

			// Anti-trous (Trail Gaps) : Si le déplacement est trop rapide (> 5px),
			// on ajoute un point intermédiaire pour l'algorithme de capture.
			if dist > engine.SCALE/2 {
				p.Trail = append(p.Trail, models.Point{
					X: (cX + lastPoint.X) / 2,
					Y: (cY + lastPoint.Y) / 2,
				})
			}
		}
		p.Trail = append(p.Trail, models.Point{X: cX, Y: cY})

	} else if !isNowOutside && wasOutside {
		// Rentrée sur le territoire : On clôture et on capture
		p.IsOutside = false
		p.Trail = append(p.Trail, models.Point{X: cX, Y: cY})

		// Capture agressive (prend en compte les murs et vole les territoires)
		engine.CaptureTerritory(p)

		// Reset de la traînée après capture
		p.Trail = []models.Point{}
	}
}

// MovePlayer (ancienne fonction - conservée pour compatibilité si nécessaire)
func MovePlayer(p *models.Player, dir string) {
	_ = UpdateDirection(p, dir)
	MovePlayerInCurrentDirection(p)
}
