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

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Update est appelé 60 fois par seconde par Ebiten (boucle principale du jeu)
func (g *Game) Update() error {
	fmt.Println("Goroutines running: ", runtime.NumGoroutine())
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

	// Mettre à jour l'état hover du bouton
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

	// Incrémenter le compteur de pas
	g.step++

	g.UpdateAllCharacCh <- 1

	// Synchroniser tous les agents vivants en parallèle
	var wg sync.WaitGroup
	currentAgtSlice := <-g.ReceiveAgtsCh
	if currentAgtSlice == nil {
		return nil
	}

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
			//demande suppression de l'agent
			go func() { g.DelAgtSlice <- agt }()
		}

	}

	// Attendre que tous les agents aient terminé leur pas
	wg.Wait()

	return nil
}
