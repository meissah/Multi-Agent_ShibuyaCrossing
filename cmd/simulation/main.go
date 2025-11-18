package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/game"
)

// Start game loop
func main() {

	//Taille fenetre du jeu
	ebiten.SetWindowSize(750, 750)

	//Titre de la fenetre du jeu
	ebiten.SetWindowTitle("Shibuya Crossing")

	ebiten.SetTPS(60) // 30 updates/draws per second

	//Initialisation de l'environnement du jeu
	game := game.NewGame()

	//lancement de la Boucle du jeu
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
