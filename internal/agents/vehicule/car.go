package vehicule

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
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
func (ag *Car) Start(env *[]Object) {

	go func() {
		var step int
		for {
			select {
			case step = <-ag.syncChan:

				ag.Percept(*env)

				ag.Deliberate()

				ag.syncChan <- step
				<-ag.syncChan

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
	ID            AgentID
	X, Y          float64
	Dir           Direction
	Angle         float64
	LaneIndex     int // 0 ou 1 pour la voie utilisée (deux voies par direction)
	State         string
	IsPlaceholder bool

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

	// Perception : voiture détectée devant
	carAhead           *Car    // Voiture devant (nil si aucune)
	distanceToCarAhead float64 // Distance à la voiture devant

	// Système de forces vectorielles
	forces utils.Vector // Forces accumulées appliquées à la voiture

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

	TURN_DURATION = 60.0 // nombre de steps pour faire le virage

	// Constantes pour le ralentissement progressif
	DETECTION_DISTANCE = 80.0 // Distance de détection en pixels (commence à ralentir)
	SAFE_DISTANCE      = 31.0 // Distance de sécurité minimale (7 mètres)
	MIN_SPEED          = 0    // Vitesse minimale (pour éviter l'arrêt complet)
	FORCE_SPEED_FACTOR = 3.0  // amplificateur de freinage
)

// NewCar constructeur
func NewCar(id ObjectID, x, y float64, dir Direction, lane int, turnleft bool, syncChan chan int) *Car {
	car := &Car{
		ID:            AgentID(id),
		X:             x,
		Y:             y,
		Dir:           dir,
		LaneIndex:     lane,
		State:         "active",
		IsPlaceholder: false,

		// virage
		TurnLeft:   turnleft,
		IsTurning:  false,
		TurnStep:   0.0, // compteur de 0 à TURN_DURATION
		TurnStartX: 0.0, // position de départ du virage
		TurnStartY: 0.0,

		MaxSpeed:           RandomFloatInRange(50, 80), // pixels/sec j'ai mis au hasard a ajuster selon les tests
		Speed:              2.0,                        // Vitesse initiale
		AmIAlive:           true,
		syncChan:           syncChan,
		endChan:            make(chan int),
		carColorIndex:      rand.Intn(5),
		carAhead:           nil,
		distanceToCarAhead: 0,
		forces:             utils.VectorZero,
	}
	// Enregistrer le spawn dans les métriques ( pour les statistiques )
	dirStr := directionToString(dir)
	stats.RecordCarSpawn(uint(id), dirStr)

	return car
}

// Convertit la direction ( qui est en iota ) en chaîne de caractères pour les métriques
func directionToString(dir Direction) string {
	switch dir {
	case North:
		return "North"
	case South:
		return "South"
	case East:
		return "East"
	case West:
		return "West"
	default:
		return "Unknown"
	}
}

func (c *Car) Percept(env []Object) {

	// Réinitialiser la perception ET les forces (qu'on reutilise aussi pour les voitures ) avant de prendre en compte l'environnement
	c.carAhead = nil
	c.distanceToCarAhead = 0
	c.forces = utils.VectorZero // Remettre à zéro à chaque perception

	var closestCar *Car
	minDistance := DETECTION_DISTANCE

	// Position actuelle comme Point (pour calculs vectoriels)
	myPos := utils.NewPoint(c.X, c.Y)

	// Parcourir tous les objets pour trouver les voitures
	for _, obj := range env {
		// Ignorer la voiture elle-meme (soi-même)
		if obj == nil || obj.ObjectID() == c.ObjectID() {
			continue
		}

		// Ne regarder que les voitures ( POUR L'INSTANT A VOIR APRES POUR LES PIETONS si voiture tourne a gauche et qu'il y a encore un pieton qui traverse) !!!!!!!!!!!!!!!!!!!!!!
		if obj.Type() != "car" {
			continue
		}

		// Convertir en Car pour accéder aux propriétés
		otherCar, ok := obj.(*Car)
		if !ok || otherCar == nil { //Mini correction (pierrick) car découverte d'un bug rare si la voiture tente de percevoir une autre voiture qui n'existe plus (pointeur nil)
			continue
		}

		if otherCar.State != "active" { // Ignorer les voitures inactives
			//fmt.Println("Ignored inactive car")
			continue
		}

		// Vérifier si la voiture est devant nous (même direction et même lane)
		if !c.isCarAhead(otherCar) {
			continue
		}

		// Calculer la distance
		dist := c.distanceTo(otherCar)

		// Garder la voiture la plus proche dans la zone de détection
		if dist < minDistance {
			minDistance = dist
			closestCar = otherCar
		}
	}

	// Stocker la voiture détectée ET calculer la force de répulsion vectorielle ( pour ralentir de maniere fluide/creshendo)
	if closestCar != nil {
		c.carAhead = closestCar
		c.distanceToCarAhead = minDistance

		// calcul de force vectorielle (logique identique aux piétons) -> repuslsion de la voiture devant
		otherPos := utils.NewPoint(closestCar.X, closestCar.Y)
		repulsionForce := c.calculateRepulsionForce(*myPos, *otherPos, minDistance)
		c.forces = c.forces.Add(repulsionForce)

	}
}

func (c *Car) Deliberate() {
	//A ENLEVER QUAND YAURA LES FEUX
	//At random stopping (to simulate stopping at red lights behavior or for pedestrians in french mode when turning left)
	// if !c.IsTurning && rand.Float64() < 0.2 && c.AgentID()%2 == 0 { //80% chance of stopping randomly si on a un agent id pair
	// 	return
	// }

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
		c.adjustSpeed()
		if c.Speed != 0 {
			c.TurnStep++
		} else {
			return
		}

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
				//fmt.Println("Lane gauche pour west is 0")
			} else {
				// Lane de droite
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX-5, c.TurnStartY
				p2x, p2y = c.TurnStartX-12, c.TurnStartY+15
				p3x, p3y = c.TurnStartX-22, c.TurnStartY+25
				//fmt.Println("Lane gauche pour west is 1")
			}
		} else if c.Dir == East { // spawn depuis l'ouest et tourne a gauche vers le nord
			if c.LaneIndex == 0 {
				// Lane de gauche
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX+5, c.TurnStartY
				p2x, p2y = c.TurnStartX+12, c.TurnStartY-15
				p3x, p3y = c.TurnStartX+22, c.TurnStartY-25
				//fmt.Println("Lane gauche pour East is 0")
			} else {
				// Lane de droite
				p0x, p0y = c.TurnStartX, c.TurnStartY
				p1x, p1y = c.TurnStartX+30, c.TurnStartY
				p2x, p2y = c.TurnStartX+40, c.TurnStartY-25
				p3x, p3y = c.TurnStartX+53, c.TurnStartY-45
				//fmt.Println("Lane gauche pour East is 1")
			}
		}

		// Formule de Bézier cubique
		u := 1 - t
		c.X = u*u*u*p0x + 3*u*u*t*p1x + 3*u*t*t*p2x + t*t*t*p3x
		c.Y = u*u*u*p0y + 3*u*u*t*p1y + 3*u*t*t*p2y + t*t*t*p3y

		/* Comprendre la formule Bézier cubique :
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
				c.LaneIndex = 1
			} else if c.Dir == South {
				c.Dir = East
				c.LaneIndex = 0
			} else if c.Dir == West {
				c.Dir = South
				c.LaneIndex = 1
			} else if c.Dir == East {
				c.Dir = North
				c.LaneIndex = 0
			}
		}

	} else {
		// Ajuster la vitesse en fonction de la voiture devant
		c.adjustSpeed()

		// Mouvement normal en ligne droite avec la vitesse ajustée
		switch c.Dir {
		case North:
			c.Y -= c.Speed
		case South:
			c.Y += c.Speed
		case East:
			c.X += c.Speed
		case West:
			c.X -= c.Speed
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
}

func (c *Car) Act() {
	// Enregistrer la vitesse pour les statistiques
	if c.Speed > 0 {
		stats.RecordCarSpeed(uint(c.ID), c.Speed)
	}
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

	// Récupérer l'image de la voiture selon sa couleur ( grc a son index)
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
	if c.IsPlaceholder == false {
		return Asset(c.carColorIndex + 1), op // Asset 1 à 5 pour les voitures colorées pour debug
	} else {
		return Asset(0), op // Asset 6 pour la voiture placeholder grise
	}
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
	// Vérifier d'abord si la voiture est hors limites
	if c.X < -50 || c.X > 800 || c.Y < -50 || c.Y > 800 {
		// Enregistrer la destruction si pas encore fait
		if c.AmIAlive {
			c.AmIAlive = false
			stats.RecordCarDestroy(uint(c.ID))
		}
		return false
	}
	return c.AmIAlive
}

// GetSpeed retourne la vitesse actuelle (pour debug/visualisation)
func (c *Car) GetSpeed() float64 {
	return c.Speed
}

// GetCarAhead retourne la voiture détectée devant (pour debug/visualisation)
func (c *Car) GetCarAhead() (*Car, float64) {
	return c.carAhead, c.distanceToCarAhead
}

// calculateRepulsionForce calcule la force de répulsion vectorielle
// Utilise la même logique que les piétons avec les forces vectorielles ( de ralentissement progressif)
func (c *Car) calculateRepulsionForce(myPos, otherPos utils.Point, distance float64) utils.Vector {
	// Si hors de la distance de détection, pas de force
	if distance <= 0 || distance > DETECTION_DISTANCE {
		return utils.VectorZero
	}

	// Vecteur de répulsion -> de l'autre voiture vers moi ( vecteur de répulsion pr rapport a l'autre voiture (devant moi) )
	repulsionDir := *utils.NewVectorFromPoints(otherPos, myPos)

	// Normaliser pour obtenir la direction
	repulsionDir = repulsionDir.Normalize()

	// Calculer l'intensité de la force selon la distance
	// Plus on est proche, plus la force est grande (inversement proportionnel) -> ralentissement progressif
	var forceMagnitude float64
	if distance <= SAFE_DISTANCE {
		// Très proche =force maximale ( ralentissement important)
		forceMagnitude = 1 // a augmenter si besoin de ralentir plus ( car voitures se touchent ) ou diminuer si slowdown trop brusque
	} else {
		// Interpolation -> force décroissante avec la distance et donc ralentissement progressif tjrs
		// Force inversement proportionnelle à la distance -> Plus on est proche, plus la force est grande
		ratio := 1.0 - ((distance - SAFE_DISTANCE) / (DETECTION_DISTANCE - SAFE_DISTANCE)) // valeur entre 0 et 1 suivant si on est proche ou loin
		forceMagnitude = ratio * 0.5
	}

	// Retourner le vecteur force = direction × intensité
	return repulsionDir.Scale(forceMagnitude)
}

// isCarAhead vérifie si une autre voiture est devant nous
// (même direction et même lane et position devant selon la direction)
func (c *Car) isCarAhead(other *Car) bool {
	if c.IsTurning {

		// Vérifier que l'autre voiture est bien devant nous selon la direction
		// pas oublier que l'axe Y augmente vers le bas de l'écran et x vers la droite
		switch c.Dir {
		case North: // On va vers le haut, donc Y diminue
			if other.LaneIndex != 1 {
				return false
			}
			if other.Direction() == West {
				return other.X < c.X // L'autre doit avoir un Y plus petit
			}
		case South: // On va vers le bas, donc Y augmente
			if other.LaneIndex != 0 {
				return false
			}
			if other.Direction() == East {
				return other.X > c.X // L'autre doit avoir un Y plus grand
			}
		case East: // On va vers la droite, donc X augmente
			if other.LaneIndex != 0 {
				return false
			}
			if other.Direction() == North {
				return other.Y < c.Y // L'autre doit avoir un X plus grand
			}
		case West: // On va vers la gauche, donc X diminue
			if other.LaneIndex != 1 {
				return false
			}
			if other.Direction() == South {
				return other.Y > c.Y // L'autre doit avoir un X plus petit
			}
		}

	}
	// Vérifier la direction
	if c.Dir != other.Dir {
		return false
	}

	// Vérifier la lane (voie)
	if c.LaneIndex != other.LaneIndex {
		return false
	}

	// Vérifier que l'autre voiture est bien devant nous selon la direction
	// pas oublier que l'axe Y augmente vers le bas de l'écran et x vers la droite
	switch c.Dir {
	case North: // On va vers le haut, donc Y diminue
		return other.Y < c.Y // L'autre doit avoir un Y plus petit
	case South: // On va vers le bas, donc Y augmente
		return other.Y > c.Y // L'autre doit avoir un Y plus grand
	case East: // On va vers la droite, donc X augmente
		return other.X > c.X // L'autre doit avoir un X plus grand
	case West: // On va vers la gauche, donc X diminue
		return other.X < c.X // L'autre doit avoir un X plus petit
	}

	return false
}

// distanceTo calcule la distance euclidienne à une autre voiture
func (c *Car) distanceTo(other *Car) float64 {
	dx := other.X - c.X
	dy := other.Y - c.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// adjustSpeed ajuste la vitesse en fonction des forces vectorielles appliquées
// logique identique au pieton -> la norme des forces détermine le ralentissement
func (c *Car) adjustSpeed() {
	// Vitesse de base qu'on a attribue aux voitures
	baseSpeed := 2.0

	// Calculer l'intensité totale des forces (norme du vecteur)
	forceIntensity := c.forces.Norm() // les forces ont été calculées par calculateRepulsionForce et on été ajouté dans le champ c.forces dans Percept " c.forces = c.forces.Add(repulsionForce)" et donc ici on récupère la norme de ces forces du champ c.forces

	// Convertir l'intensité de la force en réduction de vitesse
	// Plus la norme est grande, plus on ralentit
	// FORCE_SPEED_FACTOR détermine l'intensité du ralentissement (plus grand = freinage plus fort)
	targetSpeed := baseSpeed - (forceIntensity * FORCE_SPEED_FACTOR)

	// Contraindre la vitesse cible dans les limites acceptables
	targetSpeed = utils.Clamp(targetSpeed, MIN_SPEED, baseSpeed)

	// Ajustement progressif et fluide de la vitesse (obj d'avoir un effet smooth )
	// Évite les changements brusques pour un comportement réaliste
	speedDiff := targetSpeed - c.Speed
	if math.Abs(speedDiff) < 0.05 {
		// Changement négligeable -> adopter directement la vitesse cible
		c.Speed = targetSpeed
	} else if speedDiff > 0 {
		// Accélération douce
		c.Speed += 0.05
	} else {
		// Décélération plus rapide (important pour la sécurité)
		c.Speed -= 0.15
	}

	// Garantir que la vitesse finale reste dans les limites
	c.Speed = utils.Clamp(c.Speed, MIN_SPEED, baseSpeed)
}
