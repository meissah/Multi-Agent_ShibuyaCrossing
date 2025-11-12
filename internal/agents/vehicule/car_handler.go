package vehicule

import (
	"fmt"
)

/*
En gros le car handler est un gestionnaire de voitures qui s'occupe de créer, suivre et supprimer les voitures dans la simulation.
Il gère un ensemble de voitures, les fait apparaître périodiquement, et fournit des mises à jour sur leur position.

j'ai fait une goroutine par voiture et donc si une voiture doit modifier son état elle le fait elle même et envoie une mise à jour au car handler qui lui met a jour sa liste de voiture.
Ca passe par un canal de mise a jour (updates chan CarUpdate) qui est écouté par le game principal pour mettre a jour l'état graphique des voitures.
ca permet justement d'avoir un car handler qui s'occupe de la gestion globale des voitures et en passant par channel evite le data race
*/

// CarHandler gère toutes les voitures de la simulation
// interessant en utilisatn map c'est que ca permet d'avoir un accès rapide aux voitures par leur ID ( avec ajout suppression rapide)
// Surtout le fait que ca soit pas determinsite ( si on parcourt la map les voitures sont pas dans le meme ordre a chaque fois) -> realise et non determinsite
// donc quand on va faire un
type CarHandler struct {
	cars     map[int]*Car
	updateCh chan CarUpdate // canal pour recevoir les mises à jour des voitures, récupérer les positions et savoir quand elles doivent être supprimées
	done     chan bool
	nextID   int
}

// NewCarHandler crée un nouveau gestionnaire de voitures
func NewCarHandler() *CarHandler {
	ch := &CarHandler{
		cars:     make(map[int]*Car),
		updateCh: make(chan CarUpdate),
		done:     make(chan bool),
		nextID:   0,
	}
	// Lance la goroutine qui écoutera les mises à jour des voitures
	go ch.handleUpdates()
	return ch
}

// GetCars implémente la méthode de l'interface  v ( dans car.go).
// Elle retourne une slice de toutes les voitures gérées par le handler.
// Cette fonction est sûre car elle ne modifie pas la map.
func (ch *CarHandler) GetCars() []*Car {
	// On crée une nouvelle slice pour ne pas retourner un pointeur vers la map interne
	carSlice := make([]*Car, 0, len(ch.cars))
	for _, car := range ch.cars { // parcourt la map et ajoute chaque voiture à la slice ( de façon non déterministe car ch.cars est une map )
		carSlice = append(carSlice, car)
	}
	return carSlice
}

/* A implementer quand les feux seront faits
// pas oublier de le mettre dans  CarEnvironment aussi
func (ch *CarHandler) GetTrafficLights() []*cars_lights.VehiculeLight {
	// A faire
	return nil
}
*/

// AddCar crée une nouvelle voiture et la lance.
func (ch *CarHandler) AddCar(x, y float64, dir Direction, lane int) {
	car := NewCar(ch.nextID, x, y, dir, lane, ch.updateCh)
	ch.cars[ch.nextID] = car // on ajoute la voiture à la map avec son ID comme clé
	ch.nextID++              // Incrémente l'ID pour la prochaine voiture
	car.Start(ch, ch.done)   // Lance la goroutine de la voiture

}

// handleUpdates est la goroutine qui traite les messages des voitures.
// C'est le seul endroit où la map `ch.cars` est modifiée.
func (ch *CarHandler) handleUpdates() {
	for update := range ch.updateCh { // écoute les mises à jour des voitures ( lecture du channel )
		if update.Remove {
			// La voiture doit être supprimée
			delete(ch.cars, update.ID)
			fmt.Printf("Voiture %d supprimée.\n", update.ID)
		} else {
			// La voiture envoie une mise à jour de sa position
			if car, ok := ch.cars[update.ID]; ok {
				car.X = update.X
				car.Y = update.Y
			}
		}
	}
}

// Stop arrête toutes les goroutines
func (ch *CarHandler) Stop() {
	close(ch.done)
}

/*


┌─────────────────────────────────────────────────────────────┐
│                         GAME (game.go)                       │
│  - Appelle spawnCars() régulièrement                        │
│  - spawnCars() appelle carHandler.AddCar()                  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   CAR HANDLER (car_handler.go)               │
│                                                              │
│  Structure:                                                  │
│    - cars: map[int]*Car  (stocke toutes les voitures)      │
│    - updateCh: canal pour recevoir les updates              │
│    - done: canal pour arrêter toutes les voitures          │
│    - nextID: compteur pour les IDs uniques                  │
│                                                              │
│  Méthodes:                                                   │
│    - AddCar(): Crée une nouvelle voiture                    │
│    - GetCars(): Renvoie la liste des voitures (pour que     │
│                 chaque voiture puisse voir les autres)       │
│    - handleUpdates(): Goroutine qui écoute updateCh         │
│                                                              │
└──────┬──────────────────────────────────────────────────────┘
       │                      ▲
       │ crée                 │ envoie des
       │                      │ CarUpdate via updateCh
       ▼                      │
┌─────────────────────────────┴───────────────────────────────┐
│                      VOITURE 1 (car.go)                      │
│  Goroutine 1:                                                │
│    Start() lance une boucle infinie avec ticker             │
│      ├─ Percept(): regarde l'environnement                  │
│      │   └─ Appelle ch.GetCars() pour voir les autres      │
│      ├─ Deliberate(): décide quoi faire                     │
│      ├─ Act(): agit (bouge)                                 │
│      └─ emit(): envoie CarUpdate via updateCh               │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                      VOITURE 2 (car.go)                      │
│  Goroutine 2: (même cycle que Voiture 1)                    │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                      VOITURE N (car.go)                      │
│  Goroutine N: (même cycle que Voiture 1)                    │
└──────────────────────────────────────────────────────────────┘


// Start lance le gestionnaire de voitures
func (ch *CarHandler) Start() {
	go ch.spawnCars()
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
*/
