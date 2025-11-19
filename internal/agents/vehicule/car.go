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
			step = <-ag.syncChan

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

const (
	// Distances de sécurité et de détection
	safeDistance      = 50.0  // Distance minimale de sécurité avant collision
	detectionDistance = 150.0 // Distance à laquelle on commence à ralentir (augmenté de 100 à 150)
	stopDistance      = 40.0  // Distance à laquelle on s'arrête complètement

	// Distances pour la priorité à droite
	intersectionRadius = 120.0 // Rayon de détection d'intersection (augmenté de 100 à 120)
	priorityDistance   = 100.0 // Distance pour détecter priorité à droite (augmenté de 80 à 100)

	// Vitesses
	decelRate        = 80.0 // Accélération de décélération (augmenté de 60 à 80)
	minSlowSpeed     = 15.0 // Vitesse minimale quand on ralentit (augmenté de 10 à 15)
	decelTargetDelta = 50.0 // Réduction de vitesse cible (augmenté de 40 à 50)
)

// NewCar constructeur
func NewCar(id AgentID, x, y float64, dir Direction, lane int, syncChan chan int) *Car {
	return &Car{
		ID:           id,
		X:            x,
		Y:            y,
		Dir:          dir,
		LaneIndex:    lane,
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

// Percept: perçoit l'environnement (autres voitures et feux ( pas encore fait ))
// La voiture regarde autour d'elle et détecte naturellement les dangers
func (c *Car) Percept(env *[]Object) {
	//allCars := c.env.GetCars() // Récupère toutes les voitures de l'environnement ( pour rappel c'est le handler qui gère toutes les voitures et getCars retourne une slice de toutes les voitures )

	// Réinitialiser les perceptions à chaque cycle ( de 40ms )
	c.carInFront = nil
	c.distanceToCarInFront = math.MaxFloat64
	c.crossingCar = nil
	c.distanceToCrossingCar = math.MaxFloat64
	c.mustYield = false

	// Zone de perception (champ de vision de la voiture)
	const visionRange = 150.0 // La voiture "voit" à 150px autour d'elle

	// Analyser chaque voiture visible
	// for _, other := range allCars {
	// 	if other.ID == c.ID {
	// 		continue // C'est moi-même
	// 	}

	// 	dist := c.distanceTo(other) // Calcule la distance à l'autre voiture

	// 	// Ignorer les voitures trop loin (hors champ de vision)
	// 	if dist > visionRange {
	// 		continue
	// 	}

	// 	// CAS 1 : Voiture devant moi dans ma voie (même direction)
	// 	if c.isCarDirectlyInFront(other) {
	// 		if dist < c.distanceToCarInFront {
	// 			c.distanceToCarInFront = dist
	// 			c.carInFront = other
	// 		}
	// 	}

	// 	// CAS 2 : Voiture sur trajectoire perpendiculaire (risque de collision)
	// 	if c.isCarOnPerpendicularTrajectory(other, dist) {
	// 		if dist < c.distanceToCrossingCar {
	// 			c.distanceToCrossingCar = dist
	// 			c.crossingCar = other

	// 			// Vérifier la priorité à droite
	// 			c.mustYield = c.shouldYieldToOther(other)
	// 		}
	// 	}
	// }

	// Logs de perception (seulement si détection)
	if c.carInFront != nil {
		// fmt.Printf("[CAR %d] OtherCar %d devant moi à %.1f px\n",
		//c.ID, c.carInFront.ID, c.distanceToCarInFront)
	}

	if c.crossingCar != nil { // voiture sur trajectoire perpendiculaire
		if c.mustYield { // si je dois céder le passage
			// fmt.Printf("[CAR %d] OtherCar %d (dir: %s) à droite, je cède le passage (%.1f px)\n",
			//c.ID, c.crossingCar.ID, c.crossingCar.Dir, c.distanceToCrossingCar)
		} else {
			// fmt.Printf("[CAR %d] OtherCar %d (dir: %s) à gauche je suis prio (%.1f px)\n",
			//c.ID, c.crossingCar.ID, c.crossingCar.Dir, c.distanceToCrossingCar)
		}
	}
}

// shouldYieldToOther détermine si je dois céder le passage à l'autre voiture
// Basé sur la règle de priorité à droite
func (c *Car) shouldYieldToOther(other *Car) bool {
	// Si l'autre voiture est à l'arrêt depuis longtemps, je peux passer
	// en gros si other  est à l'arrêt depuis plus de 2 secondes, c continue de rouler ( donc other lui cede le passage)
	if other.Speed == 0 && !time.Time.IsZero(other.waitStartTime) { // rappel : waitStartTime c'est le tps depuis laquel je suis a l'arret et  time.Time.IsZero vérifie si le temps est non initialisé ( c'est a dire que la voiture n'est pas à l'arrêt (on le contraire d'ou le ! )
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
		return other.Dir == West // donc je dois céder le passage si l'autre vient de l'ouest

	case South: // Je vais vers le bas
		// Ma droite est à l'OUEST (East pour l'autre)
		return other.Dir == East // donc je dois céder le passage si l'autre vient de l'est

	case East: // Je vais vers la droite
		// Ma droite est au SUD (North pour l'autre)
		return other.Dir == North // donc je dois céder le passage si l'autre vient du nord

	case West: // Je vais vers la gauche
		// Ma droite est au NORD (South pour l'autre)
		return other.Dir == South // donc je dois céder le passage si l'autre vient du sud
	}
	return false // Par défaut, ne pas céder le passage
}

// isCarDirectlyInFront : voiture devant dans la même direction et même voie
func (c *Car) isCarDirectlyInFront(other *Car) bool {
	const laneTolerance = 35.0 // distance max pour être considéré dans la même voie

	// Doit aller dans la même direction (ou être à l'arrêt)
	if other.Dir != c.Dir {
		return false
	}

	// Doit être approximativement sur la même voie
	// Et doit être devant moi dans ma direction de déplacement
	// sachant que c'est pas possible d'avoir une voiture dans la même voie qui va dans le sens inverse ( on est a sens unique )
	switch c.Dir {
	case North:
		return math.Abs(other.X-c.X) < laneTolerance && other.Y < c.Y // si la voiture est devant moi ( y diminue quand on va vers le nord )
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
		return false // Trop loin pour être un danger
	}

	if !c.isPerpendicularDirection(other.Dir) {
		return false // Directions non perpendiculaires donc pas de risque de collision
	}

	return c.willTrajectoryIntersect(other) // Vérifie si les trajectoires vont se croiser
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
		xAligned := math.Abs(other.X-c.X) < trajectoryTolerance // si les x sont alignés ( on est sur la même colonne ) et donc trajectoire qui se croise
		yAligned := math.Abs(other.Y-c.Y) < trajectoryTolerance
		return xAligned && yAligned // si les deux sont alignés alors trajectoire qui se croise

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
	return math.Sqrt(dx*dx + dy*dy) // calcule basico qu'on connait
}

// Deliberate: la voiture prend des décisions basées sur ses perceptions
func (c *Car) Deliberate() {
	const safeDistance = 60.0
	const slowDownDistance = 90.0
	const crossingSafeDistance = 70.0
	const crossingSlowDistance = 110.0
	const maxWaitTime = 3 * time.Second // Temps max d'attente avant de forcer le passage

	const screenSize = 750.0
	const margin = 50.0

	shouldRemove := false // indique si la voiture doit être supprimée

	// Vérifie si la voiture est hors écran avec une marge
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
		c.AmIAlive = false
		//fmt.Println("AM DEAD!")
		return
	}

	// PRIORITÉ 1 : Voiture qui traverse ma trajectoire (risque de collision)
	if c.crossingCar != nil {
		// Si je dois céder le passage (priorité à droite)
		if c.mustYield {
			// Mais si j'attends depuis trop longtemps, je force le passage (éviter deadlock)
			if !time.Time.IsZero(c.waitStartTime) && time.Since(c.waitStartTime) > maxWaitTime {
				// fmt.Printf("[CAR %d] TIMEOUT, force passage après %v d'attente\n",
				//c.ID, maxWaitTime)
				c.intention = "accelerate"
				c.waitStartTime = time.Time{} // Reset du timer
				return
			}

			if c.distanceToCrossingCar < crossingSafeDistance { // si trop proche de la voiture perpendiculaire
				c.intention = "stop"
				// Marquer le début de l'attente si je viens de m'arrêter
				if c.Speed > 0 {
					c.waitStartTime = time.Now() // commence le timer d'attente
				}
				// fmt.Printf("[CAR %d] cede passage, prio droite (voiture %d à %.1f px)\n",
				//c.ID, c.crossingCar.ID, c.distanceToCrossingCar)
				return
			} else if c.distanceToCrossingCar < crossingSlowDistance {
				c.intention = "slow_down"
				// fmt.Printf("[CAR %d] RALENTISSEMENT - Approche avec priorité à droite (%.1f px)\n",
				//c.ID, c.distanceToCrossingCar)
				return
			}
		} else {
			// Je n'ai pas à céder, mais je dois quand même faire attention
			if c.distanceToCrossingCar < crossingSafeDistance {
				c.intention = "slow_down"
				// fmt.Printf("[CAR %d] PRUDENCE - J'ai la priorité mais je ralentis (%.1f px)\n",
				//c.ID, c.distanceToCrossingCar)
				return
			}
		}
	}

	// Si j'étais à l'arrêt et qu'il n'y a plus de danger, reset le timer
	if c.crossingCar == nil && !time.Time.IsZero(c.waitStartTime) {
		c.waitStartTime = time.Time{} // resert timer car je suis plus à l'arrêt
	}

	// PRIORITÉ 2 : Voiture devant dans ma voie
	if c.carInFront != nil {
		if c.distanceToCarInFront < safeDistance {
			c.intention = "stop"
			// fmt.Printf("[CAR %d] STOP - Trop proche de voiture %d (%.1f px)\n",
			//c.ID, c.carInFront.ID, c.distanceToCarInFront)
			return
		} else if c.distanceToCarInFront < slowDownDistance {
			c.intention = "slow_down"
			// fmt.Printf("[CAR %d] RALENTIT - Voiture %d devant (%.1f px)\n",
			//c.ID, c.carInFront.ID, c.distanceToCarInFront)
			return
		}
	}

	// PRIORITÉ 3 : Aucun obstacle, on peut accélérer
	c.intention = "accelerate"
}

// Act: agit en fonction de l'intention (ajuste la vitesse et bouge)
func (c *Car) Act(ActChn chan<- ObjectAddRequest) {
	dt := time.Since(c.lastStep).Seconds() // temps écoulé depuis le dernier step en secondes ( chaque step fait toutes les 40ms defini dans Start par le ticker)
	c.lastStep = time.Now()                // met à jour le temps du dernier step

	oldSpeed := c.Speed

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

	// Log uniquement si changement significatif de vitesse
	if math.Abs(c.Speed-oldSpeed) > 5 {
		// fmt.Printf("[CAR %d] 🚙 Position: (%.1f, %.1f), Vitesse: %.1f→%.1f px/s, État: %s\n",
		//c.ID, c.X, c.Y, oldSpeed, c.Speed, c.State)
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
	ActChn <- ObjectAddRequest{
		Object:   c,
		RespChan: RespChan,
	}
	<-RespChan
}

func (c *Car) Position() (float64, float64) {
	return c.X, c.Y
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
