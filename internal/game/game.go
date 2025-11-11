package game

import (
	"image/color"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	pieton "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	cars "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"

	pieton_lights "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/pedestrian_lights"
	cars_lights "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/vehicule_lights"
)

// allows for the import of an image from a file, in this case the road pavement image
var img *ebiten.Image

// initializes the import of the image
func init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFile("internal/assets/road.png")
	if err != nil {
		log.Fatal(err)
	}
}

// Structure of the environment.
type Game struct {
	Pieton       []*pieton.Pieton
	Cars         []*cars.Car
	CarLights    []*cars_lights.VehiculeLight
	PietonLights []*pieton_lights.PedestrianLight
	carHandler   *cars.CarHandler
}

func distance(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}

// Implémentation de CarEnvironment
func (g *Game) GetLights(x, y float64) []cars_lights.VehiculeLight {
	var nearby []cars_lights.VehiculeLight
	for _, light := range g.CarLights {
		if distance(x, y, light.X, light.Y) < 100 {
			nearby = append(nearby, *light)
		}
	}
	return nearby
}

func (g *Game) IsIntersectionOccupied(x, y float64) bool {
	// Zone d'intersection
	const (
		intersectionMinX = 235
		intersectionMaxX = 515
		intersectionMinY = 300
		intersectionMaxY = 450
	)

	return x >= intersectionMinX && x <= intersectionMaxX &&
		y >= intersectionMinY && y <= intersectionMaxY
}

// initializes some agents, to be removed later as agents will have to continuously spawn.
func NewGame() *Game {
	g := &Game{}
	// Initialiser les piétons
	for i := 0; i < 5; i++ {
		g.Pieton = append(g.Pieton, &pieton.Pieton{
			X: rand.Float64() * 700,
			Y: rand.Float64() * 700,
		})
	}

	// Initialiser le gestionnaire de voitures
	g.carHandler = cars.NewCarHandler(g, 10)
	g.Cars = make([]*cars.Car, 0)
	g.carHandler.Start()
	return g
}

// Function responsible for updating the game's state (but not the graphics), automatically called by RunGame in main 60 times/second

func (g *Game) Update() error {
	// Drainer toutes les mises à jour du CarHandler
	for {
		select {
		case up := <-g.carHandler.GetUpdates():
			cars := g.carHandler.GetCars()
			if up.ID < len(cars) {
				cars[up.ID].X = up.X
				cars[up.ID].Y = up.Y
				cars[up.ID].State = up.State
			}
		default:
			goto done
		}
	}
done:

	// piétons (placeholder)
	for _, p := range g.Pieton {
		p.Act(nil)
	}
	return nil
}

// Function responsible for updating the graphics according to the game's state Games, automatically called by RunGame in main 60 times/second
func (g *Game) Draw(screen *ebiten.Image) {
	// background
	screen.Fill(color.RGBA{126, 109, 97, 1.000})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Translate(300, 0)
	screen.DrawImage(img, op)

	// horizontal road
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Rotate(math.Pi / 2)
	op.GeoM.Translate(750, 300)
	screen.DrawImage(img, op)

	img1 := ebiten.NewImage(260, 150)
	img1.Fill(color.RGBA{71, 71, 71, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(245, 300)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(150, 260)
	img1.Fill(color.RGBA{71, 71, 71, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(300, 245)
	screen.DrawImage(img1, op)

	stripes := ebiten.NewImage(55, 5)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for y := 300.0; y < 445; y += 7.5 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(245, y)
		screen.DrawImage(stripes, op)
	}
	for y := 300.0; y < 445; y += 7.5 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(450, y)
		screen.DrawImage(stripes, op)
	}

	// Bandes blanches diagonales
	stripes = ebiten.NewImage(5, 55)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for x := 300.0; x < 445; x += 7.5 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, 245)
		screen.DrawImage(stripes, op)
	}
	for x := 300.0; x < 445; x += 7.5 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, 450)
		screen.DrawImage(stripes, op)
	}

	// Bandes blanches diagonales inversées
	stripes = ebiten.NewImage(5, 55)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for x := 0.0; x < 22; x++ {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Rotate(math.Pi / 4)
		op.GeoM.Translate(340.0+x*5.0, 300.0+x*5.0)
		screen.DrawImage(stripes, op)
	}

	// Trottoirs
	img1 = ebiten.NewImage(55, 300)
	img1.Fill(color.RGBA{128, 128, 128, 255})

	positions := [][2]float64{
		{245, 0}, {245, 450}, {450, 0}, {450, 450},
	}
	for _, pos := range positions {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pos[0], pos[1])
		screen.DrawImage(img1, op)
	}

	img1 = ebiten.NewImage(300, 55)
	img1.Fill(color.RGBA{128, 128, 128, 255})

	positions = [][2]float64{
		{0, 245}, {450, 245}, {0, 450}, {450, 450},
	}
	for _, pos := range positions {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pos[0], pos[1])
		screen.DrawImage(img1, op)
	}

	// Dessiner les piétons
	for _, p := range g.Pieton {
		vector.FillCircle(screen, float32(p.X), float32(p.Y), 5, color.RGBA{0, 255, 0, 255}, false)
	}

	// Dessiner les voitures
	for _, c := range g.carHandler.GetCars() {
		// Dimensions selon la direction
		var w, h int
		switch c.Dir {
		case cars.North, cars.South:
			w, h = 23, 33 // verticales : étroites et hautes
		case cars.East, cars.West:
			w, h = 33, 23 // horizontales : larges et basses
		}
		rect := ebiten.NewImage(w, h)

		// Couleur selon l'état
		var carColor color.RGBA
		switch c.State {
		case "waiting", "stopped":
			carColor = color.RGBA{255, 0, 0, 255} // Rouge si en attente
		case "turning":
			carColor = color.RGBA{255, 165, 0, 255} // Orange si tourne
		default:
			carColor = color.RGBA{0, 0, 255, 255} // Bleu si roule
		}

		rect.Fill(carColor)
		op := &ebiten.DrawImageOptions{}
		// centrer le rectangle sur (c.X, c.Y)
		op.GeoM.Translate(c.X-float64(w)/2, c.Y-float64(h)/2)
		screen.DrawImage(rect, op)
	}

}

// sets the size of the game screen, automatically called by RunGame
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
