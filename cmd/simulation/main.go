package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/game"
)

// Initialise quelques paramètres de la boucle du jeu, puis la lance
func main() {

	//Taille fenetre du jeu
	ebiten.SetWindowSize(750, 750)

	//Titre de la fenetre du jeu
	ebiten.SetWindowTitle("Crossing Simulator")

	// 60 ticks par seconde
	ebiten.SetTPS(60)

	//Initialisation de l'environnement du jeu
	game := game.NewGame()

	//La simulation continue fluidement malgré que sa fenêtre ne soit pas en 'Focus'
	ebiten.SetRunnableOnUnfocused(true)

	//Lancement de la boucle du jeu
	// Boucle principale du jeu qui lance Update() 60 fois par seconde et Draw à une fréquence adapté à l'ordinateur
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
