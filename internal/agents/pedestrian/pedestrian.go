package pedestrian

import (
	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
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

func (p Pedestrian) OutOfBounds() bool {
	if p.Coords.X() < 0 || p.Coords.X() > 800 || p.Coords.Y() < 0 || p.Coords.Y() > 750 {
		return true
	}
	return false
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
		repulsionFactor := 1.0
		repulsionDirection := diff.Normalize()
		repulsionForce := (p.DesiredSeparation - distance) / p.DesiredSeparation
		repulsion := repulsionDirection.Scale(repulsionForce * repulsionFactor * p.MaxSpeed)

		// Si la répulsion est dans la direction opposée au mouvement désiré
		// on la décale légèrement pour imiter le choix social de contourner
		// l'autre piéton.
		// goalAngle := p.DesiredDirection.Direction()
		// repulsionAngle := repulsion.Direction()

		// if math.Abs(float64(goalAngle-repulsionAngle)) > math.Pi*0.9 {
		// 	repulsion = repulsion.SetDirection(goalAngle + (math.Pi * 0.9))
		// }

		p.velocity = p.velocity.Add(repulsion)
	}

	// Appliquer les forces de l'environnement (zones de répulsion)
	for _, ra := range p.repulsionAreasUnder {
		force := ra.ForceAtPoint(p.Coords)
		p.velocity = p.velocity.Add(force.Scale(.10))
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
