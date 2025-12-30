// Ce fichier définit la structure Game et son initialisation.
// Il gère les états du jeu (menu/simulation) et la boucle principale.
package game

import (
	"sync"
	"time"

	trafficlight "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/traffic_light"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
	agentcounthandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_count_handler"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	carspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/car_spawn"
	pietonspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/pieton_spawn"
	trafficmanger "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/traffic_manager"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Game représente l'état complet de l'application
type Game struct {
	//New Game Env
	Agents           []Agent
	Objets           []Object
	PersistentObjets []Object // pour l'affichage persistent des feux

	// État actuel (menu ou simulation)
	State  config.GameState        // État actuel du jeu si menu ou simulation
	Config config.SimulationConfig // Configuration de la simulation choisie dans le menu par l'utilisateur

	// Éléments du menu de configuration
	selectedCrossing string // "shibuya" ou "french"

	// Composants UI du menu
	validateButton  ui.Button
	backButton      ui.Button // Bouton de retour au menu principal quand on est en simulation
	dashboardButton ui.Button // Bouton pour ouvrir le dashboard dans le navigateur

	// Zones cliquables du menu
	shibuyaRect    ui.Rectangle
	frenchRect     ui.Rectangle
	carSpawnSlider ui.Slider
	pedSpawnSlider ui.Slider

	syncChansAgt sync.Map
	ObjectCount  ObjectID

	//Channels des differents microservices
	AddAgtSlice   chan<- agentslicehandler.AddAgtRequest
	DelAgtSlice   chan<- Agent
	ObjectCountCh <-chan ObjectID // recevoir le prochain ID d'agent disponible

	step int

	// Cote statistiques
	saveButtonPressed bool // Pour éviter les clics multiples sur le bouton de sauvegarde des métriques

	// pour arreter les spawners à la fin de la simulation ( c'est a dire quand on clique sur le bouton retour au menu )
	carSpawnerStop chan bool
	pedSpawnerStop chan bool
}

// NewGame crée et initialise une nouvelle instance du jeu en mode menu
func NewGame() *Game {
	g := &Game{
		State:           config.StateMenu,                                // commencer dans le menu
		Config:          config.DefaultConfig(),                          // configuration par défaut si l'utilisateur ne choisit rien
		shibuyaRect:     ui.Rectangle{X: 150, Y: 360, W: 200, H: 200},    // Correspondance exacte avec drawCrossingOption -> regler le probleme de click sur les images
		frenchRect:      ui.Rectangle{X: 400, Y: 360, W: 200, H: 200},    // Correspondance exacte avec drawCrossingOption
		carSpawnSlider:  ui.NewSlider(250, 165, 250, 15, 1.0, 50.0, 1.0), // Min=1, Max=50, aligné après le label
		pedSpawnSlider:  ui.NewSlider(250, 205, 250, 15, 1.0, 50.0, 5.0), // Min=1, Max=50, aligné après le label
		validateButton:  ui.NewButton(600, 650, 120, 40, "Valider"),
		backButton:      ui.NewButton(20, 20, 120, 40, "Retour"),
		dashboardButton: ui.NewButton(620, 700, 110, 30, "Dashboard"),
	}

	// Initialiser le traffic manager et les feux de circulation
	tm := &trafficmanger.TrafficManager{
		IsShibuyaMode: true,            // Par défaut en mode Shibuya
		StepDuration:  5 * time.Second, // Durée du feu vert
		RedDuration:   2 * time.Second, // Temps de sécurité entre les changements

		CarsNS: make([]*trafficlight.TrafficLight, 0),
		CarsEW: make([]*trafficlight.TrafficLight, 0),
		PedsNS: make([]*trafficlight.TrafficLight, 0),
		PedsEW: make([]*trafficlight.TrafficLight, 0),
	}

	// Configuration des feux de circulation (positions et types)
	// Pour chaque feu on définit : position, direction, type (voiture/piéton), et groupe (NS/EW)
	configs := []struct {
		x, y, w, h float64
		dir        Direction
		id         string
		kind       string // "car" ou "pedestrian"
		group      string // "NS" (Nord/Sud) ou "EW" (Est/Ouest)
	}{ // --- FEUX VOITURES ---

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

	// Créer les feux de circulation
	for _, cfg := range configs {
		// 1. Créer l'objet visuel du feu
		tlVisu := trafficlight.NewTrafficLightObject(cfg.x, cfg.y, cfg.w, cfg.h, cfg.dir, "Red", cfg.kind)
		g.Objets = append(g.Objets, tlVisu)
		g.PersistentObjets = append(g.PersistentObjets, tlVisu)

		// 2. Créer la logique du feu
		tlLogic := trafficlight.NewTrafficLight(cfg.id, 1*time.Hour, 1*time.Hour)
		tlLogic.Start()

		// 3. Classer le feu dans le manager selon son type et son groupe
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

	// Initialiser le système de métriques (pour les statistiques)
	stats.InitMetrics(g.Config)

	// Initialiser les channels des microservices
	g.ObjectCountCh = agentcounthandler.LancerAgentCountHandler(&g.ObjectCount, &g.Objets)
	g.AddAgtSlice, g.DelAgtSlice = agentslicehandler.LancerAgentSliceHandler(&g.Agents, &g.Objets, &g.syncChansAgt)
	g.generateRepulsionAreas()

	// Lancer les spawners et GARDER leurs channels d'arrêt
	pietonspawner := pietonspawn.PietonSpawn{SpawnInterval: g.Config.PedestrianSpawnRate}
	g.pedSpawnerStop = pietonspawner.Start(g.ObjectCountCh, g.AddAgtSlice, &g.Objets)

	carspawner := carspawn.CarSpawn{SpawnInterval: g.Config.CarSpawnRate}
	g.carSpawnerStop = carspawner.Start(g.ObjectCountCh, g.AddAgtSlice, &g.Objets)
}

// Génère les zones de repulsion pour les piétons (zones où ils ne doivent pas entrer)
func (g *Game) generateRepulsionAreas() {
	// Zones de répulsion des batiments
	g.addRepulsionAreas(
		[][2]Point{
			{*NewPoint(245, -50), *NewPoint(-50, 245)}, // Bâtiment nord-ouest, direction est
			{*NewPoint(245, 245), *NewPoint(-50, -50)}, // Bâtiment nord-ouest, direction sud
			{*NewPoint(505, 245), *NewPoint(800, -50)}, // Bâtiment nord-est, direction ouest
			{*NewPoint(800, 245), *NewPoint(505, -50)}, // Bâtiment nord-est, direction ouest
			{*NewPoint(245, 505), *NewPoint(-50, 800)}, // Bâtiment sud-ouest, direction est
			{*NewPoint(-50, 505), *NewPoint(245, 800)}, // Bâtiment sud-ouest, direction nord
			{*NewPoint(505, 800), *NewPoint(800, 505)}, // Bâtiment sud-est, direction ouest
			{*NewPoint(505, 505), *NewPoint(800, 800)}, // Bâtiment sud-est, direction nord
		},
		0.2, 10.0,
	)

	// Zones de répulsion de la route
	g.addRepulsionAreas(
		[][2]Point{
			{*NewPoint(-50, 300), *NewPoint(245, 375)}, // Route ouest, moitié haute
			{*NewPoint(245, 450), *NewPoint(-50, 375)}, // Route ouest, moitié basse
			{*NewPoint(505, 300), *NewPoint(800, 375)}, // Route est, moitié haute
			{*NewPoint(800, 450), *NewPoint(505, 375)}, // Route est, moitié basse
			{*NewPoint(300, 245), *NewPoint(375, -50)}, // Route nord, moitié gauche
			{*NewPoint(450, -50), *NewPoint(375, 245)}, // Route nord, moitié droite
			{*NewPoint(300, 800), *NewPoint(375, 505)}, // Route sud, moitié gauche
			{*NewPoint(450, 505), *NewPoint(375, 800)}, // Route sud, moitié droite
			{*NewPoint(300, 300), *NewPoint(450, 375)}, // Carrefour central, moitié haute
			{*NewPoint(450, 450), *NewPoint(300, 375)}, // Carrefour central, moitié basse
			{*NewPoint(300, 450), *NewPoint(375, 300)}, // Carrefour central, moitié gauche
			{*NewPoint(450, 300), *NewPoint(375, 450)}, // Carrefour central, moitié droite
		},
		0.05, 0.0,
	)
}

// Génère un ensemble cohérent de zones de répulsion et les ajoute au jeu
func (g *Game) addRepulsionAreas(coords [][2]Point, forceStrength, bufferSize float64) []*environment.RepulsionArea {
	ras := make([]*environment.RepulsionArea, len(coords))

	for _, coord := range coords {
		ra := environment.NewRepulsionArea(<-g.ObjectCountCh, coord[0], coord[1], forceStrength, bufferSize)
		ra.Hide()

		ras = append(ras, ra)
		g.Objets = append(g.Objets, ra)
	}

	return ras
}

// Layout définit la taille de la fenêtre du jeu (appelé automatiquement par Ebiten)
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// Si on est en simulation, fenêtre plus large pour les métriques
	if g.State == config.StateSimulation {
		return 1150, 750 // 750 pour la simulation + 400 pour les métriques
	}
	return 750, 750 // Menu reste 750x750
}
