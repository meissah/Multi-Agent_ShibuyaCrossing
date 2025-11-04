package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/game"
)

// Start game loop
func main() {
	ebiten.SetWindowSize(750, 750)
	ebiten.SetWindowTitle("Render an image")
	game := game.NewGame() // use our initializer
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
