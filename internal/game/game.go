package game

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	agentcounthandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_count_handler"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	objectslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/object_slice_handler"
	carspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/car_spawn"
	pietonspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/pieton_spawn"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
)

// ShowPaths est un toggle global (package-level) pour activer/désactiver
// l'affichage visuel des chemins A* pour les piétons.
var ShowPaths = true

// SetShowPaths permet d'activer ou désactiver l'affichage des chemins.
func SetShowPaths(v bool) { ShowPaths = v }

// ShowForbidden toggle pour afficher les cellules interdites (non franchissables).
var ShowForbidden = true

// SetShowForbidden active/désactive l'affichage des zones interdites.
func SetShowForbidden(v bool) { ShowForbidden = v }

// variable pour stocker l'image road.png
var imgRoad *ebiten.Image

// init initialise le chargement de l'image (appelé au démarrage du package)
func init() {
	var err error
	imgRoad, _, err = ebitenutil.NewImageFromFile("internal/assets/road.png")
	if err != nil {
		log.Fatal(err)
	}
}

// Structure de l'environment du jeu
type Game struct {
	//slice contenant l'ensemble des agents
	Agts []Agent
	//slice contenant une copy des objets visuels, pour eviter la triche sur le next step par les agents
	ObjtsCopy []Object
	//slice contenant les objets visuels, c'est la que le prochain step est ecrit
	Objts []Object

	//Point de synchronisation agents
	syncChansAgt sync.Map

	//Agents counter, assure des agentID valables
	AgentCount AgentID

	//Channels des differents microservices
	AgentCountCh    <-chan AgentID
	AddAgtSlice     chan<- agentslicehandler.AddAgtRequest
	ReceiveAgtSlice chan []Agent
	DelAgtSlice     chan<- Agent
	ObjtsSliceCh    chan<- ObjectAddRequest

	//savoir combien de fois on est passé par la boucle du jeu
	step int

	//pedestrians
	Pedestrians []pedestrian.Pedestrian
}

// Init de l'environnement du jeu
func NewGame() *Game {
	g := &Game{}

	//initialisation des différents channel des microservices, lancent ainsi leurs goroutines respectives
	g.AgentCountCh = agentcounthandler.LancerAgentCountHandler(&g.AgentCount)
	g.AddAgtSlice, g.ReceiveAgtSlice, g.DelAgtSlice = agentslicehandler.LancerAgentSliceHandler(&g.Agts, &g.syncChansAgt)
	g.ObjtsSliceCh = objectslicehandler.LancerObjectSliceHandler(&g.Objts)

	//lance la goroutine autonome de creation de voitures, on peut choisir la frequence ici
	carspawner := carspawn.CarSpawn{SpawnInterval: 100 * time.Millisecond}
	carspawner.Start(g.AgentCountCh, g.AddAgtSlice, g.ObjtsSliceCh, &g.ObjtsCopy)

	//lance la goroutine autonome de creation de voitures, on peut choisir la frequence ici
	pietonspawner := pietonspawn.PietonSpawn{SpawnInterval: 50 * time.Millisecond}
	pietonspawner.Start(g.AgentCountCh, g.AddAgtSlice, g.ObjtsSliceCh, &g.ObjtsCopy)

	return g
}

// appelle les agents pour qu'ils fassent leur prochain mouvement, lancé automatiquement par boucle du jeu
func (g *Game) Update() error {
	//on copie l'etat précédent des objets visuels
	g.ObjtsCopy = append([]Object(nil), g.Objts...)

	//on vide l'état courant des objets visuels pour permettre les agents d'actualiser leur future step
	g.Objts = g.Objts[:0]

	//goroutines qui appellent les agents pour qu'ils fassent leur next step
	var wg sync.WaitGroup
	g.step++
	for _, agt := range <-g.ReceiveAgtSlice {
		//ne cree pas de goroutine si agt mort
		if agt.Alive() {
			wg.Add(1)
			go func(agt Agent) {
				defer wg.Done()

				c, _ := g.syncChansAgt.Load(agt.AgtID())
				c.(chan int) <- g.step
				<-c.(chan int)

			}(agt)
		} else {
			//demande suppression de l'agent
			g.DelAgtSlice <- agt
		}
	}
	//on attend a ce que toutes les goroutines vivantes donnent leur next step
	wg.Wait()

	fmt.Println("Goroutines running at the end of Update:", runtime.NumGoroutine())

	return nil
}

