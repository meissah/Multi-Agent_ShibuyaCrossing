package carspawn

import (
	"math/rand"
	"time"

	cars "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// CarSpawn gère l'apparition des voitures
type CarSpawn struct {
	SpawnInterval time.Duration
}

// ChoisirPointSpawn choisit aléatoirement un point de spawn
func ChoisirPointSpawn() (x, y float64, dir Direction, lane int) {

	// Définition des coordonnées de spawn pour chaque voie. (. que j'ai trouvé en mesurant sur l'image + par tâtonnement)
	const (
		// SPAWN au SUD (en bas)
		xSpawnSudLaneGauche = 323.0
		xSpawnSudLaneDroite = 355.0

		// SPAWN au NORD (en haut)
		xSpawnNordLaneGauche = 395.0
		xSpawnNordLaneDroite = 430.0

		// SPAWN à l'OUEST (à gauche)
		ySpawnOuestLaneHaute = 320.0
		ySpawnOuestLaneBasse = 355.0

		// SPAWN à l'EST (à droite)
		ySpawnEstLaneHaute = 395.0
		ySpawnEstLaneBasse = 431.0
	)

	// Choisir un point de SPAWN aléatoire
	spawnPoint := Direction(rand.Intn(4))
	switch spawnPoint {
	case North: // SPAWN en HAUT (Nord)
		dir = South // La voiture ira donc vers le BAS (Sud) sinon
		lane = rand.Intn(2)
		if lane == 0 {
			x = xSpawnNordLaneGauche
		} else {
			x = xSpawnNordLaneDroite
		}
		y = -23 // Point de spawn en haut de l'écran

	case South: // SPAWN en BAS (Sud)
		dir = North // La voiture ira donc vers le HAUT (Nord)
		lane = rand.Intn(2)
		if lane == 0 {
			x = xSpawnSudLaneGauche
		} else {
			x = xSpawnSudLaneDroite
		}
		y = 773 // Point de spawn en bas de l'écran ( car l'écran fait 750px de haut )

	case East: // SPAWN à DROITE (Est)
		dir = West // La voiture ira donc vers la GAUCHE (Ouest)
		lane = rand.Intn(2)
		x = 773 // Point de spawn à droite de l'écran
		if lane == 0 {
			y = ySpawnEstLaneHaute
		} else {
			y = ySpawnEstLaneBasse
		}

	case West: // SPAWN à GAUCHE (Ouest)
		dir = East // La voiture ira donc vers la DROITE (Est)
		lane = rand.Intn(2)
		x = -23 // Point de spawn à gauche de l'écran
		if lane == 0 {
			y = ySpawnOuestLaneHaute
		} else {
			y = ySpawnOuestLaneBasse
		}
	}
	return x, y, dir, lane
}

// ChoisirVirage décide si la voiture doit tourner à gauche (25% de chances)
func ChooseToTurn() bool {
	return rand.Float64() < 0.25
}

// fonction permettant de creer une seule nouvelle voiture
// Essayer de spawner plusieurs fois si bloqué
func (g *CarSpawn) spawnCars(AgentCountCh <-chan ObjectID, AgtsSliceCh chan<- agentslicehandler.AddAgtRequest, env <-chan []Object, chCharacChn chan<- ChangeCharacRequest, allCharac <-chan []PhysicalCharacteristics) {

	// Choisir un point de spawn aléatoire
	x, y, dir, lane := ChoisirPointSpawn()

	// Increased safety radius for spawning (from 50 to 150 pixels)
	const spawnSafetyRadius = 150.0

	for _, pC := range <-allCharac {
		if pC.Type != "car" {
			continue
		}
		if pC.PositionX < x+spawnSafetyRadius && pC.PositionX > x-spawnSafetyRadius &&
			pC.PositionY < y+spawnSafetyRadius && pC.PositionY > y-spawnSafetyRadius {
			// Point de spawn occupé, ne pas créer de voiture
			return
		}
	}
	//  Zone libre, créer la voiture
	turnleft := ChooseToTurn() //choix aléatoire de si la voiture va tourner à gauche ou pas
	syncChan := make(chan int)
	car := cars.NewCar(<-AgentCountCh, x, y, dir, lane, turnleft, syncChan)

	// Lancer la goroutine de la voiture
	car.Start(env, chCharacChn, allCharac)

	// Ajouter l'agent à la boucle du jeu
	AgtsSliceCh <- agentslicehandler.AddAgtRequest{Agent: car, SyncChan: syncChan}

	// Succès !
	return
}

// commence la goroutine carspawn, qui introduit des nouvelles voitures de manière autonome au tick du jeu.
func (g *CarSpawn) Start(
	AgentCountCh <-chan ObjectID,
	AgtsSliceCh chan<- agentslicehandler.AddAgtRequest,
	env <-chan []Object,
	chCharacChn chan<- ChangeCharacRequest,
	allCharac <-chan []PhysicalCharacteristics,
) {
	go func() {
		ticker := time.NewTicker(g.SpawnInterval) // emet un tick(signal) à chaque intervalle de spawn
		defer ticker.Stop()

		for range ticker.C {
			g.spawnCars(AgentCountCh, AgtsSliceCh, env, chCharacChn, allCharac)

			// petit délai aléatoire (ici 0-15% de l'intervalle) pour éviter
			// que toutes les voitures spawent simultanément depuis les 4 directions
			delaiAleatoire := time.Duration(rand.Float64() * 0.15 * float64(g.SpawnInterval))
			time.Sleep(delaiAleatoire)
		}
	}()
}
