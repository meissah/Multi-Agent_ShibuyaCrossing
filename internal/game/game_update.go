// Ce fichier gère toute la logique de mise à jour (Update) du jeu.
//
// Il contient :
//   - Update() : router principal qui dirige vers le menu ou la simulation
//   - UpdateMenu() : gestion des interactions utilisateur dans le menu
//   - UpdateSimulation() : boucle de simulation des agents
package game

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	trafficlight "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/traffic_light"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/carstopping"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/grid_handler"
	trafficmanger "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/traffic_manager"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Update est appelé 60 fois par seconde par Ebiten (boucle principale du jeu)
func (g *Game) Update() error {
	// fmt.Println("Goroutines running at the end of Update:", runtime.NumGoroutine())
	// log.Printf("TPS: %.2f, FPS: %.2f", ebiten.ActualTPS(), ebiten.ActualFPS())
	switch g.State {
	case config.StateMenu:
		g.UpdateMenu()
	case config.StateSimulation:
		g.UpdateSimulation()
	}
	return nil
}

// UpdateMenu gère les interactions utilisateur dans l'écran de menu
func (g *Game) UpdateMenu() error {
	// Récupérer la position de la souris
	mx, my := ebiten.CursorPosition()
	mouseX, mouseY := float64(mx), float64(my)

	// Mettre à jour l'état hover des boutons
	g.validateButton.UpdateHover(mouseX, mouseY)
	g.dashboardButton.UpdateHover(mouseX, mouseY)

	// Mettre à jour les sliders (récupère la position du curseur)
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	g.carSpawnSlider.Update(mouseX, mouseY, mousePressed)
	g.pedSpawnSlider.Update(mouseX, mouseY, mousePressed)

	// Détection des clics de souris
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Clic sur l'option Shibuya
		if g.shibuyaRect.Contains(mouseX, mouseY) {
			g.selectedCrossing = "shibuya"
		}

		// Clic sur l'option French
		if g.frenchRect.Contains(mouseX, mouseY) {
			g.selectedCrossing = "French"
		}

		// Clic sur le bouton Dashboard
		if g.dashboardButton.IsMouseOver && g.checkBothMetricsExist() { // vérifier que les métriques ( a la fois french et shibuya ) existent avant d'ouvrir le dashboard
			g.openDashboard()
		}

		// Clic sur le bouton Valider
		if g.validateButton.IsMouseOver && g.selectedCrossing != "" {
			// Convertir en nombre entier d'agents/seconde
			carRate := int(g.carSpawnSlider.Value)
			pedRate := int(g.pedSpawnSlider.Value)

			// Appliquer la configuration choisie
			g.Config.CrossingType = g.selectedCrossing
			g.Config.CarSpawnRate = time.Duration(float64(time.Second) / float64(carRate))
			g.Config.PedestrianSpawnRate = time.Duration(float64(time.Second) / float64(pedRate))

			// Passer en mode simulation
			g.State = config.StateSimulation
			g.InitSimulation()
		}

	}

	return nil
}

