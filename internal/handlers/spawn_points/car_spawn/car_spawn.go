package carspawn

import (
	"math/rand"
	"time"

	cars "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// on a besoin d'une durée pour connaitre la frequence d'introduction d'une voiture
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
		xSpawnNordLaneDroite = 424.0

		// SPAWN à l'OUEST (à gauche)
		ySpawnOuestLaneHaute = 323.0
		ySpawnOuestLaneBasse = 355.0

		// SPAWN à l'EST (à droite)
		ySpawnEstLaneHaute = 395.0
		ySpawnEstLaneBasse = 424.0
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
		y = 0 // Point de spawn en haut de l'écran

	case South: // SPAWN en BAS (Sud)
		dir = North // La voiture ira donc vers le HAUT (Nord)
		lane = rand.Intn(2)
		if lane == 0 {
			x = xSpawnSudLaneGauche
		} else {
			x = xSpawnSudLaneDroite
		}
		y = 750 // Point de spawn en bas de l'écran ( car l'écran fait 750px de haut )

	case East: // SPAWN à DROITE (Est)
		dir = West // La voiture ira donc vers la GAUCHE (Ouest)
		lane = rand.Intn(2)
		x = 750 // Point de spawn à droite de l'écran
		if lane == 0 {
			y = ySpawnEstLaneHaute
		} else {
			y = ySpawnEstLaneBasse
		}

	case West: // SPAWN à GAUCHE (Ouest)
		dir = East // La voiture ira donc vers la DROITE (Est)
		lane = rand.Intn(2)
		x = 0 // Point de spawn à gauche de l'écran
		if lane == 0 {
			y = ySpawnOuestLaneHaute
		} else {
			y = ySpawnOuestLaneBasse
		}
	}
	return x, y, dir, lane
}

// ChoisirVirage décide si la voiture doit tourner à gauche (30% de chances)
func ChooseToTurn() bool {
	// 30% de chances de tourner à gauche
	if rand.Float64() < 0.99 { // a modif juste la pour tester
		return true
	}
	return false
}

// fonction permettant de creer une seule nouvelle voiture
func (g *CarSpawn) spawnCars(AgentCountCh <-chan AgentID, AgtsSliceCh chan<- agentslicehandler.Request, ObjtsSliceCh chan<- ObjectAddRequest, env *[]Object) {

	//choix aléatoire du point de spawn
	//x, y, dir, lane := ChoisirPointSpawn()
	x, y, dir, lane := 323.0, 750.0, North, 0 // juste pour tester

	//choix aléatoire de si la voiture va tourner à gauche ou pas
	turnleft := ChooseToTurn()
	//turnleft := true // juste pour tester

	//creation de la channel de synchronisation de l'agent voiture
	syncChan := make(chan int) // propre a l'agent

	//cree une structure voiture
	car := cars.NewCar(<-AgentCountCh, x, y, dir, lane, turnleft, syncChan) // permettre a update d'envoyer a la goroutine voiture un signal de mise a jour

	//lancement de la goroutine voiture
	car.Start(env, ObjtsSliceCh)

	//ajout de l'agent à la boucle du jeu-> sa goroutine sera ainsi 'appelé' lors du prochain update()
	AgtsSliceCh <- agentslicehandler.Request{Agent: car, SyncChan: syncChan}

}

// commence la goroutine carspawn, qui introduit des nouvelles voitures de manière autonome au tick du jeu.
func (g *CarSpawn) Start(
	AgentCountCh <-chan AgentID,
	AgtsSliceCh chan<- agentslicehandler.Request,
	ObjtsSliceCh chan<- ObjectAddRequest,
	env *[]Object,
) {
	go func() {
		ticker := time.NewTicker(g.SpawnInterval)
		defer ticker.Stop()

		for range ticker.C {
			g.spawnCars(AgentCountCh, AgtsSliceCh, ObjtsSliceCh, env)
		}
	}()
}
