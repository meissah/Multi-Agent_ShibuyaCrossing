package vehicule

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

type Bounds struct {
	dX, dY float64
}

var carAllImgBounds = []Bounds{
	{dX: 72, dY: 114},
	{dX: 73, dY: 121},
	{dX: 66, dY: 114},
	{dX: 70, dY: 116},
	{dX: 72, dY: 117},
}

func (b Bounds) Bounds() (float64, float64) {
	return b.dX, b.dY
}

func (ag *Car) AgentID() AgentID {
	return ag.ID
}

func (ag *Car) ObjectID() ObjectID {
	return ObjectID(ag.ID)
}

func (ag *Car) EndChan() chan int {
	return ag.endChan
}

// lance la goroutine de l'agent car
func (ag *Car) Start(env <-chan []Object) {

	go func() {
		var step int
		for {
			select {
			case step = <-ag.syncChan:

				ag.Percept(<-env)

				ag.syncChan <- step
				<-ag.syncChan

				ag.Deliberate()
				ag.Act()

				ag.syncChan <- step
			case <-ag.endChan:
				close(ag.endChan)
				close(ag.syncChan)
				return
			}
		}
	}()

}

type Car struct {
	ID        AgentID
	X, Y      float64
	Dir       Direction
	Angle     float64
	LaneIndex int // 0 ou 1 pour la voie utilisée (deux voies par direction)
	wtvthisis float64

	// Pour le virage
	TurnLeft   bool
	IsTurning  bool
	TurnStep   float64 // compteur de 0 à TURN_DURATION
	TurnStartX float64 // position de départ du virage
	TurnStartY float64

	Speed        float64 // vitesse actuelle en px/sec
	MaxSpeed     float64 // Vitesse maximale souhaitée en px/sec
	DrivingState string  // "moving","stopped"

	AmIAlive bool

	// Couleur de la voiture (index dans le tableau d'images)
	// 0=blue, 1=red, 2=green, 3=lightblue, 4=yellow ...
	carColorIndex int

	// Référence à l'environnement pour les perceptions
	//env      CarEnvironment
	syncChan chan int
	endChan  chan int
}

// Constantes pour le virage
const (
	SOUTH_LEFT_TURN_X = 325.0
	SOUTH_LEFT_TURN_Y = 440.0

	NORTH_LEFT_TURN_X = 420.0
	NORTH_LEFT_TURN_Y = 307.0

	EAST_LEFT_TURN_X = 440.0
	EAST_LEFT_TURN_Y = 325.0

	WEST_LEFT_TURN_X = 307.0
	WEST_LEFT_TURN_Y = 420.0

	TURN_DURATION = 30.0 // nombre de steps pour faire le virage
)

// NewCar constructeur
func NewCar(id ObjectID, x, y float64, dir Direction, lane int, turnleft bool, syncChan chan int) *Car {
	return &Car{
		ID:        AgentID(id),
		X:         x,
		Y:         y,
		Dir:       dir,
		LaneIndex: lane,

		// virage
		TurnLeft:   turnleft,
		IsTurning:  false,
		TurnStep:   0.0, // compteur de 0 à TURN_DURATION
		TurnStartX: 0.0, // position de départ du virage
		TurnStartY: 0.0,

		MaxSpeed:      100.0, // pixels/sec j'ai mis au hasard a ajuster selon les tests
		AmIAlive:      true,
		syncChan:      syncChan,
		endChan:       make(chan int),
		carColorIndex: rand.Intn(5),
	}
}

func (c *Car) Percept(env []Object) {
	for _, obj := range env {
		c.wtvthisis = obj.PositionX()
		//fmt.Println("Pedestrian ", c.ID, " perceives object: ", obj.Type(), " at (", obj.PositionX(), ",", obj.PositionX(), ")")
		// Le piéton peut percevoir les objets autour de lui (voitures, feux, autres piétons)
		// mais pour l'instant on n'utilise pas cette information dans le comportement
	}
}

