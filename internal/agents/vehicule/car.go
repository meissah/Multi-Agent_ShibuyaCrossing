package vehicule

import (
	"fmt"
	"math"
	"time"
)

// Direction enum pour la voiture
type Direction int

const (
	North Direction = iota // monte (y diminue)
	South                  // descend (y augmente)
	East                   // va à droite (x augmente)
	West                   // va à gauche (x diminue)
)

// jouera sur les distances d'arrêt et la vitesse pour chaque conducteur
type DriverProfile struct {
	Aggressiveness float64 // 0.0 (prudent) à 1.0 (agressif)
}

type Car struct {
	ID        int
	X, Y      float64
	Dir       Direction
	LaneIndex int     // 0 ou 1 pour la voie utilisée (deux voies par direction)
	Speed     float64 // vitesse actuelle en px/sec
	MaxSpeed  float64 // Vitesse maximale souhaitée en px/sec
	State     string  // "moving","stopped"
	stopDist  float64 // distance pour s'arrêter avant le feu
	updateCh  chan<- CarUpdate
	lastStep  time.Time

	// Personnalité du conducteur
	Profile             DriverProfile
	DesiredStopDistance float64

	// Perceptions de l'environnement (mises à jour par Percept)
	carInFront           *Car // Voiture devant dans ma voie
	distanceToCarInFront float64

	crossingCar           *Car // Voiture sur trajectoire perpendiculaire à proximité
	distanceToCrossingCar float64

	mustYield     bool      // Dois-je céder le passage ? (priorité à droite)
	waitStartTime time.Time // Temps depuis lequel je suis à l'arrêt

	// Intention (décidé par Deliberate)
	intention string // "accelerate", "stop", "slow_down"

	// Référence à l'environnement pour les perceptions
	env CarEnvironment
}

// Update envoyé au handler pour mise à jour de l'état de la voiture
type CarUpdate struct {
	ID     int
	X, Y   float64
	Remove bool
}

// Interface environnement pour la voiture
type CarEnvironment interface {
	GetCars() []*Car
}

// NewCar constructeur
func NewCar(id int, x, y float64, dir Direction, lane int, updateCh chan<- CarUpdate) *Car {
	return &Car{
		ID:                    id,
		X:                     x,
		Y:                     y,
		Dir:                   dir,
		LaneIndex:             lane,
		Speed:                 0,
		MaxSpeed:              100.0, // pixels/sec
		State:                 "moving",
		updateCh:              updateCh,
		lastStep:              time.Now(),
		Profile:               DriverProfile{Aggressiveness: 0.5},
		DesiredStopDistance:   50.0,
		intention:             "accelerate",
		distanceToCarInFront:  math.MaxFloat64,
		distanceToCrossingCar: math.MaxFloat64,
		mustYield:             false,
	}
}

// GetID getter ID de la voiture
func (c *Car) GetID() int {
	return c.ID
}

// Start lance la boucle agent
func (c *Car) Start(env CarEnvironment, done <-chan bool) {
	c.env = env
	fmt.Printf("[CAR %d] 🚗 Démarrage - Position: (%.1f, %.1f), Direction: %s, Voie: %d\n",
		c.ID, c.X, c.Y, c.Dir, c.LaneIndex)

	ticker := time.NewTicker(40 * time.Millisecond)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				fmt.Printf("[CAR %d] ⛔ Arrêt de la goroutine\n", c.ID)
				return
			case <-ticker.C:
				c.Percept()
				c.Deliberate()
				c.Act()
				c.emit()
			}
		}
	}()
}

// Percept: perçoit l'environnement (autres voitures et feux)
// La voiture regarde autour d'elle et détecte naturellement les dangers
func (c *Car) Percept() {
	allCars := c.env.GetCars()

	// Réinitialiser les perceptions
	c.carInFront = nil
	c.distanceToCarInFront = math.MaxFloat64
	c.crossingCar = nil
	c.distanceToCrossingCar = math.MaxFloat64
	c.mustYield = false

	// Zone de perception (champ de vision de la voiture)
	const visionRange = 150.0 // La voiture "voit" à 150px autour d'elle

	// Analyser chaque voiture visible
	for _, other := range allCars {
		if other.ID == c.ID {
			continue // C'est moi-même
		}

		dist := c.distanceTo(other)

		// Ignorer les voitures trop loin (hors champ de vision)
		if dist > visionRange {
			continue
		}

		// CAS 1 : Voiture devant moi dans ma voie (même direction)
		if c.isCarDirectlyInFront(other) {
			if dist < c.distanceToCarInFront {
				c.distanceToCarInFront = dist
				c.carInFront = other
			}
		}

		// CAS 2 : Voiture sur trajectoire perpendiculaire (risque de collision)
		if c.isCarOnPerpendicularTrajectory(other, dist) {
			if dist < c.distanceToCrossingCar {
				c.distanceToCrossingCar = dist
				c.crossingCar = other

				// Vérifier la priorité à droite
				c.mustYield = c.shouldYieldToOther(other)
			}
		}
	}

	// Logs de perception (seulement si détection)
	if c.carInFront != nil {
		fmt.Printf("[CAR %d] 👀 Voiture %d DEVANT dans ma voie à %.1f px\n",
			c.ID, c.carInFront.ID, c.distanceToCarInFront)
	}

	if c.crossingCar != nil {
		if c.mustYield {
			fmt.Printf("[CAR %d] ⚠️  Voiture %d (dir: %s) à DROITE - JE CÈDE LE PASSAGE (%.1f px)\n",
				c.ID, c.crossingCar.ID, c.crossingCar.Dir, c.distanceToCrossingCar)
		} else {
			fmt.Printf("[CAR %d] ✋ Voiture %d (dir: %s) à GAUCHE - J'AI LA PRIORITÉ (%.1f px)\n",
				c.ID, c.crossingCar.ID, c.crossingCar.Dir, c.distanceToCrossingCar)
		}
	}
}

