// Ce fichier gère toute la logique de mise à jour (Update) du jeu.
//
// Il contient :
//   - Update() : router principal qui dirige vers le menu ou la simulation
//   - UpdateMenu() : gestion des interactions utilisateur dans le menu
//   - UpdateSimulation() : boucle de simulation des agents
package game

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Update est appelé 60 fois par seconde par Ebiten (boucle principale du jeu)
func (g *Game) Update() error {
	fmt.Println("Goroutines running at the end of Update:", runtime.NumGoroutine())
	log.Printf("TPS: %.2f, FPS: %.2f", ebiten.ActualTPS(), ebiten.ActualFPS())
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

	// // Récupérer les agents qui ont bougé
	// currentAgtSlice := <-g.ReceiveAgtsCh

	// // --- CORRECTION CRUCIALE ICI ---
	// // 1. On vide la liste de dessin
	// g.Objets = nil

	// // 2. On remet les objets fixes (les feux)
	// g.Objets = append(g.Objets, g.PersistentObjets...)

	// // 3. On ajoute les agents dynamiques (voitures/piétons)
	// for _, agt := range currentAgtSlice {
	// 	g.Objets = append(g.Objets, agt)
	// }

	// Synchroniser tous les agents vivants en parallèle
	var wg sync.WaitGroup
	currentAgtSlice := g.Agents
	var agentsToRemove []Agent

	//fmt.Println("THE AGENT SLICE: ", currentAgtSlice)
	for _, agt := range currentAgtSlice {
		if agt.Alive() {
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

	var wg2 sync.WaitGroup

	for _, agt := range currentAgtSlice {
		if agt.Alive() {
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

	}

	// Attendre que tous les agents aient terminé leur pas
	wg2.Wait()

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
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.backButton.IsMouseOver {
			// Sauvegarder les métriques avant de retourner au menu
			if stats.GlobalMetrics != nil {
				stats.GlobalMetrics.SaveToFile()
			}

			// Sauvegarder les paramètres actuels avant de nettoyer ( objectif : les réutiliser ces au prochain pour la simu suivante  )
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
	}

	return false
}

// cleanupSimulation nettoie complètement les ressources de la simulation ( quand on retourne au menu avec le bouton retour )
func (g *Game) cleanupSimulation() {
	fmt.Println("Nettoyage de la simulation ")

	// Arrêter tous les spawners
	// Cela empêche la création de nouveaux agents pendant le nettoyage

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
		select {
		case agt.EndChan() <- 1:
			// Signal envoyé avec succès
		default:
			// Le channel est peut-être déjà fermé, on ignore
		}
	}

	// Attendre un peu que les goroutines se terminent proprement
	time.Sleep(100 * time.Millisecond)

	// Nettoyer les channels de synchronisation
	g.syncChansAgt.Range(func(key, value interface{}) bool {
		// Fermer le channel si possible
		if ch, ok := value.(chan int); ok {
			select {
			case <-ch: // Vider le channel
			default:
			}
		}
		g.syncChansAgt.Delete(key)
		return true
	})

	// Réinitialiser complètement les slices
	g.Agents = make([]Agent, 0)
	g.Objets = make([]Object, 0)

	// Remettre les objets persistants (feux de circulation)
	// Important : on recrée une nouvelle slice, pas de réutilisation
	for _, obj := range g.PersistentObjets {
		g.Objets = append(g.Objets, obj)
	}

	// Réinitialiser les compteurs
	g.ObjectCount = ObjectID(len(g.Objets))
	g.step = 0

	// Force le garbage collector à nettoyer
	runtime.GC() // permet de libérer la mémoire des goroutines terminées

	// test debug
	fmt.Println("Simulation nettoyée complètement")
	fmt.Printf(" Agents actifs: %d\n", len(g.Agents))
	fmt.Printf("  Objets dans la scène: %d\n", len(g.Objets))
	fmt.Printf("  Goroutines restantes: %d\n", runtime.NumGoroutine())
}
