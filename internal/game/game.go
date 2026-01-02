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
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/carstopping"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/grid_handler"
	carspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/car_spawn"
	pietonspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/pieton_spawn"
	trafficmanger "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/traffic_manager"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/pathfinding"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Game représente l'état complet de l'application
type Game struct {
	//New Game Env
	Agents []Agent
	Objets []Object

	pedGrid *grid_handler.PedestrianGrid

	//PersistentObjets []Object // pour l'affichage persistent des feux
	TM *trafficmanger.TrafficManager

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

	DebugMode bool // true pour voir les lignes d'attraction

	//Channels des differents microservices
	AddAgtSlice   chan<- agentslicehandler.AddAgtRequest
	DelAgtSlice   chan<- Agent
	ObjectCountCh <-chan ObjectID // recevoir le prochain ID d'agent disponible

	step int

	// pour identifier les zones de repulsion pour les piétons
	CrosswalkZones map[string][]*environment.RepulsionArea

	LightsMap map[string]*trafficlight.TrafficLight
}

// NewGame crée et initialise une nouvelle instance du jeu en mode menu

func NewGame() *Game {
	g := &Game{
		State:          config.StateMenu,                                // commencer dans le menu
		Config:         config.DefaultConfig(),                          // configuration par défaut si l'utilisateur ne choisit rien
		shibuyaRect:    ui.Rectangle{X: 150, Y: 360, W: 200, H: 200},    // Correspondance exacte avec drawCrossingOption -> regler le probleme de click sur les images
		frenchRect:     ui.Rectangle{X: 400, Y: 360, W: 200, H: 200},    // Correspondance exacte avec drawCrossingOption
		carSpawnSlider: ui.NewSlider(250, 165, 250, 15, 1.0, 50.0, 1.0), // Min=1, Max=50, aligné après le label
		pedSpawnSlider: ui.NewSlider(250, 205, 250, 15, 1.0, 50.0, 5.0), // Min=1, Max=50, aligné après le label
		CrosswalkZones: make(map[string][]*environment.RepulsionArea),
		LightsMap:      make(map[string]*trafficlight.TrafficLight),
	}
	// permet d'éviter un deadlock
	g.ObjectCountCh = agentcounthandler.LancerAgentCountHandler(&g.ObjectCount, &g.Objets)

	g.AddAgtSlice, g.DelAgtSlice = agentslicehandler.LancerAgentSliceHandler(&g.Agents, &g.Objets, &g.syncChansAgt)

	tm := &trafficmanger.TrafficManager{
		ShibuyaFlag:     1,
		CarStepDuration: 8 * time.Second,  // Temps pour les voitures
		PedStepDuration: 16 * time.Second, // Temps long pour le Scramble
		RedDuration:     2 * time.Second,

		CarsNS:   make([]*trafficlight.TrafficLight, 0),
		CarsEW:   make([]*trafficlight.TrafficLight, 0),
		PedsNS:   make([]*trafficlight.TrafficLight, 0),
		PedsEW:   make([]*trafficlight.TrafficLight, 0),
		PedsDiag: make([]*trafficlight.TrafficLight, 0),
	}

	g.TM = tm

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
		{375, 375, 155, 55, North, "Ped_Diag", "pedestrian", "SHIBUYA"}, // FEU FANTÔME DIAGONALE

		// --- PIETONS
		// Ped_N (en haut) et Ped_S (en bas) traversent l'axe Vertical -> NS
		{375, 272.5, 150, 55, North, "Ped_N", "pedestrian", "NS"},
		{375, 477.5, 150, 55, North, "Ped_S", "pedestrian", "NS"},
		// Ped_W (à gauche) et Ped_E (à droite) traversent l'axe Horizontal -> EW
		{272.5, 375, 55, 150, East, "Ped_W", "pedestrian", "EW"},
		{477.5, 375, 55, 150, East, "Ped_E", "pedestrian", "EW"},
	}

	var chNS, chEW chan trafficlight.State

	for _, cfg := range configs {
		//objet visuel
		tlVisu := trafficlight.NewTrafficLightObject(cfg.id, cfg.x, cfg.y, cfg.w, cfg.h, cfg.dir, "Red", cfg.kind)

		g.Objets = append(g.Objets, tlVisu)
		//g.PersistentObjets = append(g.PersistentObjets, tlVisu)

		// 2. Logique
		tlLogic := trafficlight.NewTrafficLight(cfg.id, 1*time.Hour, 1*time.Hour)
		g.LightsMap[cfg.id] = tlLogic
		tlLogic.Start()

		// C. Classement dans le Manager
		if cfg.kind == "car" {
			if cfg.group == "NS" {
				g.TM.CarsNS = append(g.TM.CarsNS, tlLogic)
				chNS = tlLogic.Subscribe()
			} else {
				g.TM.CarsEW = append(g.TM.CarsEW, tlLogic)
				chEW = tlLogic.Subscribe()
			}
		} else {
			// Gestion des groupes piétons (NS, EW et SHIBUYA)
			// On ajoute la diagonale (SHIBUYA) aux groupes NS pour qu'elle s'allume lors du scramble
			if cfg.group == "NS" {
				g.TM.PedsNS = append(g.TM.PedsNS, tlLogic)
			} else if cfg.group == "EW" {
				g.TM.PedsEW = append(g.TM.PedsEW, tlLogic)
			} else if cfg.group == "SHIBUYA" {
				g.TM.PedsDiag = append(g.TM.PedsDiag, tlLogic)
			}
		}

		//abonnement de la zone au feu correspondant
		if cfg.kind == "pedestrian" {
			if zones, ok := g.CrosswalkZones[cfg.id]; ok {
				// Le feu va maintenant appeler Activate/Deactivate sur la zone
				for _, zone := range zones {
					tlLogic.AddZone(zone)
				}
			}
		}

		// D. Liaison Visuel <-> Logique (Goroutine d'écoute)
		go func(visu *trafficlight.TrafficLightObject, logic *trafficlight.TrafficLight) {
			chUpdate := logic.Subscribe()
			for newState := range chUpdate {
				//fmt.Println("Traffic Light", visu.ID, "changed to", newState.String())
				visu.SetState(newState.String())
			}
		}(tlVisu, tlLogic)

	}

	// Lancement du Traffic Manager

	g.pedGrid = grid_handler.NewPedestrianGrid(850, 850, 20)
	g.pedGrid.Start()

	carstopping.LancerCarStopperAtRedLight(g.ObjectCountCh, chNS, chEW, &g.Objets)

	return g
}

