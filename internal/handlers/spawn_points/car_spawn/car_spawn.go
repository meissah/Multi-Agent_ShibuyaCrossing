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
	stopChan      chan bool // pour arrêter proprement le spawner ( quand on cliquer sur le bouton retour au menu )
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
		xSpawnNordLaneDroite = 424.0

		// SPAWN à l'OUEST (à gauche)
		ySpawnOuestLaneHaute = 323.0
		ySpawnOuestLaneBasse = 355.0

		// SPAWN à l'EST (à droite)
		ySpawnEstLaneHaute = 395.0
		ySpawnEstLaneBasse = 424.0
	)

	//DirectionsHorizontal := [2]Direction{West, East}
	DirectionVertical := [2]Direction{North, South}
	// DirectionsAll := [4]Direction{North, South, East, West}

	//A ENLEVER QUAND YAURA LES FEUX
	// Choisir un point de SPAWN aléatoire
	//spawnPoint := DirectionsHorizontal[rand.Intn(len(DirectionsHorizontal))]
	spawnPoint := DirectionVertical[rand.Intn(len(DirectionVertical))]

	switch spawnPoint {
	case North: // SPAWN en HAUT (Nord)
		dir = South // La voiture ira donc vers le BAS (Sud) sinon
		lane = rand.Intn(2)
		if lane == 0 {
			x = xSpawnNordLaneGauche
		} else {
			x = xSpawnNordLaneDroite
		}
		y = -25 // Point de spawn en haut de l'écran

	case South: // SPAWN en BAS (Sud)
		dir = North // La voiture ira donc vers le HAUT (Nord)
		lane = rand.Intn(2)
		if lane == 0 {
			x = xSpawnSudLaneGauche
		} else {
			x = xSpawnSudLaneDroite
		}
		y = 775 // Point de spawn en bas de l'écran ( car l'écran fait 750px de haut )

	case East: // SPAWN à DROITE (Est)
		dir = West // La voiture ira donc vers la GAUCHE (Ouest)
		lane = rand.Intn(2)
		x = 775 // Point de spawn à droite de l'écran
		if lane == 0 {
			y = ySpawnEstLaneHaute
		} else {
			y = ySpawnEstLaneBasse
		}

	case West: // SPAWN à GAUCHE (Ouest)
		dir = East // La voiture ira donc vers la DROITE (Est)
		lane = rand.Intn(2)
		x = -25 // Point de spawn à gauche de l'écran
		if lane == 0 {
			y = ySpawnOuestLaneHaute
		} else {
			y = ySpawnOuestLaneBasse
		}
	}
	return x, y, dir, lane
}

// ChoisirVirage décide si la voiture doit tourner à gauche (25% de chances)
func ChooseToTurn(lane int, dir Direction) bool {
	if (lane != 0) && (dir == East || dir == North) {
		return false
	}
	if (lane == 0) && (dir == West || dir == South) {
		return false
	}
	return rand.Float64() < 0.50

}

// isSpawnPointClear vérifie si le point de spawn est libre de toute autre voiture proche
// pour éviter les collisions immédiates au spawn
// func (g *CarSpawn) isSpawnPointClear(x, y float64, env *[]Object) bool {
// 	const MIN_DISTANCE = 80.0 // Distance minimale entre voitures ( A Ajuster si besoin )

// 	for _, obj := range *env {
// 		if obj.Type() == "car" {
// 			carX, carY := obj.Position()

// 			// Calculer la distance euclidienne
// 			dx := carX - x
// 			dy := carY - y
// 			distance := math.Sqrt(dx*dx + dy*dy)

// 			// Si une voiture est trop proche, la zone n'est pas libre
// 			if distance < MIN_DISTANCE {
// 				return false
// 			}
// 		}
// 	}

// 	return true
// }

// fonction permettant de creer une seule nouvelle voiture
// Essayer de spawner plusieurs fois si bloqué
func (g *CarSpawn) spawnCars(AgentCountCh <-chan ObjectID, AgtsSliceCh chan<- agentslicehandler.AddAgtRequest, env *[]Object) {

	//const MAX_ATTEMPTS = 4 // Nombre maximum de tentatives
	// Vérifier si le point de spawn est libre
	const spawnSafetyInFront = 60.0 // Rayon de sécurité autour du point de spawn
	const spawnSafetyWidth = 15.0   // Rayon de sécurité autour du point de spawn

	//for attempt := 0; attempt < MAX_ATTEMPTS; attempt++ {
	// Choisir un point de spawn aléatoire
	x, y, dir, lane := ChoisirPointSpawn()

	var occupied bool = false

	// Vérifier la présence de voitures proches du point de spawn
	for _, pC := range *env {
		if pC.Type() != "car" {
			continue
		}
		switch dir {
		case North, South:
			// La voiture va vers le haut ou bas, vérifier devant elle
			if pC.PositionY() > y-spawnSafetyInFront && pC.PositionY() < y+spawnSafetyInFront &&
				pC.PositionX() < x+spawnSafetyWidth && pC.PositionX() > x-spawnSafetyWidth {
				// Point de spawn occupé, ne pas créer de voiture
				occupied = true
			}
		case East, West:
			// La voiture va vers la droite, vérifier devant elle
			if pC.PositionX() > x-spawnSafetyInFront && pC.PositionX() < x+spawnSafetyInFront &&
				pC.PositionY() < y+spawnSafetyWidth && pC.PositionY() > y-spawnSafetyWidth {
				// Point de spawn occupé, ne pas créer de voiture
				occupied = true
			}
		}

	}
	if !occupied {
		//  Zone libre, créer la voiture
		turnleft := ChooseToTurn(lane, dir) //choix aléatoire de si la voiture va tourner à gauche ou pas
		syncChan := make(chan int)
		car := cars.NewCar(<-AgentCountCh, x, y, dir, lane, turnleft, syncChan)

		// Lancer la goroutine de la voiture
		car.Start(env)

		// Ajouter l'agent à la boucle du jeu
		AgtsSliceCh <- agentslicehandler.AddAgtRequest{Agent: car, SyncChan: syncChan}

		// Succès !
		return
	}

	// Toutes les tentatives ont échoué (trop de trafic)
	// Ne pas spawner pour éviter les superpositions cas de trafic tres dense
	return

	//}
}

// commence la goroutine carspawn, qui introduit des nouvelles voitures de manière autonome au tick du jeu.
func (g *CarSpawn) Start(
	AgentCountCh <-chan ObjectID,
	AgtsSliceCh chan<- agentslicehandler.AddAgtRequest,
	env *[]Object,
) chan bool {
	// Créer le channel d'arrêt
	g.stopChan = make(chan bool)

	go func() {
		ticker := time.NewTicker(g.SpawnInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Spawn normal d'une voiture
				g.spawnCars(AgentCountCh, AgtsSliceCh, env)
			case <-g.stopChan:
				// Signal d'arrêt reçu, on termine la goroutine
				return
			}
		}
	}()

	// Retourner le channel pour que Game puisse l'arrêter
	return g.stopChan
}
