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
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	// Définition des voies pour chaque sens de circulation
	const (
		// Voitures qui spawns au sud donc  allant vers le NORD (route de gauche)
		xRouteBasLane1 = 323.0 // la plus a gauche
		xRouteBasLane2 = 355.0 // la plus a droite

		// Voitures qui spawns au nord donc allant vers le SUD (route de droite)
		xRouteHautLane1 = 395.0 // la plus a gauche
		xRouteHautLane2 = 424.0 // la plus a droite

		// Voitures qui spawns a l'ouest -> allant vers l'EST (route du haut)
		xRouteGLane1 = 323.0 // la plus en haut
		xRouteGLane2 = 355.0 // la plus en bas

		// Voitures qui spawns a l'est -> allant vers l'OUEST (route du bas)
		yRouteDLane1 = 395.0 // la plus a haut
		yRouteDLane2 = 424.0 // la plus a bas
	)
	/* A savoir que le systeme de coordonnée est comme suit :
	   (0,0) -----------------> X augmente (vers 750)
	     |
	     |
	     |
	     |
	     v
	   Y augmente (vers 750)
	*/
	for {
		select {
		case <-ch.done:
			return
		case <-ticker.C:
			ch.mu.Lock()
			if len(ch.cars) < ch.maxCars {
				dir := Direction(rand.Intn(4))
				var x, y float64
				var lane int

				switch dir {
				case North: // Vient du bas, va vers le NORD (utilise la route de gauche)
					lane = rand.Intn(2)
					if lane == 0 {
						x = xRouteBasLane1
					} else {
						x = xRouteBasLane2
					}
					y = 750
				case South: // Vient du haut, va vers le SUD (utilise la route de droite)
					lane = rand.Intn(2)
					if lane == 0 {
						x = xRouteHautLane1
					} else {
						x = xRouteHautLane2
					}
					y = 0
				case East: // Vient de gauche, va vers l'EST (utilise la route du haut)
					lane = rand.Intn(2)
					x = 0
					if lane == 0 {
						y = xRouteGLane1
					} else {
						y = xRouteGLane2
					}
				case West: // Vient de droite, va vers l'OUEST (utilise la route du bas)
					lane = rand.Intn(2)
					x = 750
					if lane == 0 {
						y = yRouteDLane1
					} else {
						y = yRouteDLane2
					}
				}

				car := NewCar(len(ch.cars), x, y, dir, lane, ch.updates)
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
