package game

import (
	"image/color"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/pedestrian_lights"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/vehicule_lights"
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
	Pedestrians []*pedestrian.Pedestrian
	Cars        []*vehicule.Vehicule
	CarLights   []*vehicule_lights.VehiculeLight
	PedLights   []*pedestrian_lights.PedestrianLight
}

// initializes some agents, to be removed later as agents will have to continuously spawn.
func NewGame() *Game {
	g := &Game{}
	for i := 0; i < 5; i++ {
		g.Pedestrians = append(g.Pedestrians, &pedestrian.Pedestrian{
			X: rand.Float64() * 700,
			Y: rand.Float64() * 700,
		})
	}
	for i := 0; i < 3; i++ {
		g.Cars = append(g.Cars, &vehicule.Vehicule{
			X: rand.Float64() * 700,
			Y: rand.Float64() * 700,
		})
	}
	return g
}

// Function responsible for updating the game's state (but not the graphics), automatically called by RunGame in main 60 times/second
func (g *Game) Update() error {
	for _, agt := range g.Pedestrians {
		agt.Act(nil)
	}
	for _, agt := range g.Cars {
		agt.Act(nil)
	}

	return nil
}

// Function responsible for updating the graphics according to the game's state Games, automatically called by RunGame in main 60 times/second
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{126, 109, 97, 1.000})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Translate(300, 0)
	screen.DrawImage(img, op)

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

	for _, p := range g.Pedestrians {
		vector.FillCircle(screen, float32(p.X), float32(p.Y), 5, color.RGBA{0, 255, 0, 255}, false)
	}

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
