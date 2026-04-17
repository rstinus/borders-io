package chasseur

import (
	"Jeu/models"
	"log"
)

// ExecuteChasseurAttack applique le ralentissement spécifique au Chasseur
func ExecuteChasseurAttack(p *models.Player) bool {
	// On vérifie que le joueur est bien un chasseur
	if p.Character != "chasseur" {
		return false
	}

	// On récupère les statistiques définies dans models
	stats, exists := models.CharacterStats["chasseur"]
	if !exists {
		return false
	}

	// On applique le multiplicateur de ralentissement (0.7)
	p.SpeedMultiplier = stats.SlowCapa

	// On définit la durée de cet effet (en ticks)
	// Ici on utilise BoostDuration (150) défini dans vos stats
	p.BoostTimer = stats.BoostDuration

	log.Printf("[CAPA] Le chasseur %s entre en mode concentration (vitesse réduite à %.1f)", p.Pseudo, p.SpeedMultiplier)

	return true
}