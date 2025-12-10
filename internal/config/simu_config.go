package config

import "time"

// Defini les paramètres de la simulation

// États du jeu
type GameState int

const (
	StateMenu GameState = iota // la ou on va choisir les paramètres de la simulation ( nbr de voitures/pietons par sec, type de carrefour, etc )
	StateSimulation
)

// Configuration de la simulation
type SimulationConfig struct {
	CarSpawnRate        time.Duration
	PedestrianSpawnRate time.Duration
	//TrafficLightDuration time.Duration // à implemnter plus tard
	CrossingType string // "shibuya" ou "french"
}

// Retourne une configuration par défaut
func DefaultConfig() SimulationConfig {
	return SimulationConfig{
		CarSpawnRate:        1 * time.Second,
		PedestrianSpawnRate: 2 * time.Second,
		//TrafficLightDuration: 10 * time.Second,
		CrossingType: "",
	}
}
