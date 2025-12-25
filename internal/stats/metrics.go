package stats

// fichier qui gere la collecte et le stockage des métriques de la simulation
/*
 Il contient :
   - Structures de données pour les métriques des agents et de la simulation
   -> Enregistrement des spawns/destructions d'agents
   - Calcul des statistiques finales (moyennes, min, max)
   - Sauvegarde des métriques dans un fichier JSON avec nom automatique ( dans le dossier metrics_data/ )
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
)

// AgentMetrics contient les métriques pour un agent individuel
type AgentMetrics struct {
	AgentID      uint
	AgentType    string // "car" ou "pedestrian"
	SpawnTime    time.Time
	DestroyTime  time.Time
	LifeDuration float64 // en secondes
	Direction    string
	MaxSpeed     float64 // Vitesse maximale atteinte (surtout pour piétons)
}

// SimulationMetrics contient toutes les métriques de la simulation
type SimulationMetrics struct {
	mu sync.Mutex

	// Configuration de la simulation
	CrossingType        string
	CarSpawnRate        float64 // voitures par seconde
	PedestrianSpawnRate float64 // piétons par seconde
	StartTime           time.Time
	EndTime             time.Time

	// Métriques des agents
	CarMetrics        []AgentMetrics
	PedestrianMetrics []AgentMetrics

	// Statistiques globales (calculées à la fin)
	AvgCarLifetime    float64
	AvgPedestrianLife float64
	TotalCars         int
	TotalPedestrians  int
	MinCarLifetime    float64
	MaxCarLifetime    float64
	MinPedestrianLife float64
	MaxPedestrianLife float64

	// Vitesse maximale des piétons
	MaxPedestrianSpeed   float64
	MaxPedestrianSpeedID uint

	// Métriques par direction
	CarsByDirection        map[string]int
	PedestriansByDirection map[string]int
}

// Global metrics collector
var GlobalMetrics *SimulationMetrics

// InitMetrics initialise le collecteur de métriques
func InitMetrics(cfg config.SimulationConfig) {
	GlobalMetrics = &SimulationMetrics{
		CrossingType:           cfg.CrossingType,
		CarSpawnRate:           1.0 / cfg.CarSpawnRate.Seconds(),
		PedestrianSpawnRate:    1.0 / cfg.PedestrianSpawnRate.Seconds(),
		StartTime:              time.Now(),
		CarMetrics:             make([]AgentMetrics, 0),
		PedestrianMetrics:      make([]AgentMetrics, 0),
		CarsByDirection:        make(map[string]int),
		PedestriansByDirection: make(map[string]int),
	}
}

// RecordCarSpawn enregistre l'apparition d'une voiture
func RecordCarSpawn(carID uint, direction string) {
	if GlobalMetrics == nil {
		return
	}
	GlobalMetrics.mu.Lock()
	defer GlobalMetrics.mu.Unlock()

	GlobalMetrics.CarMetrics = append(GlobalMetrics.CarMetrics, AgentMetrics{
		AgentID:   carID,
		AgentType: "car",
		SpawnTime: time.Now(),
		Direction: direction,
	})
	GlobalMetrics.TotalCars++
	GlobalMetrics.CarsByDirection[direction]++
}

// RecordPedestrianSpawn enregistre l'apparition d'un piéton
func RecordPedestrianSpawn(pedID uint, direction string) {
	if GlobalMetrics == nil {
		return
	}
	GlobalMetrics.mu.Lock()
	defer GlobalMetrics.mu.Unlock()

	GlobalMetrics.PedestrianMetrics = append(GlobalMetrics.PedestrianMetrics, AgentMetrics{
		AgentID:   pedID,
		AgentType: "pedestrian",
		SpawnTime: time.Now(),
		Direction: direction,
	})
	GlobalMetrics.TotalPedestrians++
	GlobalMetrics.PedestriansByDirection[direction]++
}

// RecordCarDestroy enregistre la destruction d'une voiture
func RecordCarDestroy(carID uint) {
	if GlobalMetrics == nil {
		return
	}
	GlobalMetrics.mu.Lock()
	defer GlobalMetrics.mu.Unlock()

	// Trouver la voiture correspondante
	for i := range GlobalMetrics.CarMetrics {
		if GlobalMetrics.CarMetrics[i].AgentID == carID {
			GlobalMetrics.CarMetrics[i].DestroyTime = time.Now()
			GlobalMetrics.CarMetrics[i].LifeDuration =
				GlobalMetrics.CarMetrics[i].DestroyTime.Sub(GlobalMetrics.CarMetrics[i].SpawnTime).Seconds()
			break
		}
	}
}

// RecordPedestrianDestroy enregistre la destruction d'un piéton
func RecordPedestrianDestroy(pedID uint) {
	if GlobalMetrics == nil {
		return
	}
	GlobalMetrics.mu.Lock()
	defer GlobalMetrics.mu.Unlock()

	// Trouver le piéton correspondant
	for i := range GlobalMetrics.PedestrianMetrics {
		if GlobalMetrics.PedestrianMetrics[i].AgentID == pedID {
			GlobalMetrics.PedestrianMetrics[i].DestroyTime = time.Now()
			GlobalMetrics.PedestrianMetrics[i].LifeDuration =
				GlobalMetrics.PedestrianMetrics[i].DestroyTime.Sub(GlobalMetrics.PedestrianMetrics[i].SpawnTime).Seconds()
			break
		}
	}
}

// RecordPedestrianSpeed enregistre la vitesse d'un piéton et met à jour le max si nécessaire
func RecordPedestrianSpeed(pedID uint, speed float64) {
	if GlobalMetrics == nil {
		return
	}
	GlobalMetrics.mu.Lock()
	defer GlobalMetrics.mu.Unlock()

	// Trouver le piéton et mettre à jour sa vitesse max
	for i := range GlobalMetrics.PedestrianMetrics {
		if GlobalMetrics.PedestrianMetrics[i].AgentID == pedID {
			if speed > GlobalMetrics.PedestrianMetrics[i].MaxSpeed {
				GlobalMetrics.PedestrianMetrics[i].MaxSpeed = speed
			}
			break
		}
	}

	// Mettre à jour le record global
	if speed > GlobalMetrics.MaxPedestrianSpeed {
		GlobalMetrics.MaxPedestrianSpeed = speed
		GlobalMetrics.MaxPedestrianSpeedID = pedID
	}
}

// CalculateStatistics calcule les statistiques finales
func (sm *SimulationMetrics) CalculateStatistics() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.EndTime = time.Now()

	// Calculer les statistiques des voitures
	var totalCarLife float64
	sm.MinCarLifetime = 999999.0
	sm.MaxCarLifetime = 0.0
	validCars := 0

	for _, car := range sm.CarMetrics {
		if car.LifeDuration > 0 {
			totalCarLife += car.LifeDuration
			validCars++
			if car.LifeDuration < sm.MinCarLifetime {
				sm.MinCarLifetime = car.LifeDuration
			}
			if car.LifeDuration > sm.MaxCarLifetime {
				sm.MaxCarLifetime = car.LifeDuration
			}
		}
	}

	if validCars > 0 {
		sm.AvgCarLifetime = totalCarLife / float64(validCars)
	}

	// Calculer les statistiques des piétons
	var totalPedLife float64
	sm.MinPedestrianLife = 999999.0
	sm.MaxPedestrianLife = 0.0
	validPeds := 0

	for _, ped := range sm.PedestrianMetrics {
		if ped.LifeDuration > 0 {
			totalPedLife += ped.LifeDuration
			validPeds++
			if ped.LifeDuration < sm.MinPedestrianLife {
				sm.MinPedestrianLife = ped.LifeDuration
			}
			if ped.LifeDuration > sm.MaxPedestrianLife {
				sm.MaxPedestrianLife = ped.LifeDuration
			}
		}
	}

	if validPeds > 0 {
		sm.AvgPedestrianLife = totalPedLife / float64(validPeds)
	}
}

// GetCurrentStats retourne les statistiques actuelles (pour affichage en temps réel)
func (sm *SimulationMetrics) GetCurrentStats() map[string]interface{} {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Calculer les moyennes actuelles
	var totalCarLife float64
	validCars := 0
	for _, car := range sm.CarMetrics {
		if car.LifeDuration > 0 {
			totalCarLife += car.LifeDuration
			validCars++
		}
	}

	var totalPedLife float64
	validPeds := 0
	for _, ped := range sm.PedestrianMetrics {
		if ped.LifeDuration > 0 {
			totalPedLife += ped.LifeDuration
			validPeds++
		}
	}

	avgCarLife := 0.0
	if validCars > 0 {
		avgCarLife = totalCarLife / float64(validCars)
	}

	avgPedLife := 0.0
	if validPeds > 0 {
		avgPedLife = totalPedLife / float64(validPeds)
	}

	return map[string]interface{}{
		"total_cars":            sm.TotalCars,
		"total_pedestrians":     sm.TotalPedestrians,
		"completed_cars":        validCars,
		"completed_pedestrians": validPeds,
		"avg_car_lifetime":      avgCarLife,
		"avg_ped_lifetime":      avgPedLife,
		"simulation_duration":   time.Since(sm.StartTime).Seconds(),
		"max_ped_speed":         sm.MaxPedestrianSpeed,
		"max_ped_speed_id":      sm.MaxPedestrianSpeedID,
	}
}

// SaveToFile sauvegarde les métriques dans un fichier JSON
func (sm *SimulationMetrics) SaveToFile() error {
	sm.CalculateStatistics()

	// Créer le nom de fichier avec timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("metrics_%s_cars%.1f_peds%.1f_%s.json",
		sm.CrossingType,
		sm.CarSpawnRate,
		sm.PedestrianSpawnRate,
		timestamp,
	)

	// Créer le dossier metrics s'il n'existe pas
	metricsDir := "metrics_data"
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return fmt.Errorf("erreur création dossier metrics: %v", err)
	}

	filepath := filepath.Join(metricsDir, filename)

	// Sérialiser en JSON
	data, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur sérialisation JSON: %v", err)
	}

	// Écrire dans le fichier
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("erreur écriture fichier: %v", err)
	}

	fmt.Printf("✓ Métriques sauvegardées: %s\n", filepath)
	return nil
}

// GetLifetimeDistribution retourne la distribution des durées de vie pour les graphiques
func (sm *SimulationMetrics) GetLifetimeDistribution(agentType string) []float64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var durations []float64

	if agentType == "car" {
		for _, car := range sm.CarMetrics {
			if car.LifeDuration > 0 {
				durations = append(durations, car.LifeDuration)
			}
		}
	} else {
		for _, ped := range sm.PedestrianMetrics {
			if ped.LifeDuration > 0 {
				durations = append(durations, ped.LifeDuration)
			}
		}
	}

	return durations
}
