package network

import (
	"Jeu/engine"
	"Jeu/models"
	"Jeu/player"
	"encoding/json"
	"log"
	"math"
	"time"
)

// MODIF: Déclaration du Hub dans le package network
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan interface{}
	Register   chan *Client
	Unregister chan *Client
}

var activeBoosts = []models.Boost{}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan interface{}),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	// Lancer la game loop dans une goroutine séparée
	go h.GameLoop()
	go h.BoostSpawner()

	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			jsonMsg, err := json.Marshal(message)
			if err != nil {
				log.Printf("Erreur JSON: %v", err)
				continue
			}
			for client := range h.Clients {
				select {
				case client.Send <- jsonMsg:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

// GameLoop déplace tous les joueurs automatiquement et détecte les collisions
func (h *Hub) GameLoop() {
	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()
	gridCounter := 0

	for range ticker.C {
		models.GameMutex.Lock()

		var gameEvent *models.GameEvent
		var gridToSend []int16

		if models.GameStarted {
			h.checkVictory()

			for _, p := range models.GamePlayers {
				if !p.IsAlive || p.CurrentDirection == "" {
					continue
				}

				player.MovePlayerInCurrentDirection(p)

				for i := len(activeBoosts) - 1; i >= 0; i-- {
					b := activeBoosts[i]
					dist := math.Sqrt(math.Pow(p.X-b.X, 2) + math.Pow(p.Y-b.Y, 2))
					if dist < 15 && p.StoredBoost == "" {
						p.StoredBoost = b.Type
						activeBoosts = append(activeBoosts[:i], activeBoosts[i+1:]...)
					}
				}

				if engine.CheckSelfCollision(p) {
					engine.KillPlayer(p, "suicide")
					gameEvent = &models.GameEvent{Type: "suicide", VictimID: p.ID, VictimName: p.Color}
					continue
				}

				victim := engine.CheckTrailCollisionWithOthers(p, models.GamePlayers)
				if victim != nil {
					engine.KillPlayer(victim, "killed_by_"+p.Color)
					p.Kills++
					gameEvent = &models.GameEvent{
						Type: "kill", KillerID: p.ID, KillerName: p.Color,
						VictimID: victim.ID, VictimName: victim.Color,
					}
				}
			}
			engine.RemoveDeadPlayers(models.GamePlayers)
		}

		if gridCounter%10 == 0 {
			gridToSend = engine.Grid

			scores := engine.CalculateTerritoryScores()
			for _, p := range models.GamePlayers {
				if p.IsAlive {
					p.Score = scores[int16(p.PIdx)]
				} else {
					p.Score = 0
				}
			}
		}
		gridCounter++

		// 3. BROADCAST
		update := models.UpdateMessage{
			Players:     models.GamePlayers,
			Grid:        gridToSend,
			Obstacles:   models.MapObstacles,
			LastEvent:   gameEvent,
			OnlineCount: h.GetConnectedCount(),
			GameStarted: models.GameStarted,
			Boosts:      activeBoosts,
		}

		models.GameMutex.Unlock()
		h.Broadcast <- update
	}
}

func (h *Hub) GetConnectedCount() int {
	count := 0
	for _, p := range models.GamePlayers {
		if p.Pseudo != "" {
			count++
		}
	}
	return count
}

func (h *Hub) checkVictory() {
	alivePlayers := []*models.Player{}

	for _, p := range models.GamePlayers {
		if p.IsAlive {
			alivePlayers = append(alivePlayers, p)
		}
	}

	if (len(alivePlayers) == 1) && models.GameStarted {
		models.GameStarted = false

		winner := alivePlayers[0]
		log.Printf("Victoire du joueur %s (%s) !", winner.Pseudo, winner.Color)

		update := models.UpdateMessage{
			Players:     models.GamePlayers,
			Grid:        engine.Grid,
			Victory:     true,
			GameStarted: false,
		}
		h.Broadcast <- update

		go func() {
			time.Sleep(5 * time.Second)
			h.resetGame()
		}()

	} else if len(alivePlayers) == 0 && models.GameStarted {
		models.GameStarted = false

		log.Printf("Victoire d'aucun joueur c'est une égalité !")

		update := models.UpdateMessage{
			Players:     models.GamePlayers,
			Grid:        engine.Grid,
			Victory:     true,
			GameStarted: false,
		}
		h.Broadcast <- update

		go func() {
			time.Sleep(5 * time.Second)
			h.resetGame()
		}()
	}
}

func (h *Hub) resetGame() {
	models.GameMutex.Lock()
	defer models.GameMutex.Unlock()
	engine.LoadSymmetricMap()

	engine.ResetWorld()
	activeBoosts = []models.Boost{}
	for _, p := range models.GamePlayers {
		p.IsAlive = true
		p.Trail = []models.Point{}
		p.CurrentDirection = ""
		p.SpeedMultiplier = 1.0
		p.BoostTimer = 0
		newX, newY := engine.FindValidSpawnPosition(models.GamePlayers)
		p.X = float64(newX * engine.SCALE)
		p.Y = float64(newY * engine.SCALE)
	}

	log.Println("Le monde a été réinitialisé. Prêt pour une nouvelle partie.")
}

func (h *Hub) BoostSpawner() {
	spawnTicker := time.NewTicker(10 * time.Second)
	for range spawnTicker.C {
		if !models.GameStarted {
			continue
		}

		models.GameMutex.Lock()
		bx, by := engine.FindValidSpawnPosition(models.GamePlayers)

		newBoost := models.Boost{
			X:      float64(bx * engine.SCALE),
			Y:      float64(by * engine.SCALE),
			Type:   "speed",
			Active: true,
		}

		activeBoosts = append(activeBoosts, newBoost)
		models.GameMutex.Unlock()

		log.Println("Boost généré en :", bx, by)
	}
}
