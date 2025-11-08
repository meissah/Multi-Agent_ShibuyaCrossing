package vehicule

import (
	"math"
	"time"

	cars_lights "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/vehicule_lights"
)

// Direction définit les directions possibles pour une voiture
type Direction int

const (
	North Direction = iota // Vers le haut de l'écran
	South                  // Vers le bas de l'écran
	East                   // Vers la droite de l'écran
	West                   // Vers la gauche de l'écran
)

// TurnDecision définit la décision de virage à l'intersection
type TurnDecision int

const (
	Straight TurnDecision = iota
	TurnLeft
	TurnRight
)

// CarEnvironment définit l'interface que le jeu doit implémenter
// pour fournir les informations nécessaires aux voitures
type CarEnvironment interface {
	// GetLights retourne les feux proches de la position donnée
	GetLights(x, y float64) []cars_lights.CarsLight
	// IsIntersectionOccupied vérifie si l'intersection est occupée
	IsIntersectionOccupied(x, y float64) bool
}

// Car représente une voiture dans la simulation
type Car struct {
	ID        int              // Identifiant unique de la voiture
	X, Y      float64          // Position actuelle
	Speed     float64          // Vitesse en pixels par frame
	Direction Direction        // Direction actuelle
	State     string           // État : "driving", "waiting", "turning"
	Lane      int              // Voie : 0 = voie de gauche, 1 = voie de droite
	updates   chan<- CarUpdate // Canal pour envoyer les mises à jour de position
}

// CarUpdate structure pour les mises à jour de position
type CarUpdate struct {
	ID    int
	X, Y  float64
	State string
}

// NewCar crée une nouvelle voiture
func NewCar(id int, x, y float64, dir Direction, lane int, updates chan<- CarUpdate) *Car {
	return &Car{
		ID:           id,
		X:            x,
		Y:            y,
		Speed:        2.0,
		Direction:    dir,
		State:        "driving",
		Lane:         lane,
		TurnDecision: Straight, // Par défaut, tout droit
		updates:      updates,
		isTurning:    false,
	}
}

// Start lance la goroutine qui gère le comportement de la voiture
func (c *Car) Start(env CarEnvironment, done chan bool) {
	go func() {
		// Ticker pour mettre à jour la position 60 fois par seconde
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()

		for {
			select {
			case <-done: // Signal d'arrêt
				return
			case <-ticker.C: // Mise à jour régulière
				c.Act(env)
				// Envoie la mise à jour de position
				c.updates <- CarUpdate{
					ID:    c.ID,
					X:     c.X,
					Y:     c.Y,
					State: c.State,
				}
			}
		}
	}()
}

// Act implémente le comportement de la voiture
func (c *Car) Act(env CarEnvironment) {
	// Vérification des feux
	lights := env.GetLights(c.X, c.Y)
	for _, light := range lights {
		if light.IsRed() {
			c.State = "waiting"
			return
		}
	}

	// Vérification de l'intersection
	if env.IsIntersectionOccupied(c.X, c.Y) {
		c.State = "waiting"
		return
	}

	// Si tout est clair, on avance
	c.State = "driving"
	c.move()
}

// isInIntersection vérifie si la voiture est dans l'intersection
func (c *Car) isInIntersection() bool {
	const (
		intersectionMinX = 245.0
		intersectionMaxX = 495.0
		intersectionMinY = 300.0
		intersectionMaxY = 450.0
	)

	return c.X >= intersectionMinX && c.X <= intersectionMaxX &&
		c.Y >= intersectionMinY && c.Y <= intersectionMaxY
}

// move gère le déplacement de la voiture selon sa direction
// move gère le déplacement de la voiture
func (c *Car) move() {
	const laneOffset = 27.5 // Demi-largeur de voie

	// Si on est en train de tourner, gérer le virage
	if c.isTurning {
		c.executeTurn()
		return
	}

	// Vérifier si on doit commencer un virage
	if c.shouldStartTurn() {
		c.startTurn()
		return
	}

	// Déplacement normal selon la direction
	switch c.Direction {
	case North:
		c.Y -= c.Speed
		c.X = 272.5 + float64(c.Lane)*laneOffset
	case South:
		c.Y += c.Speed
		c.X = 477.5 - float64(c.Lane)*laneOffset
	case East:
		c.X += c.Speed
		c.Y = 327.5 + float64(c.Lane)*laneOffset
	case West:
		c.X -= c.Speed
		c.Y = 422.5 - float64(c.Lane)*laneOffset
	}

	// Gestion du bouclage de l'écran
	c.wrapAroundScreen()
}

