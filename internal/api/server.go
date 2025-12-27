package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"
)

// Server gère l'API HTTP pour exposer les métriques
type Server struct {
	port int
}

// NewServer crée un nouveau serveur API
func NewServer(port int) *Server {
	return &Server{port: port}
}

// Start démarre le serveur HTTP
func (s *Server) Start() {
	// Configuration CORS pour permettre les requêtes depuis React
	http.HandleFunc("/api/metrics", s.enableCORS(s.handleGetMetrics))
	http.HandleFunc("/api/comparison", s.enableCORS(s.handleGetComparison))
	http.HandleFunc("/api/clear", s.enableCORS(s.handleClearMetrics))

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf(" Serveur API démarré sur http://localhost%s", addr)
	log.Printf(" Endpoints disponibles:")
	log.Printf("   - GET  http://localhost%s/api/metrics", addr)
	log.Printf("   - GET  http://localhost%s/api/comparison", addr)
	log.Printf("   - POST http://localhost%s/api/clear", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Erreur serveur API: %v", err)
	}
}

// enableCORS ajoute les headers CORS
func (s *Server) enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// handleGetMetrics retourne toutes les métriques disponibles
func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Lire tous les fichiers de métriques
	metricsDir := "metrics_data"
	files, err := os.ReadDir(metricsDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lecture métriques: %v", err), http.StatusInternalServerError)
		return
	}

	type MetricFile struct {
		Filename string              `json:"filename"`
		Type     string              `json:"type"`
		Data     stats.AnalysisStats `json:"data"`
	}

	var metrics []MetricFile

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(metricsDir, file.Name()))
		if err != nil {
			continue
		}

		var analysisStats stats.AnalysisStats
		if err := json.Unmarshal(data, &analysisStats); err != nil {
			continue
		}

		// Déterminer le type (shibuya ou french)
		metricType := "shibuya"
		if filepath.Base(file.Name())[:len("metrics_French")] == "metrics_French" {
			metricType = "french"
		}

		metrics = append(metrics, MetricFile{
			Filename: file.Name(),
			Type:     metricType,
			Data:     analysisStats,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleGetComparison retourne la comparaison entre Shibuya et French
func (s *Server) handleGetComparison(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	comparison, err := stats.CompareAllMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur comparaison: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

// handleClearMetrics supprime toutes les métriques
func (s *Server) handleClearMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	if err := stats.ClearMetricsFolder(); err != nil {
		http.Error(w, fmt.Sprintf("Erreur suppression: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Métriques supprimées avec succès",
	})
}
