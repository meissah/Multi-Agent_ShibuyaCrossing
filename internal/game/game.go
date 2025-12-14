// Ce fichier définit la structure Game et son initialisation.
// Il gère les états du jeu (menu/simulation) et la boucle principale.
package game

import (
	"sync"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
	agentcounthandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_count_handler"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	carspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/car_spawn"
	pietonspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/pieton_spawn"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Game représente l'état complet de l'application
type Game struct {
	//New Game Env
	Agents []Agent
	Objets []Object

	// État actuel (menu ou simulation)
	State  config.GameState        // État actuel du jeu si menu ou simulation
	Config config.SimulationConfig // Configuration de la simulation choisie dans le menu par l'utilisateur

	// Éléments du menu de configuration
	selectedCrossing string // "shibuya" ou "french"

	// Composants UI du menu
	validateButton ui.Button
	shibuyaRect    ui.Rectangle
	frenchRect     ui.Rectangle
	carSpawnSlider ui.Slider
	pedSpawnSlider ui.Slider

	syncChansAgt sync.Map
	ObjectCount  ObjectID

	//Channels des differents microservices
	AddAgtSlice    chan<- agentslicehandler.AddAgtRequest
	ReceiveAgtsCh  <-chan []Agent
	ReceiveObjtsCh <-chan []Object
	DelAgtSlice    chan<- Agent
	ObjectCountCh  <-chan ObjectID // recevoir le prochain ID d'agent disponible

	step int
}

// NewGame crée et initialise une nouvelle instance du jeu en mode menu

func NewGame() *Game {
	g := &Game{
		State:          config.StateMenu,       // commencer dans le menu
		Config:         config.DefaultConfig(), // configuration par défaut si l'utilisateur ne choisit rien
		shibuyaRect:    ui.Rectangle{X: 200, Y: 300, W: 150, H: 150},
		frenchRect:     ui.Rectangle{X: 400, Y: 300, W: 150, H: 150},
		carSpawnSlider: ui.NewSlider(250, 165, 250, 15, 1.0, 50.0, 1.0), // Min=1, Max=50, aligné après le label
		pedSpawnSlider: ui.NewSlider(250, 205, 250, 15, 1.0, 50.0, 5.0), // Min=1, Max=50, aligné après le label

	}

	g.validateButton = ui.NewButton(600, 650, 120, 40, "Valider")

	return g
}

// InitSimulation initialise tous les handlers et spawners pour démarrer la simulation
func (g *Game) InitSimulation() {
	// Initialiser les channels des microservices
	g.ObjectCountCh = agentcounthandler.LancerAgentCountHandler(&g.ObjectCount)
	g.AddAgtSlice, g.ReceiveAgtsCh, g.ReceiveObjtsCh, g.DelAgtSlice = agentslicehandler.LancerAgentSliceHandler(&g.Agents, &g.Objets, &g.syncChansAgt)

	//lance les goroutines autonome de creation de voitures et pietons
	pietonspawner := pietonspawn.PietonSpawn{SpawnInterval: g.Config.PedestrianSpawnRate}
	pietonspawner.Start(g.ObjectCountCh, g.AddAgtSlice, g.ReceiveObjtsCh)
	carspawner := carspawn.CarSpawn{SpawnInterval: g.Config.CarSpawnRate} // initialise la structure de spawn de voiture
	carspawner.Start(g.ObjectCountCh, g.AddAgtSlice, g.ReceiveObjtsCh)
}

// Layout définit la taille de la fenêtre du jeu (appelé automatiquement par Ebiten)
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