// Affiche les éléments du slice Objts contenant le next step de tout les agents
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{126, 109, 97, 1.000})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Translate(300, 0)
	screen.DrawImage(imgRoad, op)

	op = &ebiten.DrawImageOptions{}

	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Rotate(math.Pi / 2)
	op.GeoM.Translate(750, 300)
	screen.DrawImage(imgRoad, op)

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

	// // dessine éventuellement une superposition des cellules interdites (recalcule la grille depuis AllowedRects)
	// if ShowForbidden {
	// 	width := pathfinding.DefaultWorldSize / pathfinding.DefaultCellSize
	// 	height := pathfinding.DefaultWorldSize / pathfinding.DefaultCellSize
	// 	grid := pathfinding.NewGridFromRects(width, height, pathfinding.DefaultCellSize, pathfinding.AllowedRects)
	// 	cellImg := ebiten.NewImage(pathfinding.DefaultCellSize, pathfinding.DefaultCellSize)
	// 	cellImg.Fill(color.RGBA{255, 255, 255, 80})
	// 	for y := 0; y < grid.Height; y++ {
	// 		for x := 0; x < grid.Width; x++ {
	// 			t := grid.Tiles[y][x]
	// 			if !t.Walkable {
	// 				pos := grid.ToWorld(x, y)
	// 				op := &ebiten.DrawImageOptions{}
	// 				// positionne le coin supérieur-gauche de l'image de la cellule
	// 				op.GeoM.Translate(pos.X-float64(grid.CellSize)/2, pos.Y-float64(grid.CellSize)/2)
	// 				screen.DrawImage(cellImg, op)
	// 			}
	// 		}
	// 	}
	// }

	// for _, p := range g.Pedestrians {
	// 	// dessine une croix sur la cible si une cible est assignée (non-zéro)
	// 	if p.TargetX != 0 || p.TargetY != 0 {
	// 		// barre horizontale
	// 		h := ebiten.NewImage(9, 3)
	// 		h.Fill(color.RGBA{255, 0, 0, 255})
	// 		op := &ebiten.DrawImageOptions{}
	// 		op.GeoM.Translate(p.TargetX-4, p.TargetY-1)
	// 		screen.DrawImage(h, op)
	// 		// barre verticale
	// 		v := ebiten.NewImage(3, 9)
	// 		v.Fill(color.RGBA{255, 0, 0, 255})
	// 		op = &ebiten.DrawImageOptions{}
	// 		op.GeoM.Translate(p.TargetX-1, p.TargetY-4)
	// 		screen.DrawImage(v, op)
	// 	}

	// 	// dessine éventuellement les points du chemin calculé (petits points jaunes)
	// 	if ShowPaths && len(p.Path) > 0 {
	// 		for _, wp := range p.Path {
	// 			vector.FillCircle(screen, float32(wp.X), float32(wp.Y), 2, color.RGBA{255, 255, 0, 200}, false)
	// 		}
	// 	}

	// 	vector.FillCircle(screen, float32(p.X), float32(p.Y), 5, color.RGBA{0, 255, 0, 255}, false)
	// }

	// positions := [][2]float64{
	// 	{0, 245}, {450, 245}, {0, 450}, {450, 450},
	// }
	// for _, pos := range positions {
	// 	op = &ebiten.DrawImageOptions{}
	// 	op.GeoM.Translate(pos[0], pos[1])
	// 	screen.DrawImage(img1, op)
	// }

	////Dessine les objets dans le slice Objts
	for _, c := range g.Objts {
		//dessine les voitures, a remplacer par select/case plutard pour prendre en compte l'ensemble des types d'objets
		switch c.Type() {
		case "car":
			// Dimensions selon la direction
			var w, h int
			switch c.Direction() {
			case North, South:
				w, h = 23, 33 // pour que les voitures ont la bonne forme suivant leur apparition
			case East, West:
				w, h = 33, 23 //
			}
			rect := ebiten.NewImage(w, h)

			// Utiliser une couleur fixe pour toutes les voitures
			rect.Fill(color.RGBA{0, 0, 255, 255}) // Bleu

			op := &ebiten.DrawImageOptions{}
			// centrer le rectangle sur la position de la voiture
			X, Y := c.Position()
			op.GeoM.Translate(X-float64(w)/2, Y-float64(h)/2)
			screen.DrawImage(rect, op)
		case "pieton":
			X, Y := c.Position()
			vector.FillCircle(screen, float32(X), float32(Y), 5, color.RGBA{0, 255, 0, 255}, false)
		}
	}

}

// définit la taille de l'écran du jeu, appelée automatiquement par RunGame
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
