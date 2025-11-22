package game

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	agentcounthandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_count_handler"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	objectslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/object_slice_handler"
	carspawn "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/spawn_points/car_spawn"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// variable pour stocker l'image road.png
var imgRoad *ebiten.Image

// import de road.png
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
	ObjtsCopy []Object // "present state" objects to be drawn
	//slice contenant les objets visuels, c'est la que le prochain step est ecrit
	Objts []Object // "future" objects to be drawn ( se rinitialise a chaque Update )

	//Point de synchronisation agents ( stcoker tous les channels de chaque agents )
	syncChansAgt sync.Map

	//Agents counter, assure des agentID valables
	AgentCount AgentID // compteur d'agent pour donner des ID uniques

	//Channels des differents microservices
	AgentCountCh <-chan AgentID // recevoir le prochain ID d'agent disponible
	AgtsSliceCh  chan<- agentslicehandler.Request
	ObjtsSliceCh chan<- ObjectAddRequest // channel pour ajouter des objets visuels ( seul qui a acces a Objts )

	//savoir combien de fois on est passé par la boucle du jeu
	step int
}

// Init de l'environnement du jeu
func NewGame() *Game {
	g := &Game{}

	//initialisation des différents channel des microservices, lancent ainsi leurs goroutines respectives
	g.AgentCountCh = agentcounthandler.LancerAgentCountHandler(&g.AgentCount)
	g.AgtsSliceCh = agentslicehandler.LancerAgentSliceHandler(&g.Agts, &g.syncChansAgt)
	g.ObjtsSliceCh = objectslicehandler.LancerObjectSliceHandler(&g.Objts)

	//lance la goroutine autonome de creation de voitures, on peut choisir la frequence ici
	carspawner := carspawn.CarSpawn{SpawnInterval: 3 * time.Second} // initialise la structure de spawn de voiture
	carspawner.Start(g.AgentCountCh, g.AgtsSliceCh, g.ObjtsSliceCh, &g.ObjtsCopy)

	return g
}

// appelle les agents pour qu'ils fassent leur prochain mouvement, lancé automatiquement par boucle du jeu
func (g *Game) Update() error {
	fmt.Println("Je viens update :", g.step)
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
				// dire a l'agent de faire son next step grc a uprade + bloquer tant que l'agent n'a pas fini son step
				c, _ := g.syncChansAgt.Load(agt.AgtID())
				c.(chan int) <- g.step // on dit a l'agent de faire son next step grc a uprade
				<-c.(chan int)

			}(agt)
		}
	}
	//on attend a ce que toutes les goroutines vivantes donnent leur next step
	wg.Wait()

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

// sets the size of the game screen, automatically called by RunGame
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 750, 750
}