// UpdateSimulation exécute un pas de simulation pour tous les agents vivants
func (g *Game) UpdateSimulation() error {

	// Vérifier le clic sur le bouton "Retour au menu"
	if g.handleBackButton() {
		return nil
	}

	// Incrémenter le compteur de pas
	g.step++

	// g.Objets = nil

	// g.Objets = append(g.Objets, g.PersistentObjets...)

	// for _, agt := range g.Agents {
	// 	g.Objets = append(g.Objets, agt)
	// }

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.DebugMode = !g.DebugMode
	}

	// Synchroniser tous les agents vivants en parallèle
	var wg sync.WaitGroup
	currentAgtSlice := g.Agents
	var agentsToRemove, agentsSynced []Agent

	//fmt.Println("THE AGENT SLICE: ", currentAgtSlice)

	// phase de perception et delibération
	for _, agt := range currentAgtSlice {
		if agt.Alive() {
			agentsSynced = append(agentsSynced, agt)
			wg.Add(1)
			go func(agt Agent) {
				defer wg.Done()

				// Récupérer le channel de synchronisation de l'agent
				c, _ := g.syncChansAgt.Load(agt.AgentID())

				// Signaler à l'agent de faire son prochain pas
				c.(chan int) <- g.step

				// Attendre que l'agent ait terminé
				<-c.(chan int)
			}(agt)
		} else {
			agentsToRemove = append(agentsToRemove, agt)
		}

	}
	wg.Wait()

	// Action des agents
	var wg2 sync.WaitGroup

	for _, agt := range agentsSynced {
		wg2.Add(1)
		go func(agt Agent) {
			defer wg2.Done()

			// Récupérer le channel de synchronisation de l'agent
			c, _ := g.syncChansAgt.Load(agt.AgentID())

			// Signaler à l'agent de faire son prochain pas
			c.(chan int) <- g.step

			// Attendre que l'agent ait terminé
			<-c.(chan int)
		}(agt)

	}

	// Attendre que tous les agents aient terminé leur pas
	wg2.Wait()

	// Supprimer les agents =>   ceux qui ont terminé leur trajet
	for _, agt := range agentsToRemove {
		g.DelAgtSlice <- agt
	}

	return nil
}

// handleBackButton gère le clic sur le bouton de retour au menu
func (g *Game) handleBackButton() bool {

	mx, my := ebiten.CursorPosition()
	mouseX, mouseY := float64(mx), float64(my)

	// Mettre à jour l'état hover du bouton retour
	g.backButton.UpdateHover(mouseX, mouseY)

	// Détecter le clic
	if (inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.backButton.IsMouseOver) || time.Since(g.startTime) > 180*time.Second {

		// Sauvegarder les métriques avant de retourner au menu
		if stats.GlobalMetrics != nil {
			stats.GlobalMetrics.SaveToFile()
		}

		// Sauvegarder les paramètres actuels avant de nettoyer ( objectif = les réutiliser pour la prochaine simu pour garder la meme config de spawn )
		carRate := 1.0 / g.Config.CarSpawnRate.Seconds()
		pedRate := 1.0 / g.Config.PedestrianSpawnRate.Seconds()

		// Nettoyer la simulation actuelle
		g.cleanupSimulation()

		// Retourner au menu
		g.State = config.StateMenu
		g.selectedCrossing = ""

		// On garde donc les mêmes valeurs en retournant au menu ( on garde meme nbr par seconde pour chaque type d'agent )
		g.carSpawnSlider.Value = carRate
		g.pedSpawnSlider.Value = pedRate
		return true

	}

	return false
}

