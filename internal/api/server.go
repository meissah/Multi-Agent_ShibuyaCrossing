package api

// Serveur API REST pour exposer les métriques de simulation
// Permet au dashboard HTML de récupérer les stats des simulations

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"
)

// Structure du serveur API
type Server struct {
	port int
}

// Créer un nouveau serveur avec le port spécifié
func NewServer(port int) *Server {
	return &Server{port: port}
}

// Démarrage du serveur HTTP
func (s *Server) Start() {
	// Routes disponibles
	http.HandleFunc("/api/metrics", s.ajouterCORS(s.recupererMetriques))
	http.HandleFunc("/api/comparison", s.ajouterCORS(s.recupererComparaison))
	http.HandleFunc("/api/clear", s.ajouterCORS(s.supprimerMetriques))

	adresse := fmt.Sprintf(":%d", s.port)
	log.Printf(" Serveur API lancé sur http://localhost%s", adresse)
	log.Printf("  Routes API :")
	log.Printf("    GET  /api/metrics      - Toutes les métriques")
	log.Printf("    GET  /api/comparison   - Comparaison Shibuya vs French")
	log.Printf("    POST /api/clear        - Supprimer les métriques")

	if err := http.ListenAndServe(adresse, nil); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}

// Ajouter les headers CORS pour autoriser les requêtes depuis le dashboard
func (s *Server) ajouterCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Gérer les requêtes OPTIONS (preflight)
		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, req)
	}
}

// ----------------------------------------------------------------
// Route GET /api/metrics - Récupère toutes les métriques
// ----------------------------------------------------------------
func (s *Server) recupererMetriques(w http.ResponseWriter, req *http.Request) {
	// Vérifier que c'est bien une requête GET
	if req.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Lire les fichiers de métriques
	dossierMetrics := "metrics_data"
	fichiers, err := os.ReadDir(dossierMetrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lecture métriques: %v", err), http.StatusInternalServerError)
		return
	}

	// Structure pour renvoyer les données
	type FichierMetrique struct {
		Nom  string              `json:"filename"`
		Type string              `json:"type"`
		Data stats.AnalysisStats `json:"data"`
	}

	metriques := []FichierMetrique{}

	// Parcourir tous les fichiers JSON
	for _, fichier := range fichiers {
		if filepath.Ext(fichier.Name()) != ".json" {
			continue
		}

		chemin := filepath.Join(dossierMetrics, fichier.Name())
		contenu, err := os.ReadFile(chemin)
		if err != nil {
			continue
		}

		var statsAnalyse stats.AnalysisStats
		if err := json.Unmarshal(contenu, &statsAnalyse); err != nil {
			continue
		}

		// Identifier le type de crossing
		typeSimu := "shibuya"
		if len(fichier.Name()) >= len("metrics_French") {
			if fichier.Name()[:len("metrics_French")] == "metrics_French" {
				typeSimu = "french"
			}
		}

		metriques = append(metriques, FichierMetrique{
			Nom:  fichier.Name(),
			Type: typeSimu,
			Data: statsAnalyse,
		})
	}

	// Renvoyer la réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metriques)
}

// ----------------------------------------------------------------
// Route GET /api/comparison - Compare Shibuya et French
// ----------------------------------------------------------------
func (s *Server) recupererComparaison(w http.ResponseWriter, req *http.Request) {
	// Vérifier que c'est bien une requête GET
	if req.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Appeler la fonction de comparaison
	comparaison, err := stats.CompareAllMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur comparaison: %v", err), http.StatusInternalServerError)
		return
	}

	// Renvoyer la comparaison en JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparaison)
}

// ----------------------------------------------------------------
// Route POST /api/clear - Supprime toutes les métriques
// ----------------------------------------------------------------
func (s *Server) supprimerMetriques(w http.ResponseWriter, req *http.Request) {
	// Vérifier que c'est bien une requête POST
	if req.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Supprimer le dossier de métriques
	if err := stats.ClearMetricsFolder(); err != nil {
		http.Error(w, fmt.Sprintf("Erreur suppression: %v", err), http.StatusInternalServerError)
		return
	}

	// Confirmation de succès
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Métriques supprimées",
	})
}
