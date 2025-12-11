// Ce fichier définit la structure Game et son initialisation.
// Il gère les états du jeu (menu/simulation) et la boucle principale.
package game

import (
	"fmt"
	"sync"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
	agentcounthandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_count_handler"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	objectslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/object_slice_handler"
	carspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/car_spawn"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Game représente l'état complet de l'application
type Game struct {
	// État actuel (menu ou simulation)
	State  config.GameState        // État actuel du jeu si menu ou simulation
	Config config.SimulationConfig // Configuration de la simulation choisie dans le menu par l'utilisateur

	// Éléments du menu de configuration
	selectedCrossing string // "shibuya" ou "french"
	carSpawnInput    float64
	pedSpawnInput    float64
	// mettre plus tard: durée des feux de circulation

	// Composants UI du menu
	validateButton ui.Button
	shibuyaRect    ui.Rectangle
	frenchRect     ui.Rectangle
	carSpawnSlider ui.Slider
	pedSpawnSlider ui.Slider

	// Agents et objets de la simulation
	Agts         []Agent
	ObjtsCopy    []Object
	Objts        []Object
	syncChansAgt sync.Map
	AgentCount   AgentID

	//Channels des differents microservices
	AddAgtSlice     chan<- agentslicehandler.AddAgtRequest
	ReceiveAgtSlice chan []Agent
	DelAgtSlice     chan<- Agent
	AgentCountCh    <-chan AgentID          // recevoir le prochain ID d'agent disponible
	ObjtsSliceCh    chan<- ObjectAddRequest // channel pour ajouter des objets visuels ( seul qui a acces a Objts )

	step       int
	carspawner *carspawn.CarSpawn
}

// NewGame crée et initialise une nouvelle instance du jeu en mode menu

func NewGame() *Game {
	g := &Game{
		State:          config.StateMenu,       // commencer dans le menu
		Config:         config.DefaultConfig(), // configuration par défaut si l'utilisateur ne choisit rien
		carSpawnInput:  1.0,
		pedSpawnInput:  0.5,
		shibuyaRect:    ui.Rectangle{X: 200, Y: 300, W: 150, H: 150},
		frenchRect:     ui.Rectangle{X: 400, Y: 300, W: 150, H: 150},
		carSpawnSlider: ui.NewSlider(250, 165, 250, 15, 1.0, 50.0, 1.0), // Min=1, Max=50, aligné après le label
		pedSpawnSlider: ui.NewSlider(250, 205, 250, 15, 1.0, 50.0, 5.0), // Min=1, Max=50, aligné après le label
	}

	g.validateButton = ui.NewButton(600, 650, 120, 40, "Valider")

	fmt.Printf("Slider voitures: W=%.1f, H=%.1f\n", g.carSpawnSlider.Width, g.carSpawnSlider.Height)
	fmt.Printf("Slider piétons: W=%.1f, H=%.1f\n", g.pedSpawnSlider.Width, g.pedSpawnSlider.Height)

	return g
}

// InitSimulation initialise tous les handlers et spawners pour démarrer la simulation
func (g *Game) InitSimulation() {
	// Initialiser les channels des microservices
	g.AgentCountCh = agentcounthandler.LancerAgentCountHandler(&g.AgentCount)
	g.AddAgtSlice, g.ReceiveAgtSlice, g.DelAgtSlice = agentslicehandler.LancerAgentSliceHandler(&g.Agts, &g.syncChansAgt)
	g.ObjtsSliceCh = objectslicehandler.LancerObjectSliceHandler(&g.Objts)

	// Lancer le spawner de voitures avec la configuration choisie
	g.carspawner = &carspawn.CarSpawn{SpawnInterval: g.Config.CarSpawnRate} // CarSpawnRate c'est l'intervalle entre chaque spawn rentre par l'utilisateur
	g.carspawner.Start(g.AgentCountCh, g.AddAgtSlice, g.ObjtsSliceCh, &g.ObjtsCopy)

	// A faire -> Lancer le spawner de piétons quand implémenté
	// pedestrianspawner := &pedestrianspawn.PedestrianSpawn{SpawnInterval: g.Config.PedestrianSpawnRate}
	// pedestrianspawner.Start(...)

	fmt.Printf("Simulation démarrée:\n  - Type: %s\n  - Voitures: %.1f/s\n  - Piétons: %.1f/s\n",
		g.Config.CrossingType,
		1.0/g.Config.CarSpawnRate.Seconds(),
		1.0/g.Config.PedestrianSpawnRate.Seconds())
}

// Layout définit la taille de la fenêtre du jeu (appelé automatiquement par Ebiten)
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}

