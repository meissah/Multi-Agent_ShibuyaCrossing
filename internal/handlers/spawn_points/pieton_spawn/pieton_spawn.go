package pietonspawn

import (
	"math/rand"
	"time"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/grid_handler"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/messaging"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/pathfinding"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// on a besoin d'une durée pour connaitre la frequence d'introduction d'une voiture
type PietonSpawn struct {
	SpawnInterval time.Duration
}

var exitLines = []utils.Line{
	*utils.NewLineFromPoints(*utils.NewPoint(245, -50), *utils.NewPoint(300, -50)), // Nord Gauche
	*utils.NewLineFromPoints(*utils.NewPoint(450, -50), *utils.NewPoint(505, -50)), // Nord Droite
	*utils.NewLineFromPoints(*utils.NewPoint(245, 800), *utils.NewPoint(300, 800)), // Sud Gauche
	*utils.NewLineFromPoints(*utils.NewPoint(450, 800), *utils.NewPoint(505, 800)), // Sud Droite
	*utils.NewLineFromPoints(*utils.NewPoint(-50, 245), *utils.NewPoint(-50, 300)), // Ouest Haut
	*utils.NewLineFromPoints(*utils.NewPoint(-50, 450), *utils.NewPoint(-50, 505)), // Ouest Bas
	*utils.NewLineFromPoints(*utils.NewPoint(800, 245), *utils.NewPoint(800, 300)), // Est Haut
	*utils.NewLineFromPoints(*utils.NewPoint(800, 450), *utils.NewPoint(800, 505)), // Est Bas
}

// fonction permettant de creer une seule nouvelle voiture
func (g *PietonSpawn) spawnPieton(
	AgentCountCh <-chan utils.ObjectID,
	AgtsSliceCh chan<- agentslicehandler.AddAgtRequest,
	gridChan chan messaging.GridRequest,
	env *[]utils.Object,
	spawns environment.SpawnAreaCollection,
) {
	//creation de la channel de synchronisation de l'agent piéton
	syncChan := make(chan int)
	//cree un nouveau piéton
	p := pedestrian.NewPedestrian(<-AgentCountCh, syncChan, gridChan)

	p.Coords = spawns.RandomPoint()
	p.OldCoords = p.Coords
	p.DesiredSeparation = utils.RandomFloatInRange(12.0, 18.0)
	p.MaxSpeed = utils.RandomFloatInRange(0.7, 1.2)

	// Choix de la destination (Ligne d'arrivée)
	// desination != spawn
	var destLine utils.Line
	for {
		destLine = exitLines[rand.Intn(len(exitLines))]
		// Calcul de la distance entre spawn et milieu de la ligne pour éviter l'immobilité
		midDest := utils.NewPoint((destLine.StartPoint().X()+destLine.EndPoint().X())/2,
			(destLine.StartPoint().Y()+destLine.EndPoint().Y())/2)
		dist := utils.NewVectorFromPoints(p.Coords, *midDest).Norm()
		if dist > 200 {
			break
		} // On veut une destination assez éloignée
	}
	p.FinalDestLine = &destLine

	t := rand.Float64()
	targetX := destLine.StartPoint().X() + t*(destLine.EndPoint().X()-destLine.StartPoint().X())
	targetY := destLine.StartPoint().Y() + t*(destLine.EndPoint().Y()-destLine.StartPoint().Y())
	goalPoint := *utils.NewPoint(targetX, targetY)

	// 4. Calcul du chemin A*
	path, found := pathfinding.FindPath(p.Coords, goalPoint)
	if found {
		p.Path = path
		p.CurrentNodeIndex = 0
	} else {
		// Si aucun chemin trouvé (ne devrait pas arriver), on donne une direction par défaut
		p.DesiredDirection = *utils.NewVector(0, 1)
	}

	p.Start(env)
	AgtsSliceCh <- agentslicehandler.AddAgtRequest{Agent: p, SyncChan: syncChan}
	gridChan <- messaging.AddObjectRequest{Object: p}
}

// commence la goroutine carspawn, qui introduit des nouvelles voitures de manière autonome au tick du jeu.
func (g *PietonSpawn) Start(
	AgentCountCh <-chan utils.ObjectID,
	AgtsSliceCh chan<- agentslicehandler.AddAgtRequest,
	env *[]utils.Object,
) {
	spawns := environment.SpawnAreaCollection{
		&environment.SpawnArea{X: 245, Y: -40, Width: 55, Height: 20}, // Nord G
		&environment.SpawnArea{X: 450, Y: -40, Width: 55, Height: 20}, // Nord D
		&environment.SpawnArea{X: -40, Y: 245, Width: 20, Height: 55}, // Ouest H
		&environment.SpawnArea{X: -40, Y: 450, Width: 20, Height: 55}, // Ouest B
		&environment.SpawnArea{X: 245, Y: 770, Width: 55, Height: 20}, // Sud G
		&environment.SpawnArea{X: 450, Y: 770, Width: 55, Height: 20}, // Sud D
		&environment.SpawnArea{X: 770, Y: 245, Width: 20, Height: 55}, // Est H
		&environment.SpawnArea{X: 770, Y: 450, Width: 20, Height: 55}, // Est B
	}

	go func() {
		// Initialisation de la grille de navigation A* avant le premier spawn
		// pathfinding.InitNavigationGrid()

		pedGrid := grid_handler.NewPedestrianGrid(850, 850, 20)
		pedGrid.Start()
		gridChan := pedGrid.RequestChan()

		ticker := time.NewTicker(g.SpawnInterval)
		defer ticker.Stop()

		for range ticker.C {
			g.spawnPieton(AgentCountCh, AgtsSliceCh, gridChan, env, spawns)
		}
	}()
}
