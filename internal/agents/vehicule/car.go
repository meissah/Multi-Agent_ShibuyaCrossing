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
	TurnLeft     bool    // indique si la voiture envisage de tourner à gauche
	IsTurning    bool    // en train de tourner (logique interne)
	CurrentAngle float64 // angle de rotation actuel pendant le virage
	// en gros va permettre de modifier X et Y en utilisant cos et sin pour faire une belle courbe

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
// pour l'instant on ne gère que le virage à gauche en venant du sud
const (
	SOUTH_LEFT_TURN_X = 325.0
	SOUTH_LEFT_TURN_Y = 440.0
	TURN_RADIUS       = 15.0         // rayon du cercle de virage ( à modif selon la lane utilisée (a faire a plus tard ))
	TURN_SPEED        = math.Pi / 24 // π/24 radians par step ( en gros on veut tourner de 90 degre et donc on tourne de π/24 a chaque step donc en 24 steps on a fait un quart de tour 90 degres -> permet d'avoir un virage fluide)
	// virage normale 24, lent 48
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

		// pour le virage à gauche
		TurnLeft:            turnleft,
		DesiredStopDistance: 50.0,
		IsTurning:           false, // reste false si turnleft est false
		CurrentAngle:        0.0,

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

// / Deliberate: la voiture prend des décisions basées sur ses perceptions
func (c *Car) Deliberate() {
	// Décider de commencer à tourner si on atteint le point de virage
	if c.TurnLeft && !c.IsTurning && c.Dir == North && c.LaneIndex == 0 {
		// On vérifie si elle ATTEINT le point de virage
		if c.Y <= SOUTH_LEFT_TURN_Y+0.1 && c.Y >= SOUTH_LEFT_TURN_Y-0.1 {
			c.IsTurning = true
			c.CurrentAngle = 3 * math.Pi / 2 // objectif c'est d'arriver a cette angle en tournant a gauche

			// Initialiser la position sur le cercle de virage

			centerX := SOUTH_LEFT_TURN_X - TURN_RADIUS
			centerY := SOUTH_LEFT_TURN_Y - TURN_RADIUS
			c.X = centerX + TURN_RADIUS*math.Cos(c.CurrentAngle)
			c.Y = centerY + TURN_RADIUS*math.Sin(c.CurrentAngle)
		}
	}

}

// Act: exécute le mouvement
func (c *Car) Act(ActChn chan<- ObjectAddRequest) {
	// Exécuter le virage si IsTurning == true
	if c.IsTurning {
		// Tourner
		c.CurrentAngle -= TURN_SPEED // Diminue l'angle à chaque step

		// Centre du cercle de virage
		centerX := SOUTH_LEFT_TURN_X - TURN_RADIUS
		centerY := SOUTH_LEFT_TURN_Y - TURN_RADIUS

		// Calculer nouvelle position sur le cercle
		c.X = centerX + TURN_RADIUS*math.Cos(c.CurrentAngle)
		c.Y = centerY + TURN_RADIUS*math.Sin(c.CurrentAngle)

		// Fin du virage
		if c.CurrentAngle <= math.Pi {
			c.IsTurning = false
			c.Dir = West
			c.CurrentAngle = 0.0 // Remettre l'angle à 0 après le virage
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
	if c.IsTurning {
		// L'angle visuel est perpendiculaire au rayon (tangent au cercle)
		// On ajoute -π/2 pour que la voiture pointe dans la direction du mouvement
		println("Angle during turn:", c.CurrentAngle-math.Pi/2)
		return c.CurrentAngle - math.Pi/2 // Tangente au cercle
	}
	return 0.0
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
