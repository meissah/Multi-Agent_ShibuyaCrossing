package vehicule

import (
	"math/rand"
	"sync"
	"time"
)

// CarHandler gère toutes les voitures de la simulation
type CarHandler struct {
	cars    []*Car         // Liste des voitures
	updates chan CarUpdate // Canal pour les mises à jour
	done    chan bool      // Canal pour arrêter les goroutines
	mu      sync.RWMutex   // Mutex pour protéger l'accès aux voitures
	env     CarEnvironment // Référence à l'environnement
	maxCars int            // Nombre maximum de voitures
}

// NewCarHandler crée un nouveau gestionnaire de voitures
func NewCarHandler(env CarEnvironment, maxCars int) *CarHandler {
	return &CarHandler{
		cars:    make([]*Car, 0),
		updates: make(chan CarUpdate, 100), // Buffer de 100 mises à jour
		done:    make(chan bool),
		env:     env,
		maxCars: maxCars,
	}
}

// Start lance le gestionnaire de voitures
func (ch *CarHandler) Start() {
	go ch.spawnCars()
}

// Stop arrête toutes les goroutines
func (ch *CarHandler) Stop() {
	close(ch.done)
}

// spawnCars gère l'apparition de nouvelles voitures
func (ch *CarHandler) spawnCars() {
	ticker := time.NewTicker(2 * time.Second) // Une voiture toutes les 2 secondes
	defer ticker.Stop()

	for {
		select {
		case <-ch.done:
			return
		case <-ticker.C:
			ch.mu.Lock()
			if len(ch.cars) < ch.maxCars {
				// Choix aléatoire de la direction de départ
				dir := Direction(rand.Intn(4))
				var x, y float64

				// Position de départ selon la direction
				switch dir {
				case North:
					x, y = 245.0, 800.0
				case South:
					x, y = 495.0, -50.0
				case East:
					x, y = -50.0, 375.0
				case West:
					x, y = 800.0, 375.0
				}

				// Création et démarrage de la nouvelle voiture
				car := NewCar(
					len(ch.cars),
					x,
					y,
					dir,
					rand.Intn(2), // Voie aléatoire (0 ou 1)
					ch.updates,
				)

				// Attribution d'une décision de virage aléatoire
				// 50% tout droit, 25% gauche, 25% droite
				r := rand.Float64()
				if r < 0.5 {
					car.TurnDecision = Straight
				} else if r < 0.75 {
					car.TurnDecision = TurnLeft
				} else {
					car.TurnDecision = TurnRight
				}

				ch.cars = append(ch.cars, car)
				car.Start(ch.env, ch.done)
			}
			ch.mu.Unlock()
		}
	}
}

// GetUpdates retourne le canal des mises à jour
func (ch *CarHandler) GetUpdates() <-chan CarUpdate {
	return ch.updates
}

// GetCars retourne la liste des voitures (thread-safe)
func (ch *CarHandler) GetCars() []*Car {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.cars
}

// RemoveCar supprime une voiture de la simulation
func (ch *CarHandler) RemoveCar(id int) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	for i, car := range ch.cars {
		if car.ID == id {
			ch.cars = append(ch.cars[:i], ch.cars[i+1:]...)
			break
		}
	}
}
