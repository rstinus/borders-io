// Fichier : network/client.go
package network

import (
	"Jeu/characters/capacites/chasseur"
	"Jeu/characters/capacites/ouvrier"
	"Jeu/characters/capacites/zombie"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"Jeu/engine"
	"Jeu/models"
	"Jeu/player"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
	ID   string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// On utilise RemoteAddr comme ID temporaire
	client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256), ID: r.RemoteAddr}
	client.Hub.Register <- client

	go client.writePump()
	go client.readPump()
}

func eliminatePlayer(p *models.Player) {
	engine.RemovePlayerFromGrid(p.PIdx)
	p.IsAlive = false
	p.Trail = []models.Point{}
	p.X = -1000
	p.Y = -1000
}

func (c *Client) readPump() {
	defer func() {
		// MODIF: Nettoyage impératif du joueur lors de la déconnexion physique
		models.GameMutex.Lock()
		if p, ok := models.GamePlayers[c.ID]; ok {
			engine.RemovePlayerFromGrid(p.PIdx) // On efface son territoire
			delete(models.GamePlayers, c.ID)    // On le supprime de la liste
		}
		models.GameMutex.Unlock()

		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		shouldBroadcast := false
		var msg map[string]interface{}
		json.Unmarshal(message, &msg)

		msgType, _ := msg["type"].(string)

		if msgType == "join_game" {
			handleJoin(c, msg)
			shouldBroadcast = true
		} else if msgType == "move" {
			dir, _ := msg["direction"].(string)
			shouldBroadcast = handleMove(c, dir)
		} else if msgType == "start_game" {
			models.GameMutex.Lock()
			models.GameStarted = true
			models.GameMutex.Unlock()
			shouldBroadcast = true
		} else if msgType == "use_ability" {
			ability, _ := msg["ability"].(string)

			models.GameMutex.Lock()
			p := models.GamePlayers[c.ID]

			if p != nil && p.IsAlive {
				now := time.Now()

				// --- 1. CAPACITÉ BOOST (Toujours prioritaire, pas de cooldown) ---
				if ability == "boost" && p.StoredBoost == "speed" {
					p.ApplyBoost("speed")
					shouldBroadcast = true
					log.Printf("Le joueur %s a activé son boost stocké", p.Pseudo)
				}

				// --- 2. CAPACITÉS ATTACK (Soumises au cooldown) ---
				if ability == "attack" && now.After(p.NextCapacityTime) {
					success := false

					// Attaque OUVRIER : Destruction de murs
					if p.Character == "ouvrier" {
						if ouvrier.ExecuteInstantAttack(p) {
							success = true
							log.Printf("L'ouvrier %s a utilisé son attaque !", p.Pseudo)
						}
					}

					// Attaque CHASSEUR : (Prêt pour ta future logique)
					if p.Character == "chasseur" {
						if chasseur.ExecuteChasseurAttack(p) {
							success = true
							log.Printf("Le chasseur %s a utilisé son attaque !", p.Pseudo)
						}

					}

					// Attaque ZOMBIE : (Prêt pour ta future logique)
					if p.Character == "zombie" {
						if zombie.ExecuteZombieAttack(p) { // <--- Appel de ta nouvelle fonction
							success = true
							log.Printf("Le zombie %s a infecté de nouvelles terres !", p.Pseudo)
						}
					}

					// Si une des attaques a réussi, on lance le cooldown
					if success {
						stats := models.CharacterStats[p.Character]
						p.NextCapacityTime = now.Add(time.Duration(stats.CapacityCooldown) * time.Second)
						shouldBroadcast = true
					}
				} else if ability == "attack" {
					log.Printf("Capacité d'attaque en recharge pour %s", p.Pseudo)
				}
			}
			models.GameMutex.Unlock()
		}

		if shouldBroadcast {
			update := models.UpdateMessage{
				Players:     models.GamePlayers,
				Grid:        engine.Grid,
				GameStarted: models.GameStarted,
				OnlineCount: len(c.Hub.Clients),
			}
			c.Hub.Broadcast <- update
		}
	}
}

