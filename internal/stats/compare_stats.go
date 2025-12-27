package stats

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// ComparisonResult contient les résultats de la comparaison
type ComparisonResult struct {
	ShibuyaAvailable bool
	FrenchAvailable  bool

	ShibuyaStats AnalysisStats
	FrenchStats  AnalysisStats

	CarDiffPercent float64
	PedDiffPercent float64
	Conclusion     string
}

// AnalysisStats contient les statistiques moyennes
type AnalysisStats struct {
	AvgCarLifetime  float64
	AvgPedLifetime  float64
	TotalCars       int
	TotalPeds       int
	SimulationCount int
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

	// Charger tous les fichiers
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

	// Calculer les différences
	if result.ShibuyaAvailable && result.FrenchAvailable {
		result.CarDiffPercent = ((result.FrenchStats.AvgCarLifetime - result.ShibuyaStats.AvgCarLifetime) /
			result.ShibuyaStats.AvgCarLifetime) * 100

		result.PedDiffPercent = ((result.FrenchStats.AvgPedLifetime - result.ShibuyaStats.AvgPedLifetime) /
			result.ShibuyaStats.AvgPedLifetime) * 100

		// Déterminer la conclusion
		if math.Abs(result.CarDiffPercent) < 5 && math.Abs(result.PedDiffPercent) < 5 {
			result.Conclusion = "Les deux systèmes sont équivalents"
		} else if result.CarDiffPercent < 0 && result.PedDiffPercent < 0 {
			result.Conclusion = "Shibuya est plus efficace"
		} else if result.CarDiffPercent > 0 && result.PedDiffPercent > 0 {
			result.Conclusion = "French crossing est plus efficace"
		} else {
			result.Conclusion = "Résultats mixtes"
		}
	}

	return result, nil
}

// calculateAverageStats calcule les statistiques moyennes sur plusieurs simulations
func calculateAverageStats(metricsList []*SimulationMetrics) AnalysisStats {
	var totalCarLife, totalPedLife float64
	var totalCars, totalPeds int
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
	}

	avgCarLife := 0.0
	if validCars > 0 {
		avgCarLife = totalCarLife / float64(validCars)
	}

	avgPedLife := 0.0
	if validPeds > 0 {
		avgPedLife = totalPedLife / float64(validPeds)
	}

	return AnalysisStats{
		AvgCarLifetime:  avgCarLife,
		AvgPedLifetime:  avgPedLife,
		TotalCars:       totalCars / len(metricsList),
		TotalPeds:       totalPeds / len(metricsList),
		SimulationCount: len(metricsList),
	}
}

// GetTopSimulations retourne les N meilleures simulations pour chaque type
func GetTopSimulations(n int) ([]*SimulationMetrics, []*SimulationMetrics, error) {
	metricsDir := "metrics_data"
	files, err := filepath.Glob(filepath.Join(metricsDir, "metrics_*.json"))
	if err != nil {
		return nil, nil, err
	}

	var shibuyaList, frenchList []*SimulationMetrics

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
			shibuyaList = append(shibuyaList, &metrics)
		} else if metrics.CrossingType == "French" {
			frenchList = append(frenchList, &metrics)
		}
	}

	// Trier par temps de vie moyen (plus court = meilleur)
	sort.Slice(shibuyaList, func(i, j int) bool {
		return shibuyaList[i].AvgCarLifetime+shibuyaList[i].AvgPedestrianLife <
			shibuyaList[j].AvgCarLifetime+shibuyaList[j].AvgPedestrianLife
	})

	sort.Slice(frenchList, func(i, j int) bool {
		return frenchList[i].AvgCarLifetime+frenchList[i].AvgPedestrianLife <
			frenchList[j].AvgCarLifetime+frenchList[j].AvgPedestrianLife
	})

	// Prendre les N premiers
	if len(shibuyaList) > n {
		shibuyaList = shibuyaList[:n]
	}
	if len(frenchList) > n {
		frenchList = frenchList[:n]
	}

	return shibuyaList, frenchList, nil
}

// FormatComparisonReport génère un rapport textuel formaté
func (cr *ComparisonResult) FormatComparisonReport() string {
	if !cr.ShibuyaAvailable && !cr.FrenchAvailable {
		return "Aucune donnée disponible. Lancez d'abord des simulations."
	}

	report := "\n========================================\n"
	report += "    RAPPORT DE COMPARAISON\n"
	report += "========================================\n\n"

	if cr.ShibuyaAvailable {
		report += fmt.Sprintf("SHIBUYA CROSSING (%d simulations)\n", cr.ShibuyaStats.SimulationCount)
		report += fmt.Sprintf("  Temps vie voitures: %.2fs\n", cr.ShibuyaStats.AvgCarLifetime)
		report += fmt.Sprintf("  Temps vie piétons:  %.2fs\n", cr.ShibuyaStats.AvgPedLifetime)
		report += fmt.Sprintf("  Total voitures:     %d\n", cr.ShibuyaStats.TotalCars)
		report += fmt.Sprintf("  Total piétons:      %d\n\n", cr.ShibuyaStats.TotalPeds)
	}

	if cr.FrenchAvailable {
		report += fmt.Sprintf("FRENCH CROSSING (%d simulations)\n", cr.FrenchStats.SimulationCount)
		report += fmt.Sprintf("  Temps vie voitures: %.2fs\n", cr.FrenchStats.AvgCarLifetime)
		report += fmt.Sprintf("  Temps vie piétons:  %.2fs\n", cr.FrenchStats.AvgPedLifetime)
		report += fmt.Sprintf("  Total voitures:     %d\n", cr.FrenchStats.TotalCars)
		report += fmt.Sprintf("  Total piétons:      %d\n\n", cr.FrenchStats.TotalPeds)
	}

	if cr.ShibuyaAvailable && cr.FrenchAvailable {
		report += "ANALYSE COMPARATIVE\n"
		report += fmt.Sprintf("  Différence voitures: %+.1f%%\n", cr.CarDiffPercent)
		report += fmt.Sprintf("  Différence piétons:  %+.1f%%\n\n", cr.PedDiffPercent)
		report += fmt.Sprintf("CONCLUSION: %s\n", cr.Conclusion)
	}

	report += "========================================\n"

	return report
}

// ClearMetricsFolder vide le dossier metrics_data
func ClearMetricsFolder() error {
	metricsDir := "metrics_data"

	// Lire tous les fichiers JSON
	files, err := filepath.Glob(filepath.Join(metricsDir, "metrics_*.json"))
	if err != nil {
		return fmt.Errorf("erreur lecture dossier: %v", err)
	}

	// Supprimer chaque fichier
	deletedCount := 0
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			fmt.Printf(" Erreur suppression %s: %v\n", file, err)
		} else {
			deletedCount++
		}
	}

	fmt.Printf("✓ %d fichiers de métriques supprimés\n", deletedCount)
	return nil
}
