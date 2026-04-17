package models

import (
	"log"
	"sync"
	"time"
)

// MODIF: Point et Player sont maintenant isolés et accessibles de partout
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Player struct {
	ID               string    `json:"id"`
	PIdx             int       `json:"pIdx"`
	X                float64   `json:"x"`
	Y                float64   `json:"y"`
	Color            string    `json:"color"`
	Trail            []Point   `json:"trail"`
	IsOutside        bool      `json:"isOutside"`
	Kills            int       `json:"kills"`
	IsAlive          bool      `json:"isAlive"`
	CurrentDirection string    `json:"currentDirection"`      // "up", "down", "left", "right"
	DeathReason      string    `json:"deathReason,omitempty"` // "suicide", "killed_by_X"
	Pseudo           string    `json:"pseudo"`
	Character        string    `json:"character"`
	SpeedMultiplier  float64   `json:"speedMultiplier"`
	BoostTimer       int       `json:"-"`
	StoredBoost      string    `json:"storedBoost"`
	NextCapacityTime time.Time `json:"-"`
	Score            float64   `json:"score"`
}

type UpdateMessage struct {
	Players     map[string]*Player `json:"players"`
	Grid        []int16            `json:"grid"`
	LastEvent   *GameEvent         `json:"lastEvent,omitempty"`
	GameStarted bool               `json:"gameStarted"`
	OnlineCount int                `json:"onlineCount"`
	Victory     bool               `json:"victory"`
	Obstacles   []Obstacle         `json:"obstacles"` // test
	Boosts      []Boost            `json:"boosts"`
}

type GameEvent struct {
	Type       string `json:"type"` // "kill" or "suicide"
	KillerID   string `json:"killerId,omitempty"`
	KillerName string `json:"killerName,omitempty"`
	VictimID   string `json:"victimId"`
	VictimName string `json:"victimName"`
}

type Obstacle struct {
	ID       string  `json:"id"`
	Position Point   `json:"position"` // Coordonnées (Haut-Gauche)
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Health   int     `json:"health"` // Points de vie (pour l'Ouvrier)
}

var (
	GamePlayers  = make(map[string]*Player)
	NextIndex    = 1
	GameMutex    sync.Mutex
	GameStarted  = false
	MapObstacles []Obstacle
)

type Boost struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Type   string  `json:"type"`
	Active bool    `json:"active"`
}

func (p *Player) ApplyBoost(boostType string) {
	if p.SpeedMultiplier == 0 {
		p.SpeedMultiplier = 1.0
	}

	stats, exists := CharacterStats[p.Character]
	if !exists {
		stats = CharacterStats["chasseur"]
	}

	switch boostType {
	case "speed":
		log.Printf("Boost SPEED activé pour %s", p.Pseudo)
		p.SpeedMultiplier = stats.BoostMultiplier
		p.BoostTimer = stats.BoostDuration
		p.StoredBoost = ""
	}
}

var CharacterStats = map[string]struct {
	BaseSpeed        float64
	BoostMultiplier  float64
	BoostDuration    int
	CapacityCooldown int
	ExpansionPercent int
	SlowCapa         float64
}{
	"chasseur": {BaseSpeed: 3.5, BoostMultiplier: 2.0, BoostDuration: 150, CapacityCooldown: 5, ExpansionPercent: 0, SlowCapa: 0.7},
	"ouvrier":  {BaseSpeed: 2.5, BoostMultiplier: 1.5, BoostDuration: 200, CapacityCooldown: 5, ExpansionPercent: 0, SlowCapa: 1},
	"zombie":   {BaseSpeed: 3.0, BoostMultiplier: 1.8, BoostDuration: 120, CapacityCooldown: 5, ExpansionPercent: 5, SlowCapa: 1},
}
