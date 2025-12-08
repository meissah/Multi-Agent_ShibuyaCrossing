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
	ebiten.SetWindowTitle("Shibuya Crossing")

	// 24 updates par second
	ebiten.SetTPS(60)

	//Initialisation de l'environnement du jeu
	game := game.NewGame()

	ebiten.SetRunnableOnUnfocused(true)

	//lancement de la Boucle du jeu
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
