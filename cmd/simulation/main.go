package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/game"
)

// import (
// 	"image/color"
// 	_ "image/png"
// 	"log"
// 	"math"
// 	"math/rand"

// 	"github.com/hajimehoshi/ebiten/v2"
// 	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
// )

// var img *ebiten.Image

// // type Pedestrian struct {
// // 	X, Y             float64 // world position
// // 	VX, VY           float64 // velocity vector
// // 	Speed            float64 // magnitude of speed (e.g., 1.2 m/s)
// // 	State            string  // "waiting", "walking", "finished"
// // 	TargetX, TargetY float64 // destination point
// // }

// // type Car struct {
// // 	X, Y             float64
// // 	VX, VY           float64
// // 	Speed            float64 // e.g., 8–15 m/s (30–50 km/h)
// // 	Width            float64 // e.g., 1.8 (meters)
// // 	Length           float64 // e.g., 4.5 (meters)
// // 	Direction        float64 // heading in radians
// // 	State            string  // "driving", "waiting", "turning"
// // 	TargetX, TargetY float64
// // }

// // type PedestrianLight struct {
// // 	State    string             // "red" or "green"
// // 	Timer    float64            // seconds since last change
// // 	Duration map[string]float64 // e.g., {"red": 10, "green": 8}
// // }

// // type CarLight struct {
// // 	State    string // "red", "yellow", "green"
// // 	Timer    float64
// // 	Duration map[string]float64 // e.g., {"red": 10, "yellow": 2, "green": 10}
// // }

// func init() {
// 	var err error
// 	img, _, err = ebitenutil.NewImageFromFile("internal/assets/road.png")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// // // --- NEW: simple random movement ---
// // func (p *Pedestrian) Act(_ *Game) {
// // 	p.X += rand.Float64()*4 - 2 // move randomly within ±2px per frame
// // 	p.Y += rand.Float64()*4 - 2
// // 	if p.X < 0 {
// // 		p.X = 0
// // 	}
// // 	if p.Y < 0 {
// // 		p.Y = 0
// // 	}
// // 	if p.X > 750 {
// // 		p.X = 750
// // 	}
// // 	if p.Y > 750 {
// // 		p.Y = 750
// // 	}
// // }

// // func (c *Car) Act(_ *Game) {
// // 	c.X += rand.Float64()*6 - 3 // move randomly within ±3px per frame
// // 	c.Y += rand.Float64()*6 - 3
// // 	if c.X < 0 {
// // 		c.X = 0
// // 	}
// // 	if c.Y < 0 {
// // 		c.Y = 0
// // 	}
// // 	if c.X > 750 {
// // 		c.X = 750
// // 	}
// // 	if c.Y > 750 {
// // 		c.Y = 750
// // 	}
// // }

// type Game struct {
// 	Pedestrians []*Pedestrian
// 	Cars        []*Car
// 	CarLights   []*CarLight
// 	PedLights   []*PedestrianLight
// }

// // --- NEW: initialize some agents ---
// func NewGame() *Game {
// 	g := &Game{}
// 	for i := 0; i < 5; i++ {
// 		g.Pedestrians = append(g.Pedestrians, &Pedestrian{
// 			X: rand.Float64() * 700,
// 			Y: rand.Float64() * 700,
// 		})
// 	}
// 	for i := 0; i < 3; i++ {
// 		g.Cars = append(g.Cars, &Car{
// 			X:      rand.Float64() * 700,
// 			Y:      rand.Float64() * 700,
// 			Width:  20,
// 			Length: 40,
// 		})
// 	}
// 	return g
// }

// func (g *Game) Update() error {
// 	g_temp := g
// 	for _, agt := range g.Pedestrians {
// 		agt.Act(g_temp)
// 	}
// 	for _, agt := range g.Cars {
// 		agt.Act(g_temp)
// 	}

// 	return nil
// }

// func (g *Game) Draw(screen *ebiten.Image) {
// 	screen.Fill(color.RGBA{126, 109, 97, 1.000})
// 	op := &ebiten.DrawImageOptions{}
// 	op.GeoM.Scale(.15075, .392465)
// 	op.GeoM.Translate(300, 0)
// 	screen.DrawImage(img, op)

// 	op = &ebiten.DrawImageOptions{}

// 	op.GeoM.Scale(.15075, .392465)
// 	op.GeoM.Rotate(math.Pi / 2)
// 	op.GeoM.Translate(750, 300)
// 	screen.DrawImage(img, op)

