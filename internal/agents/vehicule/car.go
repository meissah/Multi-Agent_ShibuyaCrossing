package vehicule

import (
	"math"
	"time"

	cars_lights "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/vehicule_lights"
)

// Direction enum existante supposée
type Direction int

const (
	North Direction = iota // monte (y diminue)
	South                  // descend (y augmente)
	East                   // va à droite (x augmente)
	West                   // va à gauche (x diminue)
)

// Décision placeholder (non utilisée pour l’instant)
type TurnDecision int

const (
	Straight TurnDecision = iota
	TurnLeft
	TurnRight
)

// Interface environnement pour la voiture (inspiré simu3)
type CarEnvironment interface {
	GetLights(x, y float64) []cars_lights.VehiculeLight
	IsIntersectionOccupied(x, y float64) bool
}

// Update envoyé au handler
type CarUpdate struct {
	ID    int
	X, Y  float64
	State string // "moving","stopped","done"
	Dir   Direction
}

type Car struct {
	ID           int
	X, Y         float64
	Dir          Direction
	LaneIndex    int // 0 ou 1
	Speed        float64
	State        string
	TurnDecision TurnDecision
	stopDist     float64
	updateCh     chan<- CarUpdate

	// interne
	lastStep time.Time
	env      CarEnvironment
	done     <-chan bool
}

// NewCar constructeur
func NewCar(id int, x, y float64, dir Direction, lane int, ch chan<- CarUpdate) *Car {
	return &Car{
		ID:        id,
		X:         x,
		Y:         y,
		Dir:       dir,
		LaneIndex: lane,
		Speed:     90, // px/sec
		State:     "moving",
		stopDist:  90,
		updateCh:  ch,
		lastStep:  time.Now(),
	}
}

func (c *Car) IDCar() int { return c.ID }

// Start lance la boucle agent
func (c *Car) Start(env CarEnvironment, done <-chan bool) {
	c.env = env
	c.done = done
	go c.loop()
}

func (c *Car) loop() {
	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.Percept()
			c.Deliberate()
			c.Act()
			c.emit()
			c.cleanupIfOut()
		}
	}
}

// Percept: récupère les feux proches
func (c *Car) Percept() {
	// rien stocké si pas besoin, on traite directement dans Deliberate
}

// Deliberate: décider arrêt ou mouvement
func (c *Car) Deliberate() {
	if c.env == nil {
		return
	}
	lights := c.env.GetLights(c.X, c.Y)
	shouldStop := false
	for _, l := range lights {
		if l.State == "red" {
			// distance selon direction (projection simple)
			d := math.Hypot(l.X-c.X, l.Y-c.Y)
			if d < c.stopDist && c.approachingLight(l) {
				shouldStop = true
				break
			}
		}
	}
	if shouldStop {
		c.State = "stopped"
	} else {
		c.State = "moving"
	}
}

// Act: avancer si moving
func (c *Car) Act() {
	if c.State != "moving" {
		return
	}
	now := time.Now()
	dt := now.Sub(c.lastStep).Seconds()
	c.lastStep = now
	dist := c.Speed * dt

	switch c.Dir {
	case North:
		c.Y -= dist
	case South:
		c.Y += dist
	case East:
		c.X += dist
	case West:
		c.X -= dist
	}
}

// Approche du feu (projection directionnelle simplifiée)
func (c *Car) approachingLight(l cars_lights.VehiculeLight) bool {
	switch c.Dir {
	case North:
		return l.Y < c.Y
	case South:
		return l.Y > c.Y
	case East:
		return l.X > c.X
	case West:
		return l.X < c.X
	}
	return false
}

// Envoi mise à jour
func (c *Car) emit() {
	if c.updateCh != nil {
		c.updateCh <- CarUpdate{
			ID:    c.ID,
			X:     c.X,
			Y:     c.Y,
			State: c.State,
			Dir:   c.Dir,
		}
	}
}

// Retirer si hors écran
func (c *Car) cleanupIfOut() {
	if c.X < -50 || c.X > 800 || c.Y < -50 || c.Y > 800 {
		c.State = "done"
		c.emit()
	}
}
