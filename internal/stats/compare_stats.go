package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ComparisonResult contient les résultats de la comparaison
type ComparisonResult struct {
	ShibuyaAvailable bool
	FrenchAvailable  bool

	ShibuyaStats AnalysisStats
	FrenchStats  AnalysisStats
}

// AnalysisStats contient les statistiques moyennes
type AnalysisStats struct {
	AvgCarLifetime      float64 `json:"AvgCarLifetime"` // json tags pour les reponses API
	AvgPedLifetime      float64 `json:"AvgPedLifetime"`
	TotalCars           int     `json:"TotalCars"`
	TotalPeds           int     `json:"TotalPeds"`
	SimulationCount     int     `json:"SimulationCount"`
	AvgCarSpeed         float64 `json:"AvgCarSpeed"`
	CarSpawnRate        float64 `json:"CarSpawnRate"` // Paramètre d'entrée
	PedestrianSpawnRate float64 `json:"PedestrianSpawnRate"`
	Duration            float64 `json:"Duration"` // Durée de la simulation
}

// CompareAllMetrics compare toutes les simulations Shibuya vs French
func CompareAllMetrics() (*ComparisonResult, error) {
	metricsDir := "metrics_data"

	// Lire tous les fichiers JSON
	files, err := filepath.Glob(filepath.Join(metricsDir, "metrics_*.json"))
	if err != nil {
		return nil, fmt.Errorf("erreur lecture dossier: %v", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("aucun fichier de métriques trouvé dans %s", metricsDir)
	}

	var shibuyaMetrics []*SimulationMetrics
	var frenchMetrics []*SimulationMetrics

	// Charger tous les fichiers ( il y en a que 2 )
	for _, filepath := range files {
		data, err := os.ReadFile(filepath)
		if err != nil {
			continue
		}

		var metrics SimulationMetrics
		if err := json.Unmarshal(data, &metrics); err != nil {
			continue
		}

		if metrics.CrossingType == "shibuya" {
			shibuyaMetrics = append(shibuyaMetrics, &metrics)
		} else if metrics.CrossingType == "French" {
			frenchMetrics = append(frenchMetrics, &metrics)
		}
	}

	result := &ComparisonResult{
		ShibuyaAvailable: len(shibuyaMetrics) > 0,
		FrenchAvailable:  len(frenchMetrics) > 0,
	}

	// Calculer les moyennes Shibuya
	if result.ShibuyaAvailable {
		result.ShibuyaStats = calculateAverageStats(shibuyaMetrics)
	}

	// Calculer les moyennes French
	if result.FrenchAvailable {
		result.FrenchStats = calculateAverageStats(frenchMetrics)
	}

	return result, nil
}

// calculateAverageStats calcule les statistiques moyennes sur plusieurs simulations
func calculateAverageStats(metricsList []*SimulationMetrics) AnalysisStats {
	var totalCarLife, totalPedLife, totalCarSpeed, totalDuration float64
	var totalCars, totalPeds int
	var totalCarSpawnRate, totalPedSpawnRate float64
	validCars, validPeds := 0, 0

	for _, m := range metricsList {
		// Voitures
		for _, car := range m.CarMetrics {
			if car.LifeDuration > 0 {
				totalCarLife += car.LifeDuration
				validCars++
			}
		}
		totalCars += m.TotalCars

		// Piétons
		for _, ped := range m.PedestrianMetrics {
			if ped.LifeDuration > 0 {
				totalPedLife += ped.LifeDuration
				validPeds++
			}
		}
		totalPeds += m.TotalPedestrians

		// Nouvelles métriques
		totalCarSpeed += m.AvgCarSpeed
		totalCarSpawnRate += m.CarSpawnRate
		totalPedSpawnRate += m.PedestrianSpawnRate
		totalDuration += m.EndTime.Sub(m.StartTime).Seconds()
	}

	avgCarLife := 0.0
	if validCars > 0 {
		avgCarLife = totalCarLife / float64(validCars)
	}

	avgPedLife := 0.0
	if validPeds > 0 {
		avgPedLife = totalPedLife / float64(validPeds)
	}

	numMetrics := float64(len(metricsList))

	return AnalysisStats{
		AvgCarLifetime:      avgCarLife,
		AvgPedLifetime:      avgPedLife,
		TotalCars:           totalCars / len(metricsList),
		TotalPeds:           totalPeds / len(metricsList),
		SimulationCount:     len(metricsList),
		AvgCarSpeed:         totalCarSpeed / numMetrics,
		CarSpawnRate:        totalCarSpawnRate / numMetrics,
		PedestrianSpawnRate: totalPedSpawnRate / numMetrics,
		Duration:            totalDuration / numMetrics,
	}
}