/*
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

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"
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

// Liste des images de voitures (une par couleur)
var carImages []*ebiten.Image

// Liste des noms de couleurs disponibles
var carColorNames []string

// import de road.png
func init() {
	var err error
	imgRoad, _, err = ebitenutil.NewImageFromFile("internal/assets/road.png")
	if err != nil {
		log.Fatal(err)
	}

	// Liste des couleurs de voitures disponibles
	carColorNames = []string{"blue", "red", "green", "purple", "lightblue", "yellow"}

	// On charge chaque image de voiture
	for _, color := range carColorNames {
		filename := fmt.Sprintf("internal/assets/cars/car_%s.png", color)
		img, _, err := ebitenutil.NewImageFromFile(filename)
		if err != nil {
			// Si une image ne charge pas, on arrête le programme
			log.Fatalf("erreur: Impossible de charger %s: %v", filename, err)
		}
		// On ajoute l'image chargée à notre liste
		carImages = append(carImages, img)
		log.Printf("Image chargée: %s", filename)
	}

}

// Structure de l'environment du jeu
type Game struct {
	//slice contenant l'ensemble des agents
	Agts []Agent
	//slice contenant une copy des objets visuels, pour eviter la triche sur le next step par les agents
	ObjtsCopy []Object // "present state" objects to be drawn
	//slice contenant les objets visuels, c'est la que le prochain step est ecrit
	Objts []Object // "future" objects to be drawn ( se rinitialise a chaque Update )

	//Point de synchronisation agents ( stcoker tous les channels de chaque agents )
	syncChansAgt sync.Map

	//Agents counter, assure des agentID valables
	AgentCount AgentID // compteur d'agent pour donner des ID uniques

	//Channels des differents microservices
	AgentCountCh    <-chan AgentID
	AddAgtSlice     chan<- agentslicehandler.AddAgtRequest
	ReceiveAgtSlice chan []Agent
	DelAgtSlice     chan<- Agent
	ObjtsSliceCh    chan<- ObjectAddRequest
	AgentCountCh <-chan AgentID // recevoir le prochain ID d'agent disponible
	AgtsSliceCh  chan<- agentslicehandler.Request
	ObjtsSliceCh chan<- ObjectAddRequest // channel pour ajouter des objets visuels ( seul qui a acces a Objts )

	//savoir combien de fois on est passé par la boucle du jeu
	step int

	//pedestrians
	Pedestrians []pedestrian.Pedestrian

	updateCount int
	drawCount   int
	perSec      time.Time
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
	carspawner := carspawn.CarSpawn{SpawnInterval: 1 * time.Second} // initialise la structure de spawn de voiture
	carspawner.Start(g.AgentCountCh, g.AgtsSliceCh, g.ObjtsSliceCh, &g.ObjtsCopy)

	return g
}

// appelle les agents pour qu'ils fassent leur prochain mouvement, lancé automatiquement par boucle du jeu
func (g *Game) Update() error {
	//fmt.Println("Je viens update :", g.step)
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
				// dire a l'agent de faire son next step grc a uprade + bloquer tant que l'agent n'a pas fini son step
				c, _ := g.syncChansAgt.Load(agt.AgtID())
				c.(chan int) <- g.step // on dit a l'agent de faire son next step grc a uprade
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
	now := time.Now()
	g.updateCount++

	// Debug print per sec
	if now.Sub(g.perSec) >= time.Second {
		log.Printf("TPS: %.2f, FPS: %.2f", ebiten.ActualTPS(), ebiten.ActualFPS())

		g.updateCount = 0
		g.drawCount = 0
		g.perSec = now
	}

	//on copie l'etat précédent des objets visuels
	g.ObjtsCopy = append([]Object(nil), g.Objts...)
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
	for _, c := range g.ObjtsCopy { //potentiellemtn probleme de concurrence avec update(), nouveau microservice a faire?
		//dessine les voitures, a remplacer par select/case plutard pour prendre en compte l'ensemble des types d'objets
		switch c.Type() {
		case "car":
	// ////Dessine les voitures avec leurs images
	// for _, obj := range g.Objts {
	// 	if obj.Type() == "car" {
	// 		// Type assertion : convertir Object en *vehicule.Car
	// 		car, ok := obj.(*vehicule.Car)
	// 		if !ok {
	// 			continue
	// 		}

	// 		// Récupérer l'index de couleur de la voiture
	// 		colorIndex := car.CarColorIndex()
	// 		carImg := carImages[colorIndex]

	// 		// Récupérer les dimensions réelles de l'image
	// 		bounds := carImg.Bounds()
	// 		imgW, imgH := float64(bounds.Dx()), float64(bounds.Dy())

	// 		op := &ebiten.DrawImageOptions{}
	// 		var w, h float64 = 35, 50

	// 		// Calculer le scale pour avoir EXACTEMENT w x h pixels
	// 		scaleX := w / imgW
	// 		scaleY := h / imgH

	// 		// Scale pour obtenir EXACTEMENT les dimensions w x h
	// 		op.GeoM.Scale(scaleX, scaleY)

	// 		// Centrer à l'origine
	// 		op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	// 		// Centrer l'image à l'origine (comme pour le rectangle)
	// 		//op.GeoM.Translate(-imgW/2, -imgH/2)

	// 		// // // Dimensions EXACTEMENT comme les rectangles (23x33 ou 33x23)
	// 		// // var w, h float64
	// 		// // switch car.Direction() {
	// 		// // case North, South:
	// 		// // 	w, h = 35, 50 // Vertical
	// 		// // case East, West:
	// 		// // 	w, h = 50, 35 // Horizontal
	// 		// // }

	// 		// // Préparer le dessin - MÊME LOGIQUE que les rectangles

	// 		// x, y := car.Position()
	// 		angle := car.Angle()

	// 		// // Appliquer la rotation (comme pour le rectangle)
	// 		op.GeoM.Rotate(angle)

	// 		// // Déplacer à la position finale (comme pour le rectangle)
	// 		// op.GeoM.Translate(x, y)

	// 		// Dessiner l'image de la voiture
	// 		screen.DrawImage(carImg, op)
	// 	}
	// }

	////Dessine les objets dans le slice Objts ( ancienne version avec rectangles basiques pour les voitures )

	for _, c := range g.Objts {
		//dessine les voitures, a remplacer par select/case plutard pour prendre en compte l'ensemble des types d'objets
		if c.Type() == "car" {
			// Type assertion : convertir Object en *vehicule.Car
			car, ok := c.(*vehicule.Car)
			if !ok {
				continue
			}

			// Récupérer l'index de couleur de la voiture
			colorIndex := car.CarColorIndex()
			carImg := carImages[colorIndex]

			// Récupérer les dimensions réelles de l'image
			bounds := carImg.Bounds()
			imgW, imgH := float64(bounds.Dx()), float64(bounds.Dy())
			//fmt.Println(imgW, imgH)

			op := &ebiten.DrawImageOptions{}
			// Dimensions selon la direction
			var w, h int
			var scaleX, scaleY float64
			fmt.Println(w, h, scaleX, scaleY, imgH)

			angle := c.Angle() // on recupere l'angle de la voiture

			switch c.Direction() {
			case South:
				w, h = 35, 50 // pour que les voitures ont la bonne forme suivant leur apparition
				// if angle == 0 {
				op.GeoM.Rotate((math.Pi))
				op.GeoM.Translate(imgW, imgH)
				// }
				//Calculer le scale pour avoir EXACTEMENT w x h pixels
				scaleX = float64(w) / imgW
				scaleY = float64(h) / imgH
			case North:
				w, h = 35, 50 // pour que les voitures ont la bonne forme suivant leur apparition
				//Calculer le scale pour avoir EXACTEMENT w x h pixels
				scaleX = float64(w) / imgW
				scaleY = float64(h) / imgH

			case West:
				w, h = 50, 35
				// if angle != 0 {
				// 	op.GeoM.Rotate((math.Pi) / 2)
				// 	op.GeoM.Translate(imgH, 0)
				// } else {
				op.GeoM.Rotate((3 * math.Pi) / 2)
				op.GeoM.Translate(0, imgW)
				//}
				// //Calculer le scale pour avoir EXACTEMENT w x h pixels
				scaleY = float64(h) / imgW
				scaleX = float64(w) / imgH

			case East:
				w, h = 50, 35
				op.GeoM.Rotate((math.Pi) / 2)
				op.GeoM.Translate(imgH, 0)
				// //Calculer le scale pour avoir EXACTEMENT w x h pixels
				scaleY = float64(h) / imgW
				scaleX = float64(w) / imgH
			}

			// // Scale pour obtenir EXACTEMENT les dimensions w x h
			op.GeoM.Scale(scaleX, scaleY)
			// Centrer à l'origine
			op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

			//op.GeoM.Translate(50, 50)

			// Crée un rectangle pour représenter la voiture
			rect := ebiten.NewImage(10, 10)

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
	g.drawCount++
			//Calculer le scale pour avoir EXACTEMENT w x h pixels
			// scaleX := float64(w) / imgW
			// scaleY := float64(h) / imgH

			x, y := c.Position()

			//Appliquer la rotation
			if angle != 0 && c.Direction() == West {
				op.GeoM.Rotate(angle + math.Pi)
			} else if angle != 0 && c.Direction() == South {
				op.GeoM.Rotate(angle + math.Pi)
			} else {
				op.GeoM.Rotate(angle)
			}

			op.GeoM.Translate(x, y)

			screen.DrawImage(carImg, op)
			//screen.DrawImage(rect, op)
		}
	}
}

// définit la taille de l'écran du jeu, appelée automatiquement par RunGame
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}

*/
