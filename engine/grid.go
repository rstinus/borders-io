// Fichier : engine/grid.go
package engine

import (
	"Jeu/models"
)

const (
	SCALE       = 10
	WIDTH       = 1920
	LENGTH      = 1080
	GRID_WIDTH  = WIDTH / SCALE
	GRID_HEIGHT = LENGTH / SCALE
)

var Grid = make([]int16, GRID_WIDTH*GRID_HEIGHT)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CheckIfOutside(px, py float64, pIdx int) bool {
	gx := int(px / SCALE)
	gy := int(py / SCALE)
	if gx < 0 || gx >= GRID_WIDTH || gy < 0 || gy >= GRID_HEIGHT {
		return true
	}
	return Grid[gy*GRID_WIDTH+gx] != int16(pIdx)
}

func CaptureTerritory(p *models.Player) {
	if len(p.Trail) < 3 {
		return
	}

	// 1. Calculer la Bounding Box de la traînée
	minX, maxX, minY, maxY := GRID_WIDTH, 0, GRID_HEIGHT, 0
	for _, pt := range p.Trail {
		gx, gy := int(pt.X/SCALE), int(pt.Y/SCALE)
		if gx < minX {
			minX = gx
		}
		if gx > maxX {
			maxX = gx
		}
		if gy < minY {
			minY = gy
		}
		if gy > maxY {
			maxY = gy
		}
	}

	// 2. Transformer la traînée en "Mur Physique" sur la grille
	for _, pt := range p.Trail {
		gx, gy := int(pt.X/SCALE), int(pt.Y/SCALE)
		if gx >= 0 && gx < GRID_WIDTH && gy >= 0 && gy < GRID_HEIGHT {
			idx := gy*GRID_WIDTH + gx
			// On ne peint pas sur les obstacles
			if ObstaclesGrid[idx] == 0 {
				if Grid[idx] != int16(p.PIdx) {
					Grid[idx] = int16(p.PIdx)
					checkHeadCollision(gx, gy, p.ID) // Si la traînée écrase un joueur
				}
			}
		}
	}

	// 3. Remplir intelligemment tout ce qui est emprisonné (Flood Fill inversé)
	fillTrappedAreas(p, minX, maxX, minY, maxY)

	checkSurvival()
}

// fillTrappedAreas détecte et remplit les zones enfermées par la traînée, le territoire et les obstacles.
func fillTrappedAreas(p *models.Player, minX, maxX, minY, maxY int) {
	// On élargit d'une case pour créer une bordure "Extérieure" garantie (sans bloquer les bords de la map)
	minX -= 1
	minY -= 1
	maxX += 1
	maxY += 1

	targetID := int16(p.PIdx)
	outside := make(map[int]bool)
	queue := []struct{ x, y int }{}

	// ÉTAPE A : Injecter le fluide "Extérieur" sur tout le contour de la Bounding Box
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if x == minX || x == maxX || y == minY || y == maxY {
				// Vérifier qu'on est bien sur la carte
				if x >= 0 && x < GRID_WIDTH && y >= 0 && y < GRID_HEIGHT {
					idx := y*GRID_WIDTH + x
					// Si la case frontière n'est ni à nous, ni un mur, le fluide peut y exister
					if Grid[idx] != targetID && ObstaclesGrid[idx] == 0 {
						outside[idx] = true
						queue = append(queue, struct{ x, y int }{x, y})
					}
				}
			}
		}
	}

	// ÉTAPE B : Propager le fluide Extérieur
	dirs := []struct{ dx, dy int }{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, d := range dirs {
			nx, ny := curr.x+d.dx, curr.y+d.dy

			// Rester dans la Bounding Box
			if nx >= minX && nx <= maxX && ny >= minY && ny <= maxY {
				// Rester sur la grille
				if nx >= 0 && nx < GRID_WIDTH && ny >= 0 && ny < GRID_HEIGHT {
					idx := ny*GRID_WIDTH + nx

					// Si la case n'a pas été inondée, n'est pas à nous, et n'est pas un obstacle... on l'inonde !
					if !outside[idx] && Grid[idx] != targetID && ObstaclesGrid[idx] == 0 {
						outside[idx] = true
						queue = append(queue, struct{ x, y int }{nx, ny})
					}
				}
			}
		}
	}

	// ÉTAPE C : Toute case qui n'a pas été inondée est emprisonnée -> On la capture !
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if x >= 0 && x < GRID_WIDTH && y >= 0 && y < GRID_HEIGHT {
				idx := y*GRID_WIDTH + x

				if Grid[idx] != targetID && ObstaclesGrid[idx] == 0 {
					// Le miracle opère ici : si ce n'est pas outside, c'est que c'est piégé !
					if !outside[idx] {
						Grid[idx] = targetID
						checkHeadCollision(x, y, p.ID)
					}
				}
			}
		}
	}
}

func checkHeadCollision(gx, gy int, attackerID string) {
	for _, victim := range models.GamePlayers {
		if victim.ID == attackerID || !victim.IsAlive {
			continue
		}
		vGx, vGy := int(victim.X/SCALE), int(victim.Y/SCALE)
		if vGx == gx && vGy == gy {
			KillPlayer(victim, "swallowed")
		}
	}
}

func checkSurvival() {
	activeIndices := make(map[int16]bool)
	for _, val := range Grid {
		if val > 0 {
			activeIndices[val] = true
		}
	}

	for _, player := range models.GamePlayers {
		if player.IsAlive && !activeIndices[int16(player.PIdx)] {
			KillPlayer(player, "territory_lost")
		}
	}
}

func RemovePlayerFromGrid(pIdx int) {
	idToRemove := int16(pIdx)
	for i := 0; i < len(Grid); i++ {
		if Grid[i] == idToRemove {
			Grid[i] = 0
		}
	}
}

func ResetWorld() {
	for i := range Grid {
		Grid[i] = 0
	}
}

func CalculateTerritoryScores() map[int16]float64 {
	totalCapturable := 0
	for _, obsHP := range ObstaclesGrid {
		if obsHP == 0 {
			totalCapturable++
		}
	}

	if totalCapturable == 0 {
		return make(map[int16]float64)
	}

	counts := make(map[int16]int)
	for _, ownerIdx := range Grid {
		if ownerIdx > 0 {
			counts[ownerIdx]++
		}
	}

	scores := make(map[int16]float64)
	for pIdx, count := range counts {
		scores[pIdx] = (float64(count) / float64(totalCapturable)) * 100
	}
	return scores
}