// cleanupSimulation nettoie complètement les ressources de la simulation ( quand on retourne au menu avec le bouton retour, idée de relancé une simu mais donc de partir d'une base propre )
func (g *Game) cleanupSimulation() {
	fmt.Println("Nettoyage de la simulation ")

	// Il faut arreter tous les spawners avant de nettoyer
	if g.carSpawnerStop != nil {
		close(g.carSpawnerStop)
		g.carSpawnerStop = nil
	}
	if g.pedSpawnerStop != nil {
		close(g.pedSpawnerStop)
		g.pedSpawnerStop = nil
	}
	fmt.Println(" Spawners arrêtés")

	// Attendre un peu que les derniers spawns se terminent ( pour etre sur nettoyer proprement )
	time.Sleep(200 * time.Millisecond)

	// Envoyer signal de fin à tous les agents existants
	for _, agt := range g.Agents {
		g.DelAgtSlice <- agt
	}

	if g.endChanCarsStop4Peds != nil {
		g.endChanCarsStop4Peds <- 1
		g.endChanCarsStop4Peds = nil
	}

	fmt.Println("Arrivé ici 1")

	g.endChanCarsRedLight <- 1
	fmt.Println("Arrivé ici 3")
	g.TM.Stop()
	fmt.Println("Arrivé ici 4")
	g.pedGrid.Stop()

	// Remettre les objets persistants (feux de circulation)
	// Important : on recrée une nouvelle slice, pas de réutilisation
	// for _, obj := range g.PersistentObjets {
	// 	g.Objets = append(g.Objets, obj)
	// }

	fmt.Println("Arrivé ici 4321")
	//reinit de traffic
	g.CrosswalkZones = make(map[string][]*environment.RepulsionArea)
	g.LightsMap = make(map[string]*trafficlight.TrafficLight)

	fmt.Println("Arrivé ici 1")

	g.TM = &trafficmanger.TrafficManager{
		ShibuyaFlag:     1,
		CarStepDuration: 8 * time.Second,  // Temps pour les voitures
		PedStepDuration: 16 * time.Second, // Temps long pour le Scramble
		RedDuration:     5 * time.Second,

		CarsNS:   make([]*trafficlight.TrafficLight, 0),
		CarsEW:   make([]*trafficlight.TrafficLight, 0),
		PedsNS:   make([]*trafficlight.TrafficLight, 0),
		PedsEW:   make([]*trafficlight.TrafficLight, 0),
		PedsDiag: make([]*trafficlight.TrafficLight, 0),
	}

	g.Objets = make([]Object, 0)

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

		// 2. Créer la logique du feu
		tlLogic := trafficlight.NewTrafficLight(cfg.id, 1*time.Hour, 1*time.Hour)
		g.LightsMap[cfg.id] = tlLogic
		tlLogic.Start()

		// 3. Classer le feu dans le manager selon son type et son groupe
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
			mychUpdate := logic.Subscribe()
			myEndChan := make(chan int)
			g.endChanTL = append(g.endChanTL, myEndChan)
			for {
				select {
				case newState := <-mychUpdate:
					//fmt.Println("Traffic Light", visu.ID, "changed to", newState.String())
					visu.SetState(newState.String())

				case <-myEndChan:
					return
				}
			}
		}(tlVisu, tlLogic)

	}

	// Lancement du Traffic Manager

	g.pedGrid = grid_handler.NewPedestrianGrid(850, 850, 20)
	g.pedGrid.Start()

	g.endChanCarsRedLight = carstopping.LancerCarStopperAtRedLight(g.ObjectCountCh, chNS, chEW, &g.Objets)

	// Force le garbage collector à nettoyer
	runtime.GC() // permet de libérer la mémoire des goroutines terminées

	// test debug
	fmt.Println("Simulation nettoyée complètement")
	fmt.Printf(" Agents actifs: %d\n", len(g.Agents))
	fmt.Printf("  Objets dans la scène: %d\n", len(g.Objets))
	fmt.Printf("  Goroutines restantes: %d\n", runtime.NumGoroutine())
}

// Objectif d'avoir un bouton dashboard qui ouvre le dashboard HTML dans le navigateur lorsqu'on clique dessus ( s'active uniquement si les deux fichiers de métriques (french et shibuya ) existent )
// checkBothMetricsExist vérifie si les deux fichiers de métriques existent
func (g *Game) checkBothMetricsExist() bool {
	metricsDir := "metrics_data"

	shibuyaFile := filepath.Join(metricsDir, "metrics_shibuya.json")
	frenchFile := filepath.Join(metricsDir, "metrics_French.json")

	_, errShibuya := os.Stat(shibuyaFile)
	_, errFrench := os.Stat(frenchFile)

	return errShibuya == nil && errFrench == nil
}

// openDashboard ouvre le fichier HTML du dashboard dans le navigateur
func (g *Game) openDashboard() {
	dashboardPath := "dashboard.html" // Le fichier est à la racine du projet

	// Obtenir le chemin absolu
	absPath, err := filepath.Abs(dashboardPath)
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture du dashboard: %v\n", err)
		return
	}

	// Ouvrir dans le navigateur selon l'OS
	var cmd string
	switch runtime.GOOS {
	case "windows":
		cmd = "start"
	case "darwin": // macOS
		cmd = "open"
	default: // linux et autres
		cmd = "xdg-open"
	}

	// Utiliser la commande système pour ouvrir le fichier
	err = exec.Command(cmd, absPath).Start()
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture du dashboard: %v\n", err)
	} else {
		fmt.Println("Dashboard ouvert dans le navigateur")
	}
}
