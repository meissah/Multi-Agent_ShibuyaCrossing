package game

import (
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	agentcounthandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_count_handler"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	objectslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/object_slice_handler"
	carspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/car_spawn"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/pathfinding"
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
	AgentCountCh <-chan AgentID
	AgtsSliceCh  chan<- agentslicehandler.Request
	ObjtsSliceCh chan<- ObjectAddRequest

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
	g.AgtsSliceCh = agentslicehandler.LancerAgentSliceHandler(&g.Agts, &g.syncChansAgt)
	g.ObjtsSliceCh = objectslicehandler.LancerObjectSliceHandler(&g.Objts)

	//lance la goroutine autonome de creation de voitures, on peut choisir la frequence ici
	carspawner := carspawn.CarSpawn{SpawnInterval: 50 * time.Millisecond}
	carspawner.Start(g.AgentCountCh, g.AgtsSliceCh, g.ObjtsSliceCh, &g.ObjtsCopy)

	////Pietons pierrick part
	// définir les zones autorisées et de spawn basées sur le dessin (rectangles rouges + carré central)
	allowed := []pathfinding.Rect{
		// rectangles rouges verticaux (55x300) placés à 245 et 450
		{X: 245, Y: 0, W: 55, H: 300},
		{X: 245, Y: 450, W: 55, H: 300},
		{X: 450, Y: 0, W: 55, H: 300},
		{X: 450, Y: 450, W: 55, H: 300},
		// rectangles rouges horizontaux (300x55) à 0 et 450
		{X: 0, Y: 245, W: 300, H: 55},
		{X: 450, Y: 245, W: 300, H: 55},
		{X: 0, Y: 450, W: 300, H: 55},
		{X: 450, Y: 450, W: 300, H: 55},
		// carré central (approx.)
		{X: 245, Y: 245, W: 205, H: 205},
		// passages piétons (bandes blanches) reliant le carré central aux trottoirs
		// étendre les passages pour couvrir entièrement les bords du carré central
		// afin que les quatre trottoirs soient connectés sans trou
		{X: 245, Y: 245, W: 55, H: 205}, // zone de passage verticale gauche
		{X: 450, Y: 245, W: 55, H: 205}, // zone de passage verticale droite
		{X: 245, Y: 245, W: 205, H: 55}, // zone de passage horizontale supérieure
		{X: 245, Y: 450, W: 205, H: 55}, // zone de passage horizontale inférieure
	}
	spawns := []pathfinding.Rect{
		// spawn / destination uniquement à l'intérieur des rectangles rouges (pas le carré central)
		{X: 245, Y: 0, W: 55, H: 300},
		{X: 245, Y: 450, W: 55, H: 300},
		{X: 450, Y: 0, W: 55, H: 300},
		{X: 450, Y: 450, W: 55, H: 300},
		{X: 0, Y: 245, W: 300, H: 55},
		{X: 450, Y: 245, W: 300, H: 55},
		{X: 0, Y: 450, W: 300, H: 55},
		{X: 450, Y: 450, W: 300, H: 55},
	}
	pathfinding.SetAllowedRects(allowed)
	pathfinding.SetSpawnRects(spawns)

	for i := 0; i < 5; i++ {
		// spawn initial dans les rectangles de spawn
		if x, y, ok := pathfinding.RandomPointInRects(spawns, rand.Float64); ok {
			g.Pedestrians = append(g.Pedestrians, pedestrian.Pedestrian{X: x, Y: y})
		} else {
			g.Pedestrians = append(g.Pedestrians, pedestrian.Pedestrian{X: rand.Float64() * 700, Y: rand.Float64() * 700})
		}
	}

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
	for _, agt := range g.Agts {
		//ne cree pas de goroutine si agt mort
		if agt.Alive() {
			wg.Add(1)
			go func(agt Agent) {
				defer wg.Done()

				c, _ := g.syncChansAgt.Load(agt.AgtID())
				c.(chan int) <- g.step
				<-c.(chan int)

			}(agt)
		}
	}
	//on attend a ce que toutes les goroutines vivantes donnent leur next step
	wg.Wait()

	// retirer les piétons terminés dont l'attente est expirée ; créer de nouveaux piétons pour maintenir le nombre
	now := time.Now()
	var remaining []pedestrian.Pedestrian
	removed := 0
	for _, p := range g.Pedestrians {
		if p.Done && now.After(p.WaitUntil) {
			removed++
			continue
		}
		remaining = append(remaining, p)
	}
	g.Pedestrians = remaining
	// ré-apparitions pour maintenir la population (garder 5 au total)
	for i := 0; i < removed; i++ {
		// faire apparaître un piéton dans SpawnRects si possible, sinon repli sur la stratégie précédente
		var sx, sy float64
		spawnPlaced := false
		if len(pathfinding.SpawnRects) > 0 {
			for attempt := 0; attempt < 50; attempt++ {
				rx, ry, ok := pathfinding.RandomPointInRects(pathfinding.SpawnRects, rand.Float64)
				if !ok {
					break
				}
				// éviter d'apparaitre trop proche des autres agents / de leurs cibles
				ok2 := true
				for _, q := range g.Pedestrians {
					dx := rx - q.TargetX
					dy := ry - q.TargetY
					if q.TargetX != 0 || q.TargetY != 0 {
						if math.Hypot(dx, dy) < 30 {
							ok2 = false
							break
						}
					}
					dx = rx - q.X
					dy = ry - q.Y
					if math.Hypot(dx, dy) < 25 {
						ok2 = false
						break
					}
				}
				if ok2 {
					sx = rx
					sy = ry
					spawnPlaced = true
					break
				}
			}
			// si après les tentatives le spawn n'a pas été placé, utiliser le centre du premier rectangle de spawn
			if !spawnPlaced {
				r := pathfinding.SpawnRects[0]
				sx = float64(r.X + r.W/2)
				sy = float64(r.Y + r.H/2)
				spawnPlaced = true
			}
		}
		if !spawnPlaced {
			// repli : placement aléatoire avec logique d'évitement
			placed := false
			for attempt := 0; attempt < 30; attempt++ {
				sx = rand.Float64() * 700
				sy = rand.Float64() * 700
				ok := true
				for _, q := range g.Pedestrians {
					dx := sx - q.TargetX
					dy := sy - q.TargetY
					if q.TargetX != 0 || q.TargetY != 0 {
						if math.Hypot(dx, dy) < 30 {
							ok = false
							break
						}
					}
					dx = sx - q.X
					dy = sy - q.Y
					if math.Hypot(dx, dy) < 25 {
						ok = false
						break
					}
				}
				if ok {
					placed = true
					break
				}
			}
			if !placed {
				sx = rand.Float64() * 700
				sy = rand.Float64() * 700
			}
		}
		// ajouter un léger jitter pour que le point d'apparition ne soit pas
		// exactement centré sur une case de la grille (évite des artefacts visuels)
		offset := float64(pathfinding.DefaultCellSize) * 0.25
		sx += (rand.Float64() - 0.5) * offset
		sy += (rand.Float64() - 0.5) * offset
		// s'assurer que la cellule de spawn est franchissable ; si ce n'est pas
		// le cas, tenter quelques ajustements aléatoires autour du point choisi
		// pour éviter d'apparaitre sur une zone interdite.
		if !pathfinding.IsWorldPointWalkable(sx, sy) {
			placed := false
			for attempt := 0; attempt < 20; attempt++ {
				tx := sx + (rand.Float64()-0.5)*float64(pathfinding.DefaultCellSize)
				ty := sy + (rand.Float64()-0.5)*float64(pathfinding.DefaultCellSize)
				if pathfinding.IsWorldPointWalkable(tx, ty) {
					sx = tx
					sy = ty
					placed = true
					break
				}
			}
			if !placed {
				// en dernier recours, utiliser le centre du premier rectangle de
				// spawn (ce point devrait être franchissable car défini par
				// l'auteur de la carte)
				if len(pathfinding.SpawnRects) > 0 {
					r := pathfinding.SpawnRects[0]
					sx = float64(r.X + r.W/2)
					sy = float64(r.Y + r.H/2)
				}
			}
		}

		g.Pedestrians = append(g.Pedestrians, pedestrian.Pedestrian{
			X: sx,
			Y: sy,
		})
	}

	return nil
}

// Affiche les éléments du slice Objts contenant le next step de tout les agents
func (g *Game) Draw(screen *ebiten.Image) {
	////affichage du background habituel
	// background
	screen.Fill(color.RGBA{126, 109, 97, 1.000})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Translate(300, 0)
	screen.DrawImage(imgRoad, op)

	// horizontal road
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
	// stripes = ebiten.NewImage(5, 55)
	// stripes.Fill(color.RGBA{255, 255, 255, 255})
	// for x := 0.0; x < 22; x++ {
	// 	op = &ebiten.DrawImageOptions{}
	// 	op.GeoM.Rotate(math.Pi / 4)
	// 	op.GeoM.Translate(340.0+x*5.0, 300.0+x*5.0)
	// 	screen.DrawImage(stripes, op)
	// }

	//trotoires
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
		if c.Type() == "car" {
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
		}
	}

}

// définit la taille de l'écran du jeu, appelée automatiquement par RunGame
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
