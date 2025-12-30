package pedestrian

import (
	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Réprésente un piéton dans la simulation.
type Pedestrian struct {
	utils.Drawable
	utils.Agent

	Coords            utils.Point
	MaxSpeed          float64
	DesiredDirection  utils.Vector
	DesiredSeparation float64

	id                  utils.ObjectID
	velocity            utils.Vector
	syncChan            chan int
	endChan             chan int
	pedestriansAround   []Pedestrian
	repulsionAreasUnder []environment.RepulsionArea

	// pour statistiques
	spawnRecorded   bool // Pour éviter d'enregistrer plusieurs fois
	destroyRecorded bool // Pour éviter d'enregistrer plusieurs fois la destruction
}

// Constructeur de piéton
func NewPedestrian(id utils.ObjectID, syncChan chan int) *Pedestrian {
	return &Pedestrian{
		id:       id,
		syncChan: syncChan,
		endChan:  make(chan int),
	}
}

func (p Pedestrian) ObjectID() utils.ObjectID   { return p.id }
func (p Pedestrian) AgentID() utils.AgentID     { return utils.AgentID(p.id) }
func (p Pedestrian) SyncChan() chan int         { return p.syncChan }
func (p Pedestrian) EndChan() chan int          { return p.endChan }
func (p Pedestrian) PositionX() float64         { return p.Coords.X() }
func (p Pedestrian) PositionY() float64         { return p.Coords.Y() }
func (p Pedestrian) Direction() utils.Direction { return utils.North }      // Placeholder
func (p Pedestrian) TheAngle() float64          { return 0.0 }              // Placeholder
func (p Pedestrian) State() string              { return "walking" }        // Placeholder
func (p Pedestrian) Type() string               { return "pieton" }         // Placeholder
func (p Pedestrian) Alive() bool                { return !p.OutOfBounds() } // Placeholder

// Vérifie si le piéton est hors des limites de la zone de simulation.
func (p *Pedestrian) OutOfBounds() bool {
	isOutsideX := p.Coords.X() < -50 || p.Coords.X() > 800
	isOutsideY := p.Coords.Y() < -50 || p.Coords.Y() > 800
	isOut := isOutsideX || isOutsideY

	// Si le piéton sort, enregistrer la destruction (une seule fois) -> pour les statistiques
	if isOut && p.spawnRecorded && !p.destroyRecorded {
		stats.RecordPedestrianDestroy(uint(p.id))
		p.destroyRecorded = true
	}

	return isOut
}

// Méthode de dessin du piéton.
func (c *Pedestrian) Draw() (utils.Asset, *ebiten.DrawImageOptions) {

	X, Y := c.Coords.X(), c.Coords.Y()

	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(X, Y)
	return utils.Asset(1), op
}

// Lance la goroutine du piéton.
func (p *Pedestrian) Start(env *[]utils.Object) {
	// Enregistrer le spawn dans les métriques (une seule fois)
	if !p.spawnRecorded {
		dirStr := vectorToDirectionStr(p.DesiredDirection)
		stats.RecordPedestrianSpawn(uint(p.id), dirStr)
		p.spawnRecorded = true
	}
	// Boucle principale du piéton
	go func() {
		var step int
		for {
			select {
			case step = <-p.syncChan:
				p.Percept(*env)
				p.Deliberate()

				p.syncChan <- step
				<-p.syncChan

				p.Act()

				p.syncChan <- step
			case <-p.endChan:
				close(p.endChan)
				close(p.syncChan)
				return
			}
		}
	}()
}

// Convertit un vecteur en une chaîne de caractères représentant la direction principale -> besoin pour les statistiques
// car les piétons n'ont pas de direction fixe comme les voitures
func vectorToDirectionStr(v utils.Vector) string {
	// Normaliser le vecteur
	norm := v.Norm()
	if norm < 0.1 {
		return "stationnaire"
	}

	nx := v.X() / norm
	ny := v.Y() / norm

	// Déterminer la direction principale suivant les composantes
	if ny < -0.5 {
		return "North"
	} else if ny > 0.5 {
		return "South"
	} else if nx > 0.5 {
		return "East"
	} else if nx < -0.5 {
		return "West"
	}

	return "Diagonal"
}

// Perception du piéton
// Garde en mémoire les objets environnants pertinents
func (p *Pedestrian) Percept(env []utils.Object) {
	p.pedestriansAround = []Pedestrian{}

	for _, obj := range env {
		// Ajouter les piétons proches
		if otherPedestrian, ok := obj.(*Pedestrian); ok && otherPedestrian.id != p.id {
			diff := utils.NewVectorFromPoints(otherPedestrian.Coords, p.Coords)
			distance := diff.Norm()
			if distance > p.DesiredSeparation {
				continue
			}

			p.pedestriansAround = append(p.pedestriansAround, *otherPedestrian)
		}

		// Ajouter les routes sous le piéton
		if ra, ok := obj.(*environment.RepulsionArea); ok {
			if ra.IsPointInside(p.Coords) {
				p.repulsionAreasUnder = append(p.repulsionAreasUnder, *ra)
			}
		}
	}
}

// Délibération du piéton
func (p *Pedestrian) Deliberate() {
	p.velocity = utils.VectorZero

	// Le piéton se dirige vers sa direction désirée, à sa vitesse maximale
	goalVelocity := p.DesiredDirection.Normalize().Scale(p.MaxSpeed)
	p.velocity = goalVelocity.Add(p.velocity)

	// Le piéton s'éloigne des autres piétons autour de lui
	for _, otherPedestrian := range p.pedestriansAround {
		diff := utils.NewVectorFromPoints(otherPedestrian.Coords, p.Coords)
		distance := diff.Norm()

		// La force de répulsion est telle que les piétons mettent
		// plus d'effort à s'éloigner qu'à se diriger vers leur objectif
		repulsionFactor := 3.0
		repulsionDirection := diff.Normalize()
		repulsionForce := (p.DesiredSeparation - distance) / p.DesiredSeparation
		repulsion := repulsionDirection.Scale(repulsionForce * repulsionFactor * p.MaxSpeed)

		p.velocity = p.velocity.Add(repulsion)
	}

	// Appliquer les forces de l'environnement (zones de répulsion)
	for _, ra := range p.repulsionAreasUnder {
		force := ra.ForceAtPoint(p.Coords)
		p.velocity = p.velocity.Add(force)
	}

	// Limiter la vitesse
	if p.velocity.Norm() > p.MaxSpeed {
		p.velocity = p.velocity.Normalize().Scale(p.MaxSpeed)
	}
}

// Action du piéton
func (p *Pedestrian) Act() {
	// Met à jour la position du piéton en fonction de sa vélocité
	// fmt.Println(p.DesiredDirection)
	p.Coords = p.Coords.Translate(p.velocity)
}
