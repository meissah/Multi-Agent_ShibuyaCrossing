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
	CarLights    []*cars_lights.CarsLights
	PietonLights []*pieton_lights.PietonLight
	carHandler   *cars.CarHandler
}

func distance(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}

// Implémentation de CarEnvironment
func (g *Game) GetLights(x, y float64) []cars_lights.CarsLight {
	var nearby []cars_lights.CarsLight
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
		intersectionMinX = 245
		intersectionMaxX = 495
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
	g.Cars = make([]*cars.Car, 0) // Ajouter cette ligne
	g.carHandler.Start()
	return g
}

// Function responsible for updating the game's state (but not the graphics), automatically called by RunGame in main 60 times/second
func (g *Game) Update() error {
	// Mise à jour des voitures
	select {
	case update := <-g.carHandler.GetUpdates():
		cars := g.carHandler.GetCars()
		if update.ID < len(cars) {
			cars[update.ID].X = update.X
			cars[update.ID].Y = update.Y
			cars[update.ID].State = update.State
		}
	default:
	}

	// Mise à jour des piétons
	for _, p := range g.Pieton { // garder ce que ahmad a fait
		p.Act(g)
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

	// rotated road

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Rotate(math.Pi / 2)
	op.GeoM.Translate(750, 300)
	screen.DrawImage(img, op)

	// horizontal road
	img1 := ebiten.NewImage(260, 150)
	img1.Fill(color.RGBA{71, 71, 71, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(245, 300)
	screen.DrawImage(img1, op)

	// vertical road

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

	stripes = ebiten.NewImage(5, 55)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for x := 0.0; x < 22; x++ {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Rotate(math.Pi / 4)
		op.GeoM.Translate(340.0+x*5.0, 300.0+x*5.0)
		screen.DrawImage(stripes, op)
	}

	img1 = ebiten.NewImage(55, 300)
	img1.Fill(color.RGBA{139, 26, 26, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(245, 0)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(55, 300)
	img1.Fill(color.RGBA{139, 26, 26, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(245, 450)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(55, 300)
	img1.Fill(color.RGBA{139, 26, 26, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(450, 0)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(55, 300)
	img1.Fill(color.RGBA{139, 26, 26, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(450, 450)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(300, 55)
	img1.Fill(color.RGBA{139, 26, 26, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 245)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(300, 55)
	img1.Fill(color.RGBA{139, 26, 26, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(450, 245)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(300, 55)
	img1.Fill(color.RGBA{139, 26, 26, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 450)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(300, 55)
	img1.Fill(color.RGBA{139, 26, 26, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(450, 450)
	screen.DrawImage(img1, op)

	for _, p := range g.Pieton {
		vector.FillCircle(screen, float32(p.X), float32(p.Y), 5, color.RGBA{0, 255, 0, 255}, false)
	}

	// Dessin des voitures
	for _, c := range g.Cars {
		rect := ebiten.NewImage(int(50), int(50))
		rect.Fill(color.RGBA{0, 0, 255, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(c.X, c.Y)
		screen.DrawImage(rect, op)
	}
}

// sets the size of the game screen, automatically called by RunGame
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