func (c *Client) writePump() {
	for message := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, message)
	}
}

func handleJoin(c *Client, msg map[string]interface{}) {
	models.GameMutex.Lock()
	defer models.GameMutex.Unlock()

	// MODIF: On autorise le join si la partie n'est pas lancée
	// OU si le joueur est déjà dans la liste (reconnexion)
	if models.GameStarted {
		log.Printf("Join refusé: partie déjà en cours")
		return
	}

	color, _ := msg["color"].(string)
	pseudo, _ := msg["pseudo"].(string)
	character, _ := msg["character"].(string)

	if pseudo == "" {
		pseudo = "Anonyme"
	}

	stats, exists := models.CharacterStats[character]
	if !exists {
		character = "chasseur"
		stats = models.CharacterStats["chasseur"]
	}

	if _, exists := models.GamePlayers[c.ID]; exists {
		delete(models.GamePlayers, c.ID)
	}

	pIdx := models.NextIndex
	models.NextIndex++

	startX, startY := engine.FindValidSpawnPosition(models.GamePlayers)

	p := &models.Player{
		ID:               c.ID,
		PIdx:             pIdx,
		X:                float64(startX * engine.SCALE),
		Y:                float64(startY * engine.SCALE),
		Color:            color,
		Trail:            []models.Point{},
		Kills:            0,
		IsAlive:          true,
		CurrentDirection: "",
		Pseudo:           pseudo,
		Character:        character,
		SpeedMultiplier:  1.0,
		BoostTimer:       0,
		NextCapacityTime: time.Now(),
	}
	models.GamePlayers[c.ID] = p

	// Territoire de départ
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {
			gx, gy := startX+j, startY+i
			if gx >= 0 && gx < engine.GRID_WIDTH && gy >= 0 && gy < engine.GRID_HEIGHT {
				engine.Grid[gy*engine.GRID_WIDTH+gx] = int16(pIdx)
			}
		}
	}
	log.Printf("Joueur %s (%s) a rejoint. Vitesse de base: %.1f", pseudo, character, stats.BaseSpeed)
}

func handleMove(c *Client, dir string) bool {
	models.GameMutex.Lock()
	defer models.GameMutex.Unlock()

	if !models.GameStarted {
		return false
	}

	p := models.GamePlayers[c.ID]
	if p == nil || !p.IsAlive {
		return false
	}

	return player.UpdateDirection(p, dir)
}

func handleAbility(c *Client, ability string) {
	models.GameMutex.Lock()
	defer models.GameMutex.Unlock()

	p := models.GamePlayers[c.ID]
	if p == nil || !p.IsAlive {
		return
	}

	// --- ÉTAPE 1 : VÉRIFICATION DU COOLDOWN ---
	now := time.Now()
	if now.Before(p.NextCapacityTime) {
		log.Printf("Capacité en recharge pour %s (attendre encore %.1fs)",
			p.Pseudo, p.NextCapacityTime.Sub(now).Seconds())
		return // ON ARRÊTE TOUT ICI : la capacité ne se fait pas
	}

	// --- ÉTAPE 2 : EXÉCUTION SELON LE PERSONNAGE ---
	success := false

	if p.Character == "ouvrier" && ability == "attack" {
		// Ton attaque de zone de 15px
		if ouvrier.ExecuteInstantAttack(p) {
			success = true
		}
	} else if ability == "boost" && p.StoredBoost == "speed" {
		p.ApplyBoost("speed")
		success = true
	}

	// --- ÉTAPE 3 : DÉCLENCHEMENT DU COOLDOWN ---
	if success {
		// On récupère la durée depuis tes CharacterStats
		stats := models.CharacterStats[p.Character]

		// On définit le prochain moment autorisé
		p.NextCapacityTime = now.Add(time.Duration(stats.CapacityCooldown) * time.Second)
	}
}