// shouldStartTurn détermine si la voiture doit commencer à tourner
func (c *Car) shouldStartTurn() bool {
	if c.TurnDecision == Straight {
		return false
	}

	const intersectionCenter = 370.0
	const turnThreshold = 20.0

	switch c.Direction {
	case North:
		return c.Y <= intersectionCenter+turnThreshold && c.Y >= intersectionCenter-turnThreshold
	case South:
		return c.Y >= intersectionCenter-turnThreshold && c.Y <= intersectionCenter+turnThreshold
	case East:
		return c.X >= intersectionCenter-turnThreshold && c.X <= intersectionCenter+turnThreshold
	case West:
		return c.X <= intersectionCenter+turnThreshold && c.X >= intersectionCenter-turnThreshold
	}

	return false
}

// startTurn initialise le virage
func (c *Car) startTurn() {
	c.isTurning = true
	c.turnStartX = c.X
	c.turnStartY = c.Y
	c.turnProgress = 0.0
	c.State = "turning"
}

// executeTurn exécute le virage progressivement
func (c *Car) executeTurn() {
	const turnSpeed = 0.05 // Vitesse de progression du virage
	c.turnProgress += turnSpeed

	if c.turnProgress >= 1.0 {
		// Virage terminé
		c.finishTurn()
		return
	}

	// Calculer la nouvelle position sur l'arc de cercle
	const turnRadius = 50.0
	const intersectionCenter = 370.0

	angle := c.turnProgress * 1.5708 // π/2 radians (90 degrés)

	// Ajuster l'angle selon la direction et le type de virage
	if c.TurnDecision == TurnLeft {
		angle = -angle
	}

	// Calculer la position sur l'arc
	switch c.Direction {
	case North:
		if c.TurnDecision == TurnLeft {
			c.X = intersectionCenter - turnRadius + turnRadius*math.Cos(angle)
			c.Y = intersectionCenter - turnRadius*math.Sin(angle)
		} else {
			c.X = intersectionCenter + turnRadius - turnRadius*math.Cos(angle)
			c.Y = intersectionCenter - turnRadius*math.Sin(angle)
		}
	case South:
		if c.TurnDecision == TurnLeft {
			c.X = intersectionCenter + turnRadius - turnRadius*math.Cos(angle)
			c.Y = intersectionCenter + turnRadius*math.Sin(angle)
		} else {
			c.X = intersectionCenter - turnRadius + turnRadius*math.Cos(angle)
			c.Y = intersectionCenter + turnRadius*math.Sin(angle)
		}
	case East:
		if c.TurnDecision == TurnLeft {
			c.X = intersectionCenter + turnRadius*math.Sin(angle)
			c.Y = intersectionCenter - turnRadius + turnRadius*math.Cos(angle)
		} else {
			c.X = intersectionCenter + turnRadius*math.Sin(angle)
			c.Y = intersectionCenter + turnRadius - turnRadius*math.Cos(angle)
		}
	case West:
		if c.TurnDecision == TurnLeft {
			c.X = intersectionCenter - turnRadius*math.Sin(angle)
			c.Y = intersectionCenter + turnRadius - turnRadius*math.Cos(angle)
		} else {
			c.X = intersectionCenter - turnRadius*math.Sin(angle)
			c.Y = intersectionCenter - turnRadius + turnRadius*math.Cos(angle)
		}
	}
}

// finishTurn termine le virage et met à jour la direction
func (c *Car) finishTurn() {
	c.isTurning = false
	c.State = "driving"

	if c.TurnDecision == TurnLeft {
		switch c.Direction {
		case North:
			c.Direction = West
		case South:
			c.Direction = East
		case East:
			c.Direction = North
		case West:
			c.Direction = South
		}
	} else if c.TurnDecision == TurnRight {
		switch c.Direction {
		case North:
			c.Direction = East
		case South:
			c.Direction = West
		case East:
			c.Direction = South
		case West:
			c.Direction = North
		}
	}

	// Réinitialiser la décision de virage pour le prochain carrefour
	c.TurnDecision = Straight
}

// wrapAroundScreen gère le bouclage de l'écran
func (c *Car) wrapAroundScreen() {
	if c.Y < -50 {
		c.Y = 800
	}
	if c.Y > 800 {
		c.Y = -50
	}
	if c.X < -50 {
		c.X = 800
	}
	if c.X > 800 {
		c.X = -50
	}
}