// shouldYieldToOther détermine si je dois céder le passage à l'autre voiture
// Basé sur la règle de priorité à droite
func (c *Car) shouldYieldToOther(other *Car) bool {
	// Si l'autre voiture est à l'arrêt depuis longtemps, je peux passer
	if other.Speed == 0 && !time.Time.IsZero(other.waitStartTime) {
		waitDuration := time.Since(other.waitStartTime)
		if waitDuration > 2*time.Second {
			// L'autre attend depuis plus de 2 secondes, elle me cède probablement le passage
			return false
		}
	}

	// Déterminer si l'autre voiture vient de ma droite
	return c.isComingFromRight(other)
}

// isComingFromRight détermine si une autre voiture vient de ma droite
// selon la règle de priorité à droite
func (c *Car) isComingFromRight(other *Car) bool {
	switch c.Dir {
	case North: // Je vais vers le haut
		// Ma droite est à l'EST (West pour l'autre car elle vient de la droite)
		return other.Dir == West

	case South: // Je vais vers le bas
		// Ma droite est à l'OUEST (East pour l'autre)
		return other.Dir == East

	case East: // Je vais vers la droite
		// Ma droite est au SUD (North pour l'autre)
		return other.Dir == North

	case West: // Je vais vers la gauche
		// Ma droite est au NORD (South pour l'autre)
		return other.Dir == South
	}
	return false
}

// isCarDirectlyInFront : voiture devant dans la même direction et même voie
func (c *Car) isCarDirectlyInFront(other *Car) bool {
	const laneTolerance = 35.0

	// Doit aller dans la même direction (ou être à l'arrêt)
	if other.Dir != c.Dir {
		return false
	}

	// Doit être approximativement sur la même voie
	// Et doit être devant moi dans ma direction de déplacement
	switch c.Dir {
	case North:
		return math.Abs(other.X-c.X) < laneTolerance && other.Y < c.Y
	case South:
		return math.Abs(other.X-c.X) < laneTolerance && other.Y > c.Y
	case East:
		return math.Abs(other.Y-c.Y) < laneTolerance && other.X > c.X
	case West:
		return math.Abs(other.Y-c.Y) < laneTolerance && other.X < c.X
	}
	return false
}

// isCarOnPerpendicularTrajectory détecte si une voiture représente un risque de collision
func (c *Car) isCarOnPerpendicularTrajectory(other *Car, distance float64) bool {
	const dangerZone = 120.0

	if distance > dangerZone {
		return false
	}

	if !c.isPerpendicularDirection(other.Dir) {
		return false
	}

	return c.willTrajectoryIntersect(other)
}

// isPerpendicularDirection vérifie si une direction est perpendiculaire à la mienne
func (c *Car) isPerpendicularDirection(otherDir Direction) bool {
	switch c.Dir {
	case North, South:
		return otherDir == West || otherDir == East
	case West, East:
		return otherDir == North || otherDir == South
	}
	return false
}

// willTrajectoryIntersect vérifie si nos trajectoires vont se croiser
func (c *Car) willTrajectoryIntersect(other *Car) bool {
	const trajectoryTolerance = 80.0

	switch c.Dir {
	case North, South:
		xAligned := math.Abs(other.X-c.X) < trajectoryTolerance
		yAligned := math.Abs(other.Y-c.Y) < trajectoryTolerance
		return xAligned && yAligned

	case West, East:
		xAligned := math.Abs(other.X-c.X) < trajectoryTolerance
		yAligned := math.Abs(other.Y-c.Y) < trajectoryTolerance
		return xAligned && yAligned
	}
	return false
}

