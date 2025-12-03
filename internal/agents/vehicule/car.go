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

	NORTH_LEFT_TURN_X = 420.0
	NORTH_LEFT_TURN_Y = 307.0

	EAST_LEFT_TURN_X = 440.0
	EAST_LEFT_TURN_Y = 325.0

	WEST_LEFT_TURN_X = 307.0
	WEST_LEFT_TURN_Y = 420.0

	TURN_DURATION = 30.0 // nombre de steps pour faire le virage
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

/*
	func (c *Car) Deliberate() {
		if c.TurnLeft && !c.IsTurning && c.Dir == North {
			if c.Y <= SOUTH_LEFT_TURN_Y+5 && c.Y >= SOUTH_LEFT_TURN_Y-5 {
				c.IsTurning = true
				c.TurnStep = 0
				c.TurnStartX = c.X // Sauvegarde position actuelle
				c.TurnStartY = c.Y
			}
		}
	}
*/

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
}

func (c *Car) Act(ActChn chan<- ObjectAddRequest) {
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
			if c.Dir == North {
				c.Dir = West
			} else if c.Dir == South {
				c.Dir = East
			} else if c.Dir == West {
				c.Dir = South
			} else if c.Dir == East {
				c.Dir = North
			}
			c.AngleRad = 0.0
		}

	} else {
		// Mouvement normal en ligne droite
		switch c.Dir {
		case North:
			c.Y -= 3.0 // LA VOITURE MONTE
		case South:
			c.Y += 3.0 // pour le test
		case East:
			c.X += 3.0
		case West:
			c.X -= 3.0
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

// Angle pour l'affichage, quand je faisais que tourner en partant du sud
/*
func (c *Car) Angle() float64 {
	if c.IsTurning {
		t := c.TurnStep / TURN_DURATION

		if c.Dir == North {
			// Sud vers Ouest : rotation de 0 à -π/2
			return -t * (math.Pi / 2)
		} else if c.Dir == South {
			// Nord vers Est : rotation de π à π/2
			return math.Pi - t*(math.Pi/2)
		}
	}
	return 0
}
*/

func (c *Car) Angle() float64 {
	if c.IsTurning {
		t := c.TurnStep / TURN_DURATION

		if c.Dir == North {
			// Sud vers Ouest : rotation de 0 à -π/2
			return -t * (math.Pi / 2)
		} else if c.Dir == South {
			// Nord vers Est : rotation de π à π/2
			return math.Pi - t*(math.Pi/2)
		} else if c.Dir == West {
			// Est vers Sud : rotation de -π/2 à -π
			return -math.Pi - t*(math.Pi/2)
		} else if c.Dir == East {
			// Ouest vers Nord : rotation de π/2 à 0
			return -t * (math.Pi / 2)
		}
	}
	return 0
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