func (c *Car) Deliberate() {
	if c.TurnLeft && !c.IsTurning {
		// Virage pour les voitures venant du Sud (direction North)
		if c.Dir == North {
			if c.Y <= SOUTH_LEFT_TURN_Y+5 && c.Y >= SOUTH_LEFT_TURN_Y-5 {
				c.IsTurning = true
				c.TurnStep = 0
				c.TurnStartX = c.X
				c.TurnStartY = c.Y
			}
		}
		// Virage pour les voitures venant du Nord (direction South)
		if c.Dir == South {
			if c.Y >= NORTH_LEFT_TURN_Y-5 && c.Y <= NORTH_LEFT_TURN_Y+5 {
				c.IsTurning = true
				c.TurnStep = 0
				c.TurnStartX = c.X
				c.TurnStartY = c.Y
			}
		}
		// Virage pour les voitures venant de l'Est (direction West)
		if c.Dir == West {
			if c.X <= EAST_LEFT_TURN_X+5 && c.X >= EAST_LEFT_TURN_X-5 {
				c.IsTurning = true
				c.TurnStep = 0
				c.TurnStartX = c.X
				c.TurnStartY = c.Y
			}
		}
		// Virage pour les voitures venant de l'Ouest (direction East)
		if c.Dir == East {
			if c.X >= WEST_LEFT_TURN_X-5 && c.X <= WEST_LEFT_TURN_X+5 {
				c.IsTurning = true
				c.TurnStep = 0
				c.TurnStartX = c.X
				c.TurnStartY = c.Y
			}
		}
	}

	if c.IsTurning {
		c.TurnStep++
		t := c.TurnStep / TURN_DURATION

		var p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y float64
		// Ajustement des points de contrôle selon la direction initiale
		if c.Dir == North { // spawn depuis le sud et tourne a gauche vers l'ouest
			if c.LaneIndex == 0 {
				// Lane de gauche
				p0x, p0y = c.TurnStartX, c.TurnStartY       // point de départ virage
				p1x, p1y = c.TurnStartX, c.TurnStartY-5     // direction depart
				p2x, p2y = c.TurnStartX-15, c.TurnStartY-12 // direction arrivée
				p3x, p3y = c.TurnStartX-25, c.TurnStartY-22 // point d'arrivée virage
			} else {
				// Lane de droite
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX, c.TurnStartY-30
				p2x, p2y = c.TurnStartX-25, c.TurnStartY-40
				p3x, p3y = c.TurnStartX-45, c.TurnStartY-50
			}
		} else if c.Dir == South { // spawn depuis le nord et tourne a gauche vers l'est
			if c.LaneIndex == 0 {
				// Lane de gauche
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX, c.TurnStartY+30
				p2x, p2y = c.TurnStartX+25, c.TurnStartY+40
				p3x, p3y = c.TurnStartX+45, c.TurnStartY+53

			} else {
				// Lane de droite
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX, c.TurnStartY+5
				p2x, p2y = c.TurnStartX+15, c.TurnStartY+12
				p3x, p3y = c.TurnStartX+25, c.TurnStartY+22

			}
		} else if c.Dir == West { // spawn depuis l'est et tourne a gauche vers le sud
			if c.LaneIndex == 0 {
				// Lane de gauche
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX-30, c.TurnStartY
				p2x, p2y = c.TurnStartX-40, c.TurnStartY+25
				p3x, p3y = c.TurnStartX-50, c.TurnStartY+45
			} else {
				// Lane de droite
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX-5, c.TurnStartY
				p2x, p2y = c.TurnStartX-12, c.TurnStartY+15
				p3x, p3y = c.TurnStartX-22, c.TurnStartY+25
			}
		} else if c.Dir == East { // spawn depuis l'ouest et tourne a gauche vers le nord
			if c.LaneIndex == 0 {
				// Lane de gauche
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX+5, c.TurnStartY
				p2x, p2y = c.TurnStartX+12, c.TurnStartY-15
				p3x, p3y = c.TurnStartX+22, c.TurnStartY-25
			} else {
				// Lane de droite
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX+30, c.TurnStartY
				p2x, p2y = c.TurnStartX+40, c.TurnStartY-25
				p3x, p3y = c.TurnStartX+53, c.TurnStartY-45
			}
		}

		// Formule de Bézier cubique
		u := 1 - t
		c.X = u*u*u*p0x + 3*u*u*t*p1x + 3*u*t*t*p2x + t*t*t*p3x
		c.Y = u*u*u*p0y + 3*u*u*t*p1y + 3*u*t*t*p2y + t*t*t*p3y

		/*
		   B(t) = (1-t)³·P0 + 3(1-t)²t·P1 + 3(1-t)t²·P2 + t³·P3 // formule bézier cubique
		   c.X = u*u*u*p0x + 3*u*u*t*p1x + 3*u*t*t*p2x + t*t*t*p3x
		         └─────┘     └───────┘     └───────┘     └─────┘
		         (1-t)³·P0   3(1-t)²t·P1   3(1-t)t²·P2   t³·P3

		   // Position Y sur la courbe (même formule)
		   c.Y = u*u*u*p0y + 3*u*u*t*p1y + 3*u*t*t*p2y + t*t*t*p3y
		*/

		// Fin du virage
		if c.TurnStep >= TURN_DURATION {
			c.IsTurning = false
			c.Angle = 0.0
			if c.Dir == North {
				c.Dir = West
			} else if c.Dir == South {
				c.Dir = East
			} else if c.Dir == West {
				c.Dir = South
			} else if c.Dir == East {
				c.Dir = North
			}
		}

	} else {
		// Mouvement normal en ligne droite
		switch c.Dir {
		case North:
			c.Y -= 2.0 // la voiture monte
		case South:
			c.Y += 2.0
		case East:
			c.X += 2.0
		case West:
			c.X -= 2.0
		}
	}

	if c.IsTurning {
		t := c.TurnStep / TURN_DURATION

		if c.Dir == North {
			// Sud vers Ouest : rotation de 0 à -π/2
			c.Angle = -t * (math.Pi / 2)
		} else if c.Dir == South {
			// Nord vers Est : rotation de π à π/2
			c.Angle = math.Pi - t*(math.Pi/2)
		} else if c.Dir == West {
			// Est vers Sud : rotation de -π/2 à -π
			c.Angle = -math.Pi - t*(math.Pi/2)
		} else if c.Dir == East {
			// Ouest vers Nord : rotation de π/2 à 0
			c.Angle = -t * (math.Pi / 2)
		}
	}

	if c.X < -50 || c.X > 800 || c.Y < -50 || c.Y > 800 {
		c.AmIAlive = false
	}
}

