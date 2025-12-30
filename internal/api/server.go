package api

// Serveur API REST simplifié pour exposer uniquement les métriques de comparaison
// Permet au dashboard HTML de récupérer les statistiques des deux types de crossing

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
	// Une seule route -> celle des comparaisons
	http.HandleFunc("/api/comparison", s.ajouterCORS(s.recupererComparaison))

	adresse := fmt.Sprintf(":%d", s.port)
	log.Printf("Serveur API lancé sur http://localhost%s", adresse)
	log.Printf("Route disponible : GET /api/comparison")

	if err := http.ListenAndServe(adresse, nil); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}

// Ajouter les headers CORS pour autoriser les requêtes depuis le dashboard ( a recup les données depuis l'api)
func (s *Server) ajouterCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Headers CORS nécessaires pour les requêtes depuis le navigateur
		w.Header().Set("Access-Control-Allow-Origin", "*")             // Autorise toutes les origines (pour le dashboard local)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS") // Autorise GET et OPTIONS
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // Autorise l'en-tête Content-Type

		// Gérer les requêtes OPTIONS (preflight)
		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, req)
	}
}

// Route GET /api/comparison - Compare Shibuya et French
// C'est la seule route nécessaire pour le dashboard
func (s *Server) recupererComparaison(w http.ResponseWriter, req *http.Request) {
	// Vérifier que c'est bien une requête GET
	if req.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Appeler la fonction de comparaison des stats
	comparaison, err := stats.CompareAllMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur comparaison: %v", err), http.StatusInternalServerError)
		return
	}

	// Renvoyer la comparaison en JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparaison)
}
