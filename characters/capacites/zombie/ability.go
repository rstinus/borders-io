package zombie

import (
	"Jeu/engine"
	"Jeu/models"
	"math/rand"
)

func ExecuteZombieAttack(p *models.Player) bool {
	pIdx := int16(p.PIdx)

	// 1. Trouver toutes les cases du joueur
	playerCells := []int{}
	for i, val := range engine.Grid {
		if val == pIdx {
			playerCells = append(playerCells, i)
		}
	}

	if len(playerCells) == 0 {
		return false
	}

	// 2. Calculer le nombre de cases à infecter (X%)
	stats := models.CharacterStats["zombie"]
	targetCount := (len(playerCells) * stats.ExpansionPercent) / 100
	if targetCount < 5 {
		targetCount = 5
	} // Minimum de 5 cases

	// 3. Trouver les "foyers d'infection" (cases du joueur ayant un voisin libre)
	borders := []int{}
	dirs := []int{1, -1, engine.GRID_WIDTH, -engine.GRID_WIDTH}

	for _, cellIdx := range playerCells {
		isBorder := false
		for _, d := range dirs {
			neighbor := cellIdx + d
			if neighbor >= 0 && neighbor < len(engine.Grid) && engine.Grid[neighbor] != pIdx && engine.ObstaclesGrid[neighbor] == 0 {
				isBorder = true
				break
			}
		}
		if isBorder {
			borders = append(borders, cellIdx)
		}
	}

	if len(borders) == 0 {
		return false
	}

	// 4. Choisir UN foyer aléatoire pour démarrer l'infection
	startCell := borders[rand.Intn(len(borders))]

	// 5. Propagation (BFS) pour atteindre le quota
	queue := []int{startCell}
	infected := 0
	visited := make(map[int]bool)
	visited[startCell] = true

	for len(queue) > 0 && infected < targetCount {
		curr := queue[0]
		queue = queue[1:]

		// Mélanger les directions pour une propagation organique
		rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })

		for _, d := range dirs {
			neighbor := curr + d
			if neighbor >= 0 && neighbor < len(engine.Grid) && !visited[neighbor] {
				// Si la case n'est pas déjà au joueur et n'est pas un mur
				if engine.Grid[neighbor] != pIdx && engine.ObstaclesGrid[neighbor] == 0 {
					// CONTAMINATION !
					engine.Grid[neighbor] = pIdx
					visited[neighbor] = true
					queue = append(queue, neighbor)
					infected++

					if infected >= targetCount {
						break
					}
				}
			}
		}
	}

	return infected > 0
}
