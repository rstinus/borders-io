// Fichier : engine/obstacles.go
package engine

import (
	"Jeu/models"
)

// ObstaclesGrid est la grille parallèle (0 = vide, >0 = points de vie du mur)
var ObstaclesGrid = make([]int16, GRID_WIDTH*GRID_HEIGHT)

// LoadSymmetricMap initialise la carte avec les 4 bastions et le centre
func LoadSymmetricMap() {
	bastionSize := float64(200) // 20 cases * SCALE (10)
	health := 100

	// 1. Définition géométrique des obstacles (en pixels)
	models.MapObstacles = []models.Obstacle{
		{ID: "bastion_hg", Position: models.Point{X: 300, Y: 200}, Width: bastionSize, Height: bastionSize, Health: health},
		{ID: "bastion_hd", Position: models.Point{X: WIDTH - 300 - bastionSize, Y: 200}, Width: bastionSize, Height: bastionSize, Health: health},
		{ID: "bastion_bg", Position: models.Point{X: 300, Y: LENGTH - 200 - bastionSize}, Width: bastionSize, Height: bastionSize, Health: health},
		{ID: "bastion_bd", Position: models.Point{X: WIDTH - 300 - bastionSize, Y: LENGTH - 200 - bastionSize}, Width: bastionSize, Height: bastionSize, Health: health},
		{ID: "mur_central", Position: models.Point{X: WIDTH/2 - 50, Y: LENGTH/2 - 150}, Width: 100, Height: 300, Health: health},
	}

	// 2. Remplissage de la grille (ObstaclesGrid) pour des calculs ultra-rapides
	for _, obs := range models.MapObstacles {
		// Convertir les pixels en index de grille
		startX := int(obs.Position.X / SCALE)
		startY := int(obs.Position.Y / SCALE)
		endX := int((obs.Position.X + obs.Width) / SCALE)
		endY := int((obs.Position.Y + obs.Height) / SCALE)

		for y := startY; y < endY; y++ {
			for x := startX; x < endX; x++ {
				if x >= 0 && x < GRID_WIDTH && y >= 0 && y < GRID_HEIGHT {
					ObstaclesGrid[y*GRID_WIDTH+x] = int16(obs.Health)
				}
			}
		}
	}
}

// CheckObstacleCollision vérifie si un point en pixels touche un mur
func CheckObstacleCollision(px, py float64) bool {
	gx := int(px / SCALE)
	gy := int(py / SCALE)

	// Si en dehors de la carte, on considère que c'est un mur (sécurité)
	if gx < 0 || gx >= GRID_WIDTH || gy < 0 || gy >= GRID_HEIGHT {
		return true
	}

	// Si la valeur est > 0, c'est un mur
	return ObstaclesGrid[gy*GRID_WIDTH+gx] > 0
}

// RemoveObstacleAt identifie l'obstacle contenant l'index idx,
// le supprime de la grille et met ses PV à 0.
func RemoveObstacleAt(idx int) {
	if idx < 0 || idx >= len(ObstaclesGrid) {
		return
	}
	// Calculer les coordonnées de la case
	gx := idx % GRID_WIDTH
	gy := idx / GRID_WIDTH
	px, py := float64(gx*SCALE), float64(gy*SCALE)

	// Chercher l'obstacle structurel correspondant
	for i := range models.MapObstacles {
		obs := &models.MapObstacles[i]
		if px >= obs.Position.X && px < obs.Position.X+obs.Width &&
			py >= obs.Position.Y && py < obs.Position.Y+obs.Height {

			// 1. Vider toutes les cases de cet obstacle dans la grille physique
			startX := int(obs.Position.X / SCALE)
			startY := int(obs.Position.Y / SCALE)
			endX := int((obs.Position.X + obs.Width) / SCALE)
			endY := int((obs.Position.Y + obs.Height) / SCALE)

			for y := startY; y < endY; y++ {
				for x := startX; x < endX; x++ {
					ObstaclesGrid[y*GRID_WIDTH+x] = 0
				}
			}
			// 2. Mettre les PV à 0 pour que le client arrête de le dessiner
			obs.Health = 0
			return
		}
	}
}