// 	img1 := ebiten.NewImage(260, 150)
// 	img1.Fill(color.RGBA{71, 71, 71, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(245, 300)
// 	screen.DrawImage(img1, op)

// 	img1 = ebiten.NewImage(150, 260)
// 	img1.Fill(color.RGBA{71, 71, 71, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(300, 245)
// 	screen.DrawImage(img1, op)

// 	stripes := ebiten.NewImage(55, 5)
// 	stripes.Fill(color.RGBA{255, 255, 255, 255})
// 	for y := 300.0; y < 445; y += 7.5 {
// 		op = &ebiten.DrawImageOptions{}
// 		op.GeoM.Translate(245, y)
// 		screen.DrawImage(stripes, op)
// 	}
// 	for y := 300.0; y < 445; y += 7.5 {
// 		op = &ebiten.DrawImageOptions{}
// 		op.GeoM.Translate(450, y)
// 		screen.DrawImage(stripes, op)
// 	}

// 	stripes = ebiten.NewImage(5, 55)
// 	stripes.Fill(color.RGBA{255, 255, 255, 255})
// 	for x := 300.0; x < 445; x += 7.5 {
// 		op = &ebiten.DrawImageOptions{}
// 		op.GeoM.Translate(x, 245)
// 		screen.DrawImage(stripes, op)
// 	}
// 	for x := 300.0; x < 445; x += 7.5 {
// 		op = &ebiten.DrawImageOptions{}
// 		op.GeoM.Translate(x, 450)
// 		screen.DrawImage(stripes, op)
// 	}

// 	stripes = ebiten.NewImage(5, 55)
// 	stripes.Fill(color.RGBA{255, 255, 255, 255})
// 	for x := 0.0; x < 22; x++ {
// 		op = &ebiten.DrawImageOptions{}
// 		op.GeoM.Rotate(math.Pi / 4)
// 		op.GeoM.Translate(340.0+x*5.0, 300.0+x*5.0)
// 		screen.DrawImage(stripes, op)
// 	}

// 	img1 = ebiten.NewImage(55, 300)
// 	img1.Fill(color.RGBA{139, 26, 26, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(245, 0)
// 	screen.DrawImage(img1, op)

// 	img1 = ebiten.NewImage(55, 300)
// 	img1.Fill(color.RGBA{139, 26, 26, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(245, 450)
// 	screen.DrawImage(img1, op)

// 	img1 = ebiten.NewImage(55, 300)
// 	img1.Fill(color.RGBA{139, 26, 26, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(450, 0)
// 	screen.DrawImage(img1, op)

// 	img1 = ebiten.NewImage(55, 300)
// 	img1.Fill(color.RGBA{139, 26, 26, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(450, 450)
// 	screen.DrawImage(img1, op)

// 	img1 = ebiten.NewImage(300, 55)
// 	img1.Fill(color.RGBA{139, 26, 26, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(0, 245)
// 	screen.DrawImage(img1, op)

// 	img1 = ebiten.NewImage(300, 55)
// 	img1.Fill(color.RGBA{139, 26, 26, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(450, 245)
// 	screen.DrawImage(img1, op)

// 	img1 = ebiten.NewImage(300, 55)
// 	img1.Fill(color.RGBA{139, 26, 26, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(0, 450)
// 	screen.DrawImage(img1, op)

// 	img1 = ebiten.NewImage(300, 55)
// 	img1.Fill(color.RGBA{139, 26, 26, 255})
// 	op = &ebiten.DrawImageOptions{}
// 	op.GeoM.Translate(450, 450)
// 	screen.DrawImage(img1, op)

// 	// --- NEW: Draw pedestrians (circles) ---
// 	for _, p := range g.Pedestrians {
// 		ebitenutil.DrawCircle(screen, p.X, p.Y, 5, color.RGBA{0, 255, 0, 255})
// 	}

// 	// --- NEW: Draw cars (rectangles) ---
// 	for _, c := range g.Cars {
// 		rect := ebiten.NewImage(int(c.Width), int(c.Length))
// 		rect.Fill(color.RGBA{0, 0, 255, 255})
// 		op := &ebiten.DrawImageOptions{}
// 		op.GeoM.Translate(c.X, c.Y)
// 		screen.DrawImage(rect, op)
// 	}
// }

// func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
// 	return 750, 750
// }

func main() {
	ebiten.SetWindowSize(750, 750)
	ebiten.SetWindowTitle("Render an image")
	game := game.NewGame() // use our initializer
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
