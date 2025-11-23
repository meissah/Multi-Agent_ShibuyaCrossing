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
	ID           AgentID
	X, Y         float64
	Dir          Direction
	TurnLeft     bool    // indique si la voiture envisage de tourner à gauche
	LaneIndex    int     // 0 ou 1 pour la voie utilisée (deux voies par direction)
	Speed        float64 // vitesse actuelle en px/sec
	MaxSpeed     float64 // Vitesse maximale souhaitée en px/sec
	DrivingState string  // "moving","stopped"
	stopDist     float64 // distance pour s'arrêter avant le feu
	//updateCh  chan<- CarUpdate
	lastStep time.Time

	// Personnalité du conducteur
	Profile             DriverProfile
	DesiredStopDistance float64

	// Perceptions de l'environnement (mises à jour par Percept)
	carInFront           *Car // Voiture devant dans ma voie
	distanceToCarInFront float64

	crossingCar           *Car // Voiture sur trajectoire perpendiculaire à proximité
	distanceToCrossingCar float64

	mustYield     bool      // Savoir si ceder le passage ou pas (ici priorité à droite implementée)
	waitStartTime time.Time // Temps depuis lequel je suis à l'arrêt

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

// Interface environnement pour la voiture ( on mettra plus tard les feux )
type CarEnvironment interface {
	GetCars() []*Car // Retourne toutes les voitures dans l'environnement
}

// NewCar constructeur
func NewCar(id AgentID, x, y float64, dir Direction, lane int, turnleft bool, syncChan chan int) *Car {
	return &Car{
		ID:           id,
		X:            x,
		Y:            y,
		Dir:          dir,
		LaneIndex:    lane,
		TurnLeft:     turnleft,
		Speed:        0,
		MaxSpeed:     100.0,    // pixels/sec j'ai mis au hasard a ajuster selon les tests
		DrivingState: "moving", // par défaut en mouvement mais a voir parce que si trop de voiture tecnhiquemment elle devrait s'arreter
		AmIAlive:     true,
		//updateCh:              updateCh,
		lastStep:              time.Now(),
		Profile:               DriverProfile{Aggressiveness: 0.5}, // profil normale par défaut
		DesiredStopDistance:   50.0,
		intention:             "accelerate",    // pareil que pour le State moving à voir plus tard
		distanceToCarInFront:  math.MaxFloat64, // initialisé à l'infini
		distanceToCrossingCar: math.MaxFloat64,
		mustYield:             false, // par défaut pas besoin de céder le passage
		syncChan:              syncChan,
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

// Deliberate: la voiture prend des décisions basées sur ses perceptions
func (c *Car) Deliberate() {

	shouldRemove := false // indique si la voiture doit être supprimée
	if shouldRemove {
		c.AmIAlive = false
		//fmt.Println("AM DEAD!")
		return
	}

	c.intention = "accelerate"
}

// Act: agit en fonction de l'intention (ajuste la vitesse et bouge)
func (c *Car) Act(ActChn chan<- ObjectAddRequest) {
	dt := time.Since(c.lastStep).Seconds() // temps écoulé depuis le dernier step en secondes ( chaque step fait toutes les 40ms defini dans Start par le ticker)
	c.lastStep = time.Now()                // met à jour le temps du dernier step

	// Ajuster la vitesse
	switch c.intention {
	case "stop":
		c.Speed = 0
		c.DrivingState = "stopped"
	case "slow_down":
		c.Speed *= 0.90
		if c.Speed < 25 {
			c.Speed = 25
		}
		c.DrivingState = "moving"
	case "accelerate":
		if c.Speed < c.MaxSpeed {
			c.Speed += 120.0 * dt
		}
		if c.Speed > c.MaxSpeed {
			c.Speed = c.MaxSpeed
		}
		c.DrivingState = "moving"
	}

	// Mettre à jour la position suite au changement de vitesse
	if c.Speed > 0 {
		switch c.Dir {
		case North:
			c.Y -= c.Speed * dt
		case South:
			c.Y += c.Speed * dt
		case East:
			c.X += c.Speed * dt
		case West:
			c.X -= c.Speed * dt
		}
	}

	c.emit(ActChn)
}

// emit envoie la mise à jour au handler
func (c *Car) emit(ActChn chan<- ObjectAddRequest) {
	// const screenSize = 750.0
	// const margin = 50.0

	// shouldRemove := false // indique si la voiture doit être supprimée

	// // Vérifie si la voiture est hors écran avec une marge
	// switch c.Dir {
	// case North:
	// 	shouldRemove = c.Y < -margin
	// case South:
	// 	shouldRemove = c.Y > screenSize+margin
	// case East:
	// 	shouldRemove = c.X > screenSize+margin
	// case West:
	// 	shouldRemove = c.X < -margin
	// }

	// if shouldRemove {
	// 	// fmt.Printf("[CAR %d] SORTIE DE L'ÉCRAN - Position finale: (%.1f, %.1f)\n",
	// 	//c.ID, c.X, c.Y)
	// 	// c.updateCh <- CarUpdate{
	// 	// 	ID:     c.ID,
	// 	// 	X:      c.X,
	// 	// 	Y:      c.Y,
	// 	// 	Remove: true, // indique au handler de supprimer la voiture
	// 	// }
	// 	return
	// }
	RespChan := make(chan int)
	ActChn <- ObjectAddRequest{ // ajouter next step d'un agent donné
		Object:   c,
		RespChan: RespChan,
	}
	<-RespChan // voir ligne l19 (Start) dans carspawn pour explication ( attend le append dans l'object slice handler )
}

func (c *Car) Position() (float64, float64) {
	return c.X, c.Y
}
func (c *Car) TurningLeft() bool {
	return c.TurnLeft
}

func (c *Car) Angle() float64 {
	switch c.Dir {
	case East:
		return 0.0
	case South:
		return math.Pi / 2
	case West:
		return math.Pi
	case North:
		return 3 * math.Pi / 2
	default:
		return 0.0
	}
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