// InitSimulation initialise tous les handlers et spawners pour démarrer la simulation
func (g *Game) InitSimulation() {
	g.TM.SetMode(g.Config.CrossingType == "shibuya")

	pathfinding.InitNavigationGrid(g.Config.CrossingType == "shibuya")

	g.generateRepulsionAreas()
	for id, zones := range g.CrosswalkZones {
		if tl, ok := g.LightsMap[id]; ok {
			for _, zone := range zones {
				tl.AddZone(zone)
			}
		}
	}
	g.TM.Start()

	//lance les goroutines autonome de creation de voitures et pietons
	pietonspawner := pietonspawn.PietonSpawn{SpawnInterval: g.Config.PedestrianSpawnRate}
	pietonspawner.Start(g.ObjectCountCh, g.AddAgtSlice, &g.Objets, g.pedGrid)
	if g.selectedCrossing == "French" {
		carstopping.LancerCarStopperForPietons(g.ObjectCountCh, g.pedGrid, &g.Objets)
	}
	carspawner := carspawn.CarSpawn{SpawnInterval: g.Config.CarSpawnRate} // initialise la structure de spawn de voiture
	carspawner.Start(g.ObjectCountCh, g.AddAgtSlice, &g.Objets)

}

// Génère les zones de repulsion pour les piétons (zones où ils ne doivent pas entrer)
func (g *Game) generateRepulsionAreas() {

	//  Définition des sous-zones pour diviser les forces
	type SubZone struct {
		p1, p2       Point // Le rectangle de la sous-zone
		lineA, lineB Point // La ligne vers laquelle on pousse (le trottoir)
	}

	crosswalksSplit := map[string][]SubZone{
		// PASSAGE NORD (Ped_N)
		"Ped_N": {
			// Moitié Gauche : Pousse vers l'Ouest (Trottoir gauche)
			{
				*NewPoint(300, 245), *NewPoint(375, 300),
				*NewPoint(300, 245), *NewPoint(300, 300),
			},
			// Moitié Droite : Pousse vers l'Est (Trottoir droite)
			{
				*NewPoint(375, 245), *NewPoint(450, 300),
				*NewPoint(450, 245), *NewPoint(450, 300),
			},
		},
		// PASSAGE SUD (Ped_S) - Divisé Gauche/Droite
		"Ped_S": {
			// Moitié Gauche : Pousse vers l'Ouest
			{
				*NewPoint(300, 450), *NewPoint(375, 505),
				*NewPoint(300, 450), *NewPoint(300, 505),
			},
			// Moitié Droite : Pousse vers l'Est
			{
				*NewPoint(375, 450), *NewPoint(450, 505),
				*NewPoint(450, 450), *NewPoint(450, 505),
			},
		},
		// PASSAGE OUEST (Ped_W) - Divisé Haut/Bas
		"Ped_W": {
			// Moitié Haute : Pousse vers le Nord
			{
				*NewPoint(245, 300), *NewPoint(300, 375),
				*NewPoint(245, 300), *NewPoint(300, 300),
			},
			// Moitié Basse : Pousse vers le Sud
			{
				*NewPoint(245, 375), *NewPoint(300, 450),
				*NewPoint(245, 450), *NewPoint(300, 450),
			},
		},
		// PASSAGE EST (Ped_E) - Divisé Haut/Bas
		"Ped_E": {
			// Moitié Haute : Pousse vers le Nord
			{
				*NewPoint(450, 300), *NewPoint(505, 375),
				*NewPoint(450, 300), *NewPoint(505, 300),
			},
			// Moitié Basse : Pousse vers le Sud
			{
				*NewPoint(450, 375), *NewPoint(505, 450),
				*NewPoint(450, 450), *NewPoint(505, 450),
			},
		},
	}

	//  Génération des zones piétonnes divisées
	for id, subs := range crosswalksSplit {
		for _, sub := range subs {
			// Création de la sous-zone
			ra := environment.NewRepulsionArea(<-g.ObjectCountCh, sub.p1, sub.p2, 1.5, 5.0)
			ra.SetPedestrian(true)

			switch id {
			case "Ped_N":
				ra.ExpandSide("N", 10.0) // Étend vers le haut (Route Nord)
			case "Ped_S":
				ra.ExpandSide("S", 10.0) // Étend vers le bas (Route Sud)
			case "Ped_W":
				ra.ExpandSide("W", 10.0) // Étend vers la gauche (Route Ouest)
			case "Ped_E":
				ra.ExpandSide("E", 10.0) // Étend vers la droite (Route Est)
			}

			ra.SetAttractorLine(sub.lineA, sub.lineB)
			ra.Show()
			ra.Activate()

			ra.SetAttractorLine(sub.lineA, sub.lineB)

			g.Objets = append(g.Objets, ra)
			//g.PersistentObjets = append(g.PersistentObjets, ra)

			// On ajoute à la liste des zones pour cet ID de feu
			g.CrosswalkZones[id] = append(g.CrosswalkZones[id], ra)
		}
	}
	// ____________________________________Trouver une solution pour virer cette partie____________________________________
	//
	//  Gestion Spéciale de la Diagonale (Réintégrée pour éviter le crash)
	diagP1, diagP2 := *NewPoint(305, 305), *NewPoint(445, 445)
	raDiag := environment.NewRepulsionArea(<-g.ObjectCountCh, diagP1, diagP2, 0.5, 0.0)
	raDiag.SetPedestrian(true)
	raDiag.Hide()       // Invisible
	raDiag.Deactivate() // Inactive (ne pousse pas)

	g.Objets = append(g.Objets, raDiag)
	//g.PersistentObjets = append(g.PersistentObjets, raDiag)
	g.CrosswalkZones["Ped_Diag"] = append(g.CrosswalkZones["Ped_Diag"], raDiag)
	// ________________________________________________________________________________________________________________

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
		2.5, 10.0,
	)

	// Zones de répulsion de la route
	g.addRepulsionAreas(
		[][2]Point{
			{*NewPoint(-50, 300), *NewPoint(300, 375)}, // Route ouest, haut
			{*NewPoint(300, 450), *NewPoint(-50, 375)}, // Route ouest, bas
			{*NewPoint(450, 300), *NewPoint(800, 375)}, // Route est, haut
			{*NewPoint(800, 450), *NewPoint(450, 375)}, // Route est, bas

			{*NewPoint(300, 245), *NewPoint(375, -50)}, // Route nord, gauche (245 pour coller au trottoir)
			{*NewPoint(450, -50), *NewPoint(375, 245)}, // Route nord, droite
			{*NewPoint(300, 800), *NewPoint(375, 505)}, // Route sud, gauche
			{*NewPoint(450, 505), *NewPoint(375, 800)}, // Route sud, droite

			// Carrefour central strict
			// {*NewPoint(300, 300), *NewPoint(450, 375)},
			// {*NewPoint(450, 450), *NewPoint(300, 375)},
			// {*NewPoint(300, 450), *NewPoint(375, 300)},
			// {*NewPoint(450, 300), *NewPoint(375, 450)},
		},
		0.05, 0.0,
	)

	center := *NewPoint(375, 375)
	nw, ne := *NewPoint(300, 300), *NewPoint(450, 300)
	sw, se := *NewPoint(300, 450), *NewPoint(450, 450)

	// Création des 4 triangles d'éjection
	// Force élevée (2.5) pour un effet d'expulsion immédiat
	triangles := []struct {
		a, b, c Point
		push    Vector
		id      string
	}{
		{nw, ne, center, *NewVector(0, -1), "Ped_N"}, // Triangle Nord -> Pousse vers le haut
		{sw, se, center, *NewVector(0, 1), "Ped_S"},  // Triangle Sud -> Pousse vers le bas
		{nw, sw, center, *NewVector(-1, 0), "Ped_W"}, // Triangle Ouest -> Pousse vers la gauche
		{ne, se, center, *NewVector(1, 0), "Ped_E"},  // Triangle Est -> Pousse vers la droite
	}

	for _, t := range triangles {
		strength := 3.0

		if g.Config.CrossingType != "shibuya" {
			strength = 0.5
		}

		ra := environment.NewTriangleRepulsionArea(<-g.ObjectCountCh, t.a, t.b, t.c, t.push, strength)
		ra.Hide()

		// En mode French, ces zones sont toujours actives pour bloquer le centre
		// En mode Shibuya, elles seront gérées par le TrafficManager (Actives au rouge)
		// if g.Config.CrossingType != "shibuya" {
		// 	ra.Activate()
		// } else {
		// 	ra.Deactivate()
		// 	g.CrosswalkZones["Ped_Diag"] = append(g.CrosswalkZones["Ped_Diag"], ra)
		// }
		if g.Config.CrossingType == "shibuya" {
			ra.Deactivate() // Inactif au vert
			g.CrosswalkZones["Ped_Diag"] = append(g.CrosswalkZones["Ped_Diag"], ra)
		} else {

			ra.Activate()
		}
		g.Objets = append(g.Objets, ra)
	}
}

// Génère un ensemble cohérent de zones de répulsion et les ajoute au jeu
func (g *Game) addRepulsionAreas(coords [][2]Point, forceStrength, bufferSize float64) []*environment.RepulsionArea {
	ras := make([]*environment.RepulsionArea, 0, len(coords))

	for _, coord := range coords {
		ra := environment.NewRepulsionArea(<-g.ObjectCountCh, coord[0], coord[1], forceStrength, bufferSize)
		ra.Hide()

		ras = append(ras, ra)
		g.Objets = append(g.Objets, ra)
		//g.PersistentObjets = append(g.PersistentObjets, ra)
	}

	return ras
}

// Layout définit la taille de la fenêtre du jeu (appelé automatiquement par Ebiten)
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
