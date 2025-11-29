package vehicule

import (
	"math"
	"time"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

func (ag Car) AgtID() AgentID {
	return ag.ID
}

// lance la goroutine de l'agent car
func (ag *Car) Start(env *[]Object, ObjectCh chan<- ObjectAddRequest) {

	go func() {
		var step int
		for {
			step = <-ag.syncChan // update met valeur dans le synchan ( bloquant tant que le update n'a pas fait son step )

			ag.Percept(env)
			ag.Deliberate()
			ag.Act(ObjectCh)
			//fmt.Println("J'ai avancé ", ag.ID)
			step++
			ag.syncChan <- step
		}
	}()

}

// jouera sur les distances d'arrêt et la vitesse pour chaque conducteur
type DriverProfile struct {
	Aggressiveness float64 // 0.0 (prudent) à 1.0 (agressif)
}

type Car struct {
	ID        AgentID
	X, Y      float64
	Dir       Direction
	LaneIndex int // 0 ou 1 pour la voie utilisée (deux voies par direction)

	// Pour le virage
	TurnLeft   bool
	IsTurning  bool
	TurnStep   float64 // compteur de 0 à TURN_DURATION
	TurnStartX float64 // position de départ du virage
	TurnStartY float64
	AngleRad   float64

	Speed        float64 // vitesse actuelle en px/sec
	MaxSpeed     float64 // Vitesse maximale souhaitée en px/sec
	DrivingState string  // "moving","stopped"

	lastStep time.Time

	// Personnalité du conducteur
	Profile             DriverProfile
	DesiredStopDistance float64

	// Intention (décidé par Deliberate)
	intention string // "accelerate", "stop", "slow_down"

	AmIAlive bool

	// Référence à l'environnement pour les perceptions
	//env      CarEnvironment
	syncChan chan int
}

// Update envoyé au handler pour mise à jour de l'état de la voiture
type CarUpdate struct {
	ID     AgentID
	X, Y   float64
	Remove bool // true si la voiture doit être supprimée ( hors écran par exemple )
}

// Constantes pour le virage
const (
	SOUTH_LEFT_TURN_X = 325.0
	SOUTH_LEFT_TURN_Y = 440.0
	TURN_DURATION     = 30.0 // nombre de steps pour faire le virage
)

// Interface environnement pour la voiture ( on mettra plus tard les feux )
type CarEnvironment interface {
	GetCars() []*Car // Retourne toutes les voitures dans l'environnement
}

// NewCar constructeur
func NewCar(id AgentID, x, y float64, dir Direction, lane int, turnleft bool, syncChan chan int) *Car {
	return &Car{
		ID:        id,
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
		AngleRad:   0.0,

		DesiredStopDistance: 50.0,

		Speed:        0,
		MaxSpeed:     100.0,    // pixels/sec j'ai mis au hasard a ajuster selon les tests
		DrivingState: "moving", // par défaut en mouvement mais a voir parce que si trop de voiture tecnhiquemment elle devrait s'arreter
		AmIAlive:     true,
		//updateCh:              updateCh,
		lastStep: time.Now(),
		Profile:  DriverProfile{Aggressiveness: 0.5}, // profil normale par défaut

		intention: "accelerate", // pareil que pour le State moving à voir plus tard
		syncChan:  syncChan,
	}
}

// GetID getter ID de la voiture
func (c *Car) GetID() AgentID {
	return c.ID
}

// Percept: perçoit l'environnement (autres voitures et feux ( pas encore fait ))
// La voiture regarde autour d'elle et détecte naturellement les dangers
func (c *Car) Percept(env *[]Object) {

}

func (c *Car) Deliberate() {
	if c.TurnLeft && !c.IsTurning && c.Dir == North && c.LaneIndex == 0 {
		if c.Y <= SOUTH_LEFT_TURN_Y+5 && c.Y >= SOUTH_LEFT_TURN_Y-5 {
			c.IsTurning = true
			c.TurnStep = 0
			c.TurnStartX = c.X // Sauvegarde position actuelle
			c.TurnStartY = c.Y
		}
	}
}

// Act: exécute le mouvement
func (c *Car) Act(ActChn chan<- ObjectAddRequest) {
	if c.IsTurning {
		c.TurnStep++
		t := c.TurnStep / TURN_DURATION // progression 0→1

		// Calcul de la position selon une courbe de Bézier cubique ( sur 4 points )
		// Points de contrôle de la courbe
		p0x, p0y := c.TurnStartX, c.TurnStartY       // point de départ virage
		p1x, p1y := c.TurnStartX, c.TurnStartY-25    // direction depart
		p2x, p2y := c.TurnStartX-25, c.TurnStartY-40 // direction arrivée
		p3x, p3y := c.TurnStartX-50, c.TurnStartY-50 // point d'arrivée virage

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

		angle := math.Atan2(c.Y, c.X)
		c.AngleRad = angle

		if c.TurnStep >= TURN_DURATION {
			c.IsTurning = false
			c.Dir = West
			c.AngleRad = 0.0
		}

	} else {
		// Mouvement normal en ligne droite
		switch c.Dir {
		case North:
			c.Y -= 2.0 // LA VOITURE MONTE
		case South:
			c.Y += 2.0 // pour le test
		case East:
			c.X += 2.0
		case West:
			c.X -= 2.0
		}
	}

	// Envoyer la position
	RespChan := make(chan int)
	ActChn <- ObjectAddRequest{
		Object:   c,
		RespChan: RespChan,
	}
	<-RespChan
}

func (c *Car) Position() (float64, float64) {
	return c.X, c.Y
}

func (c *Car) TurningLeft() bool {
	return c.IsTurning
}

// Angle pour l'affichage

func (c *Car) Angle() float64 {
	return c.AngleRad
}

func (c *Car) Type() string {
	return "car"
}

func (c *Car) State() string {
	return c.DrivingState
}

func (c *Car) Direction() Direction {
	return c.Dir
}

func (c *Car) Alive() bool {
	return c.AmIAlive
}