func (c *Car) Act() {
}

func (c *Car) Type() string {
	return "car"
}

func (c *Car) Direction() Direction {
	return c.Dir
}

func (c *Car) TheAngle() float64 {
	return c.Angle
}

func (c *Car) PositionX() float64 {
	return c.X
}

func (c *Car) PositionY() float64 {
	return c.Y
}

func (c *Car) Draw() (Asset, *ebiten.DrawImageOptions) {

	// Récupérer l'image de la voiture selon sa couleur
	colorIndex := c.CarColorIndex()
	carImgBounds := carAllImgBounds[colorIndex]
	imgW, imgH := carImgBounds.Bounds()

	op := &ebiten.DrawImageOptions{}
	var w, h int
	var scaleX, scaleY float64

	// Appliquer la rotation de base selon la direction
	switch c.Direction() {
	case South:
		w, h = 15, 25
		op.GeoM.Rotate(math.Pi)
		op.GeoM.Translate(imgW, imgH)
		scaleX = float64(w) / imgW
		scaleY = float64(h) / imgH
	case North:
		w, h = 15, 25
		scaleX = float64(w) / imgW
		scaleY = float64(h) / imgH
	case West:
		w, h = 25, 15
		op.GeoM.Rotate((3 * math.Pi) / 2)
		op.GeoM.Translate(0, imgW)
		scaleY = float64(h) / imgW
		scaleX = float64(w) / imgH
	case East:
		w, h = 25, 15
		op.GeoM.Rotate(math.Pi / 2)
		op.GeoM.Translate(imgH, 0)
		scaleY = float64(h) / imgW
		scaleX = float64(w) / imgH
	}

	// Appliquer le scale et centrer
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	// Récupérer la position
	x, y := c.PositionX(), c.PositionY()

	// Appliquer la rotation du virage si nécessaire
	if c.Angle != 0 {
		if c.Direction() == West || c.Direction() == South {
			op.GeoM.Rotate(c.Angle + math.Pi)
		} else {
			op.GeoM.Rotate(c.Angle)
		}
	}

	// Positionner à la position finale
	op.GeoM.Translate(x, y)

	return Asset(c.carColorIndex + 1), op
}

func (c *Car) Position() (float64, float64) {
	return c.X, c.Y
}

func (c *Car) TurningLeft() bool {
	return c.IsTurning
}

// CarColorIndex retourne l'index de la couleur de la voiture (pour l'affichage)
func (c *Car) CarColorIndex() int {
	return c.carColorIndex
}

func (c *Car) Alive() bool {
	return c.AmIAlive
}
