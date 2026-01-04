package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/api"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/game"
)

// Initialise quelques paramètres de la boucle du jeu, puis la lance
func main() {
	// Démarrer le serveur API dans une goroutine
	go func() {
		server := api.NewServer(8080)
		server.Start()
	}()

	// Configuration de la fenêtre du jeu (simulation uniquement)
	ebiten.SetWindowSize(1150, 750)
	ebiten.SetWindowTitle("Crossing Simulator")

	// 60 ticks par seconde
	ebiten.SetTPS(60)

	//La simulation continue fluidement malgré que sa fenêtre ne soit pas en 'Focus'
	ebiten.SetRunnableOnUnfocused(true)

	//Initialisation de l'environnement du jeu
	game := game.NewGame()

	log.Println("Simulation démarrée")
	log.Println("Dashboard disponible sur: http://localhost:3000")
	log.Println("API disponible sur: http://localhost:8080")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