// distanceTo calcule la distance euclidienne à une autre voiture
func (c *Car) distanceTo(other *Car) float64 {
	dx := other.X - c.X
	dy := other.Y - c.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Deliberate: la voiture prend des décisions basées sur ses perceptions
func (c *Car) Deliberate() {
	const safeDistance = 60.0
	const slowDownDistance = 100.0
	const crossingSafeDistance = 70.0
	const crossingSlowDistance = 110.0
	const maxWaitTime = 3 * time.Second // Temps max d'attente avant de forcer le passage

	// PRIORITÉ 1 : Voiture qui traverse ma trajectoire (risque de collision)
	if c.crossingCar != nil {
		// Si je dois céder le passage (priorité à droite)
		if c.mustYield {
			// Mais si j'attends depuis trop longtemps, je force le passage (éviter deadlock)
			if !time.Time.IsZero(c.waitStartTime) && time.Since(c.waitStartTime) > maxWaitTime {
				fmt.Printf("[CAR %d] ⏱️  TIMEOUT - Je force le passage après %v d'attente\n",
					c.ID, maxWaitTime)
				c.intention = "accelerate"
				c.waitStartTime = time.Time{} // Reset du timer
				return
			}

			if c.distanceToCrossingCar < crossingSafeDistance {
				c.intention = "stop"
				// Marquer le début de l'attente si je viens de m'arrêter
				if c.Speed > 0 {
					c.waitStartTime = time.Now()
				}
				fmt.Printf("[CAR %d] 🛑 CÈDE LE PASSAGE - Priorité à droite (voiture %d à %.1f px)\n",
					c.ID, c.crossingCar.ID, c.distanceToCrossingCar)
				return
			} else if c.distanceToCrossingCar < crossingSlowDistance {
				c.intention = "slow_down"
				fmt.Printf("[CAR %d] 🐌 RALENTISSEMENT - Approche avec priorité à droite (%.1f px)\n",
					c.ID, c.distanceToCrossingCar)
				return
			}
		} else {
			// Je n'ai pas à céder, mais je dois quand même faire attention
			if c.distanceToCrossingCar < crossingSafeDistance {
				c.intention = "slow_down"
				fmt.Printf("[CAR %d] 🐌 PRUDENCE - J'ai la priorité mais je ralentis (%.1f px)\n",
					c.ID, c.distanceToCrossingCar)
				return
			}
		}
	}

	// Si j'étais à l'arrêt et qu'il n'y a plus de danger, reset le timer
	if c.crossingCar == nil && !time.Time.IsZero(c.waitStartTime) {
		c.waitStartTime = time.Time{}
	}

	// PRIORITÉ 2 : Voiture devant dans ma voie
	if c.carInFront != nil {
		if c.distanceToCarInFront < safeDistance {
			c.intention = "stop"
			fmt.Printf("[CAR %d] 🛑 STOP - Trop proche de voiture %d (%.1f px)\n",
				c.ID, c.carInFront.ID, c.distanceToCarInFront)
			return
		} else if c.distanceToCarInFront < slowDownDistance {
			c.intention = "slow_down"
			fmt.Printf("[CAR %d] 🐌 RALENTIT - Voiture %d devant (%.1f px)\n",
				c.ID, c.carInFront.ID, c.distanceToCarInFront)
			return
		}
	}

	// PRIORITÉ 3 : Aucun obstacle, on peut accélérer
	c.intention = "accelerate"
}

// Act: agit en fonction de l'intention (ajuste la vitesse et bouge)
func (c *Car) Act() {
	dt := time.Since(c.lastStep).Seconds()
	c.lastStep = time.Now()

	oldSpeed := c.Speed

	// Ajuster la vitesse
	switch c.intention {
	case "stop":
		c.Speed = 0
		c.State = "stopped"
	case "slow_down":
		c.Speed *= 0.90
		if c.Speed < 25 {
			c.Speed = 25
		}
		c.State = "moving"
	case "accelerate":
		if c.Speed < c.MaxSpeed {
			c.Speed += 120.0 * dt
		}
		if c.Speed > c.MaxSpeed {
			c.Speed = c.MaxSpeed
		}
		c.State = "moving"
	}

	// Appliquer le mouvement
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

	// Log uniquement si changement significatif de vitesse
	if math.Abs(c.Speed-oldSpeed) > 5 {
		fmt.Printf("[CAR %d] 🚙 Position: (%.1f, %.1f), Vitesse: %.1f→%.1f px/s, État: %s\n",
			c.ID, c.X, c.Y, oldSpeed, c.Speed, c.State)
	}
}

// emit envoie la mise à jour au handler
func (c *Car) emit() {
	const screenSize = 750.0
	const margin = 50.0

	shouldRemove := false
	switch c.Dir {
	case North:
		shouldRemove = c.Y < -margin
	case South:
		shouldRemove = c.Y > screenSize+margin
	case East:
		shouldRemove = c.X > screenSize+margin
	case West:
		shouldRemove = c.X < -margin
	}

	if shouldRemove {
		fmt.Printf("[CAR %d] ✅ SORTIE DE L'ÉCRAN - Position finale: (%.1f, %.1f)\n",
			c.ID, c.X, c.Y)
		c.updateCh <- CarUpdate{
			ID:     c.ID,
			X:      c.X,
			Y:      c.Y,
			Remove: true,
		}
		return
	}

	c.updateCh <- CarUpdate{
		ID:     c.ID,
		X:      c.X,
		Y:      c.Y,
		Remove: false,
	}
}
