package pietonspawn

import (
	"math/rand/v2"
	"time"

	pieton "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/pathfinding"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// on a besoin d'une durée pour connaitre la frequence d'introduction d'une voiture
type PietonSpawn struct {
	SpawnInterval time.Duration
}

// fonction permettant de creer une seule nouvelle voiture
func (g *PietonSpawn) spawnPieton(AgentCountCh <-chan ObjectID, AgtsSliceCh chan<- agentslicehandler.AddAgtRequest, env *[]Object) {

	////Pietons pierrick part
	// définir les zones autorisées et de spawn basées sur le dessin (rectangles rouges + carré central)
	allowed := []pathfinding.Rect{
		// rectangles rouges verticaux (55x300) placés à 245 et 450
		{X: 245, Y: 0, W: 55, H: 300},
		{X: 245, Y: 450, W: 55, H: 300},
		{X: 450, Y: 0, W: 55, H: 300},
		{X: 450, Y: 450, W: 55, H: 300},
		// rectangles rouges horizontaux (300x55) à 0 et 450
		{X: 0, Y: 245, W: 300, H: 55},
		{X: 450, Y: 245, W: 300, H: 55},
		{X: 0, Y: 450, W: 300, H: 55},
		{X: 450, Y: 450, W: 300, H: 55},
		// carré central (approx.)
		{X: 245, Y: 245, W: 205, H: 205},
		// passages piétons (bandes blanches) reliant le carré central aux trottoirs
		// étendre les passages pour couvrir entièrement les bords du carré central
		// afin que les quatre trottoirs soient connectés sans trou
		{X: 245, Y: 245, W: 55, H: 205}, // zone de passage verticale gauche
		{X: 450, Y: 245, W: 55, H: 205}, // zone de passage verticale droite
		{X: 245, Y: 245, W: 205, H: 55}, // zone de passage horizontale supérieure
		{X: 245, Y: 450, W: 205, H: 55}, // zone de passage horizontale inférieure
	}
	spawns := []pathfinding.Rect{
		// spawn / destination uniquement à l'intérieur des rectangles rouges (pas le carré central)
		{X: 245, Y: 0, W: 55, H: 300},
		{X: 245, Y: 450, W: 55, H: 300},
		{X: 450, Y: 0, W: 55, H: 300},
		{X: 450, Y: 450, W: 55, H: 300},
		{X: 0, Y: 245, W: 300, H: 55},
		{X: 450, Y: 245, W: 300, H: 55},
		{X: 0, Y: 450, W: 300, H: 55},
		{X: 450, Y: 450, W: 300, H: 55},
	}
	pathfinding.SetAllowedRects(allowed)
	pathfinding.SetSpawnRects(spawns)

	for i := 0; i < 5; i++ {

	}

	//creation de la channel de synchronisation de l'agent voiture
	syncChan := make(chan int)

	//cree une structure voiture
	// spawn initial dans les rectangles de spawn
	if x, y, ok := pathfinding.RandomPointInRects(spawns, rand.Float64); ok {
		npieton := pieton.NewPieton(<-AgentCountCh, x, y, syncChan)
		//lancement de la goroutine voiture
		npieton.Start(env)

		//ajout de l'agent à la boucle du jeu-> sa goroutine sera ainsi 'appelé' lors du prochain update()
		AgtsSliceCh <- agentslicehandler.AddAgtRequest{Agent: npieton, SyncChan: syncChan}
	} else {
		npieton := pieton.NewPieton(<-AgentCountCh, rand.Float64()*700, rand.Float64()*700, syncChan)
		//lancement de la goroutine voiture
		npieton.Start(env)

		//ajout de l'agent à la boucle du jeu-> sa goroutine sera ainsi 'appelé' lors du prochain update()
		AgtsSliceCh <- agentslicehandler.AddAgtRequest{Agent: npieton, SyncChan: syncChan}
	}

}

// commence la goroutine carspawn, qui introduit des nouvelles voitures de manière autonome au tick du jeu.
func (g *PietonSpawn) Start(
	AgentCountCh <-chan ObjectID,
	AgtsSliceCh chan<- agentslicehandler.AddAgtRequest,
	env *[]Object,
) {
	go func() {
		ticker := time.NewTicker(g.SpawnInterval)
		defer ticker.Stop()

		for range ticker.C {
			g.spawnPieton(AgentCountCh, AgtsSliceCh, env)
		}
	}()
}
