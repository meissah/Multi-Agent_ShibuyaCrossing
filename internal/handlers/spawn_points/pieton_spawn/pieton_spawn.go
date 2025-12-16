package pietonspawn

import (
	"time"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// on a besoin d'une durée pour connaitre la frequence d'introduction d'une voiture
type PietonSpawn struct {
	SpawnInterval time.Duration
}

// fonction permettant de creer une seule nouvelle voiture
func (g *PietonSpawn) spawnPieton(AgentCountCh <-chan utils.ObjectID, AgtsSliceCh chan<- agentslicehandler.AddAgtRequest, env *[]utils.Object) {

	////Pietons pierrick part
	// définir les zones autorisées et de spawn basées sur le dessin (rectangles rouges + carré central)
	// allowed := []pathfinding.Rect{
	// 	// rectangles rouges verticaux (55x300) placés à 245 et 450
	// 	{X: 245, Y: 0, W: 55, H: 300},
	// 	{X: 245, Y: 450, W: 55, H: 300},
	// 	{X: 450, Y: 0, W: 55, H: 300},
	// 	{X: 450, Y: 450, W: 55, H: 300},
	// 	// rectangles rouges horizontaux (300x55) à 0 et 450
	// 	{X: 0, Y: 245, W: 300, H: 55},
	// 	{X: 450, Y: 245, W: 300, H: 55},
	// 	{X: 0, Y: 450, W: 300, H: 55},
	// 	{X: 450, Y: 450, W: 300, H: 55},
	// 	// carré central (approx.)
	// 	{X: 245, Y: 245, W: 205, H: 205},
	// 	// passages piétons (bandes blanches) reliant le carré central aux trottoirs
	// 	// étendre les passages pour couvrir entièrement les bords du carré central
	// 	// afin que les quatre trottoirs soient connectés sans trou
	// 	{X: 245, Y: 245, W: 55, H: 205}, // zone de passage verticale gauche
	// 	{X: 450, Y: 245, W: 55, H: 205}, // zone de passage verticale droite
	// 	{X: 245, Y: 245, W: 205, H: 55}, // zone de passage horizontale supérieure
	// 	{X: 245, Y: 450, W: 205, H: 55}, // zone de passage horizontale inférieure
	// }

	// Zones de spawn des piétons
	spawns := environment.SpawnAreaCollection{
		&environment.SpawnArea{X: 245, Y: 0, Width: 55, Height: 300},
		&environment.SpawnArea{X: 245, Y: 450, Width: 55, Height: 300},
		&environment.SpawnArea{X: 450, Y: 0, Width: 55, Height: 300},
		&environment.SpawnArea{X: 450, Y: 450, Width: 55, Height: 300},
		&environment.SpawnArea{X: 0, Y: 245, Width: 300, Height: 55},
		&environment.SpawnArea{X: 450, Y: 245, Width: 300, Height: 55},
		&environment.SpawnArea{X: 0, Y: 450, Width: 300, Height: 55},
		&environment.SpawnArea{X: 450, Y: 450, Width: 300, Height: 55},
	}

	//creation de la channel de synchronisation de l'agent voiture
	syncChan := make(chan int)

	//cree une structure voiture
	// spawn initial dans les rectangles de spawn
	p := pedestrian.NewPedestrian(<-AgentCountCh, syncChan)
	p.Start(env)
	p.DesiredDirection = *utils.NewVector(1, 0)
	p.DesiredSeparation = 15.0
	p.MaxSpeed = .5
	p.Coords = spawns.RandomPoint()

	AgtsSliceCh <- agentslicehandler.AddAgtRequest{Agent: p, SyncChan: syncChan}

}

// commence la goroutine carspawn, qui introduit des nouvelles voitures de manière autonome au tick du jeu.
func (g *PietonSpawn) Start(
	AgentCountCh <-chan utils.ObjectID,
	AgtsSliceCh chan<- agentslicehandler.AddAgtRequest,
	env *[]utils.Object,
) {
	go func() {
		ticker := time.NewTicker(g.SpawnInterval)
		defer ticker.Stop()

		for range ticker.C {
			g.spawnPieton(AgentCountCh, AgtsSliceCh, env)
		}
	}()
}
