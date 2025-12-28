package pietonspawn

import (
	"time"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/grid_handler"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// on a besoin d'une durée pour connaitre la frequence d'introduction d'une voiture
type PietonSpawn struct {
	SpawnInterval time.Duration
}

// fonction permettant de creer une seule nouvelle voiture
func (g *PietonSpawn) spawnPieton(
	AgentCountCh <-chan utils.ObjectID,
	AgtsSliceCh chan<- agentslicehandler.AddAgtRequest,
	gridChan chan utils.Request,
	env *[]utils.Object,
	spawns environment.SpawnAreaCollection,
	unitVectors []utils.Vector,
) {
	//creation de la channel de synchronisation de l'agent piéton
	syncChan := make(chan int)

	//cree un nouveau piéton
	p := pedestrian.NewPedestrian(<-AgentCountCh, syncChan, gridChan)
	p.DesiredSeparation = utils.RandomFloatInRange(10.0, 20.0)
	p.MaxSpeed = .8
	p.Coords = spawns.RandomPoint()
	p.OldCoords = p.Coords
	p.DesiredDirection = unitVectors[utils.RandomIntInRange(0, len(unitVectors)-1)]
	p.Start(env)

	AgtsSliceCh <- agentslicehandler.AddAgtRequest{Agent: p, SyncChan: syncChan}
	gridChan <- utils.AddObjectRequest{Object: p}
}

// commence la goroutine carspawn, qui introduit des nouvelles voitures de manière autonome au tick du jeu.
func (g *PietonSpawn) Start(
	AgentCountCh <-chan utils.ObjectID,
	AgtsSliceCh chan<- agentslicehandler.AddAgtRequest,
	env *[]utils.Object,
) {
	spawns := environment.SpawnAreaCollection{
		&environment.SpawnArea{X: 245, Y: -50, Width: 55, Height: 50},
		&environment.SpawnArea{X: 450, Y: -50, Width: 55, Height: 50},
		&environment.SpawnArea{X: -50, Y: 245, Width: 50, Height: 55},
		&environment.SpawnArea{X: -50, Y: 450, Width: 50, Height: 55},
		&environment.SpawnArea{X: 245, Y: 750, Width: 55, Height: 50},
		&environment.SpawnArea{X: 450, Y: 750, Width: 55, Height: 50},
		&environment.SpawnArea{X: 750, Y: 245, Width: 50, Height: 55},
		&environment.SpawnArea{X: 750, Y: 450, Width: 50, Height: 55},
		&environment.SpawnArea{X: 750, Y: 450, Width: 50, Height: 55},
	}

	unitVectors := []utils.Vector{
		*utils.NewVector(0, 1),
		*utils.NewVector(0, -1),
		*utils.NewVector(1, 0),
		*utils.NewVector(-1, 0),
		*utils.NewVector(0, 0),
	}

	go func() {
		pedGrid := grid_handler.NewPedestrianGrid(850, 850, 20)
		pedGrid.Start()
		gridChan := pedGrid.RequestChan()

		ticker := time.NewTicker(g.SpawnInterval)
		defer ticker.Stop()

		for range ticker.C {
			g.spawnPieton(AgentCountCh, AgtsSliceCh, gridChan, env, spawns, unitVectors)
		}
	}()
}
