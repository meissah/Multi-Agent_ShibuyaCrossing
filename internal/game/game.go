package game

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	pieton "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	cars "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"

	pieton_lights "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/pedestrian_lights"
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
	// pas implementé pour l'instant
	Pieton       []*pieton.Pieton
	PietonLights []*pieton_lights.PedestrianLight

	carHandler *cars.CarHandler
	lastSpawn  time.Time // Un simple timer pour faire apparaître les voitures
}

func distance(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}

// initializes some agents, to be removed later as agents will have to continuously spawn.
func NewGame() *Game {
	g := &Game{}
	// Initialiser les piétons ( fait par ahmad pas touché )
	for i := 0; i < 5; i++ {
		g.Pieton = append(g.Pieton, &pieton.Pieton{
			X: rand.Float64() * 700,
			Y: rand.Float64() * 700,
		})
	}

	// COTE VOITURES
	// Initialiser le gestionnaire de voitures
	g.carHandler = cars.NewCarHandler()
	g.lastSpawn = time.Now()
	return g
}

/*
Rappel pour les resolution des spawn points des voitures  ( sachant qu'on travaille ici avec un écran de 750x750px )
 A savoir que le systeme de coordonnée est comme suit :
	   (0,0) -----------------> X augmente (vers 750)
	     |
	     |
	     |
	     |
	     v
	   Y augmente (vers 750)
*/
// spawnCars est la fonction qui crée les voitures en respectant les voies de circulation.
func (g *Game) spawnCars() {
	// On fait apparaître une voiture toutes les 1.5 secondes
	if time.Since(g.lastSpawn) < 1500*time.Millisecond {
		return
	}

	g.lastSpawn = time.Now()

	// Définition des coordonnées de spawn pour chaque voie. (. que j'ai trouvé en mesurant sur l'image + par tâtonnement)
	const (
		// SPAWN au SUD (en bas)
		xSpawnSudLaneGauche = 323.0
		xSpawnSudLaneDroite = 355.0

		// SPAWN au NORD (en haut)
		xSpawnNordLaneGauche = 395.0
		xSpawnNordLaneDroite = 424.0

		// SPAWN à l'OUEST (à gauche)
		ySpawnOuestLaneHaute = 323.0
		ySpawnOuestLaneBasse = 355.0

		// SPAWN à l'EST (à droite)
		ySpawnEstLaneHaute = 395.0
		ySpawnEstLaneBasse = 424.0
	)

	// Choisir un point de SPAWN aléatoire
	spawnPoint := cars.Direction(rand.Intn(4))
	var x, y float64
	var lane int
	var dir cars.Direction // La direction de déplacement sera déterminée par le point de spawn ( par default la voiture va tout droit (donc la direction est opposée au spawn) )
	switch spawnPoint {
	case cars.North: // SPAWN en HAUT (Nord)
		dir = cars.South // La voiture ira donc vers le BAS (Sud)
		lane = rand.Intn(2)
		if lane == 0 {
			x = xSpawnNordLaneGauche
		} else {
			x = xSpawnNordLaneDroite
		}
		y = 0 // Point de spawn en haut de l'écran

	case cars.South: // SPAWN en BAS (Sud)
		dir = cars.North // La voiture ira donc vers le HAUT (Nord)
		lane = rand.Intn(2)
		if lane == 0 {
			x = xSpawnSudLaneGauche
		} else {
			x = xSpawnSudLaneDroite
		}
		y = 750 // Point de spawn en bas de l'écran ( car l'écran fait 750px de haut )

	case cars.East: // SPAWN à DROITE (Est)
		dir = cars.West // La voiture ira donc vers la GAUCHE (Ouest)
		lane = rand.Intn(2)
		x = 750 // Point de spawn à droite de l'écran
		if lane == 0 {
			y = ySpawnEstLaneHaute
		} else {
			y = ySpawnEstLaneBasse
		}

	case cars.West: // SPAWN à GAUCHE (Ouest)
		dir = cars.East // La voiture ira donc vers la DROITE (Est)
		lane = rand.Intn(2)
		x = 0 // Point de spawn à gauche de l'écran
		if lane == 0 {
			y = ySpawnOuestLaneHaute
		} else {
			y = ySpawnOuestLaneBasse
		}
	}

	// Ajoute la voiture avec les bonnes coordonnées et la bonne direction
	g.carHandler.AddCar(x, y, dir, lane)
}

// Est responsable de la mise à jour de l'état du jeu, automatiquement appelé par RunGame dans main 60 fois/seconde
func (g *Game) Update() error {
	//COTE VOITURES
	//Declenche l'apparition des voitures
	g.spawnCars() // ps les voitures sont gérées par le car handler et se mettent à jour toutes seules

	// piétons ( pas fait par moi )
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
			w, h = 23, 33 // pour que les voitures ont la bonne forme suivant leur apparition
		case cars.East, cars.West:
			w, h = 33, 23 //
		}
		rect := ebiten.NewImage(w, h)

		// Utiliser une couleur fixe pour toutes les voitures
		rect.Fill(color.RGBA{0, 0, 255, 255}) // Bleu

		op := &ebiten.DrawImageOptions{}
		// centrer le rectangle sur la position de la voiture
		op.GeoM.Translate(c.X-float64(w)/2, c.Y-float64(h)/2)
		screen.DrawImage(rect, op)
	}

}

// sets the size of the game screen, automatically called by RunGame
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
