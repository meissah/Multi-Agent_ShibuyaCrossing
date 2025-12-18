// Ce fichier définit la structure Game et son initialisation.
// Il gère les états du jeu (menu/simulation) et la boucle principale.
package game

import (
	"sync"
	"time"

	trafficlight "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/traffic_light"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
	agentcounthandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_count_handler"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	carspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/car_spawn"
	pietonspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/pieton_spawn"
	trafficmanger "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/traffic_manager"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Game représente l'état complet de l'application
type Game struct {
	//New Game Env
	Agents []Agent
	Objets []Object

	//PersistentObjets []Object // pour l'affichage persistent des feux

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

	tm := &trafficmanger.TrafficManager{
		IsShibuyaMode: true,            // tester les modes (true/false)
		StepDuration:  5 * time.Second, // Temps de feu vert
		RedDuration:   2 * time.Second, // Temps de sécurité

		CarsNS: make([]*trafficlight.TrafficLight, 0),
		CarsEW: make([]*trafficlight.TrafficLight, 0),
		PedsNS: make([]*trafficlight.TrafficLight, 0),
		PedsEW: make([]*trafficlight.TrafficLight, 0),
	}

	g.validateButton = ui.NewButton(600, 650, 120, 40, "Valider")

	configs := []struct {
		x, y, w, h float64
		dir        Direction
		id         string
		kind       string // "car" ou "pedestrian"
		group      string // "NS" (Nord/Sud) ou "EW" (Est/Ouest)
	}{ // --- VOITURES ---

		{325, 525, 0, 0, East, "Car_Sud", "car", "NS"},
		{225, 325, 0, 0, North, "Car_Ouest", "car", "EW"},
		{525, 425, 0, 0, North, "Car_Est", "car", "EW"},
		{425, 225, 0, 0, East, "Car_Nord", "car", "NS"},

		// --- PIETONS
		// Ped_N (en haut) et Ped_S (en bas) traversent l'axe Vertical -> NS
		{375, 272.5, 150, 55, North, "Ped_N", "pedestrian", "NS"},
		{375, 477.5, 150, 55, North, "Ped_S", "pedestrian", "NS"},
		// Ped_W (à gauche) et Ped_E (à droite) traversent l'axe Horizontal -> EW
		{272.5, 375, 55, 150, East, "Ped_W", "pedestrian", "EW"},
		{477.5, 375, 55, 150, East, "Ped_E", "pedestrian", "EW"},
	}

	for _, cfg := range configs {
		tlVisu := trafficlight.NewTrafficLightObject(cfg.x, cfg.y, cfg.w, cfg.h, cfg.dir, "Red", cfg.kind)

		g.Objets = append(g.Objets, tlVisu)
		//g.PersistentObjets = append(g.PersistentObjets, tlVisu)

		// 2. Logique
		tlLogic := trafficlight.NewTrafficLight(cfg.id, 1*time.Hour, 1*time.Hour)
		tlLogic.Start()

		// C. Classement dans le Manager
		if cfg.kind == "car" {
			if cfg.group == "NS" {
				tm.CarsNS = append(tm.CarsNS, tlLogic)
			} else {
				tm.CarsEW = append(tm.CarsEW, tlLogic)
			}
		} else { // pedestrian
			if cfg.group == "NS" {
				tm.PedsNS = append(tm.PedsNS, tlLogic)
			} else {
				tm.PedsEW = append(tm.PedsEW, tlLogic)
			}
		}

		// D. Liaison Visuel <-> Logique (Goroutine d'écoute)
		go func(visu *trafficlight.TrafficLightObject, logic *trafficlight.TrafficLight) {
			chUpdate := logic.Subscribe()
			for newState := range chUpdate {
				visu.SetState(newState.String())
			}
		}(tlVisu, tlLogic)

	}

	// Lancement du Traffic Manager
	tm.Start()

	return g
}

// InitSimulation initialise tous les handlers et spawners pour démarrer la simulation
func (g *Game) InitSimulation() {
	// Initialiser les channels des microservices
	g.ObjectCountCh = agentcounthandler.LancerAgentCountHandler(&g.ObjectCount, &g.Objets)
	g.AddAgtSlice, g.ReceiveAgtsCh, g.ReceiveObjtsCh, g.DelAgtSlice = agentslicehandler.LancerAgentSliceHandler(&g.Agents, &g.Objets, &g.syncChansAgt)

	//lance les goroutines autonome de creation de voitures et pietons
	pietonspawner := pietonspawn.PietonSpawn{SpawnInterval: g.Config.PedestrianSpawnRate}
	pietonspawner.Start(g.ObjectCountCh, g.AddAgtSlice, &g.Objets)
	carspawner := carspawn.CarSpawn{SpawnInterval: g.Config.CarSpawnRate} // initialise la structure de spawn de voiture
	carspawner.Start(g.ObjectCountCh, g.AddAgtSlice, &g.Objets)
}

// Layout définit la taille de la fenêtre du jeu (appelé automatiquement par Ebiten)
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
