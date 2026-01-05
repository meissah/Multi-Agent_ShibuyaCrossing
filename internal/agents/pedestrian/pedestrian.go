package pedestrian

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/messaging"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Réprésente un piéton dans la simulation.
type Pedestrian struct {
	utils.Drawable
	utils.Agent

	Coords            utils.Point
	OldCoords         utils.Point
	MaxSpeed          float64
	DesiredDirection  utils.Vector
	DesiredSeparation float64
	// ---champs pour le Pathfinding ---
	Path             []utils.Point // Liste des points du chemin A*
	CurrentNodeIndex int           // Index du noeud que l'on vise actuellement
	FinalDestLine    *utils.Line   // La ligne d'arrivée (pour l'anti-entonnoir final)
	WanderOffset     utils.Vector  // Un décalage aléatoire unique à ce piéton pour ne pas marcher au centre

	DebugTarget utils.Point // Le point exact que le piéton vise
	//-----------------------------------
	id                  utils.ObjectID
	velocity            utils.Vector
	syncChan            chan int
	endChan             chan int
	gridChan            chan messaging.GridRequest
	pedestriansAround   []*Pedestrian
	repulsionAreasUnder []*environment.RepulsionArea

	StartTime time.Time

	// pour statistiques
	spawnRecorded   bool // Pour éviter d'enregistrer plusieurs fois
	destroyRecorded bool // Pour éviter d'enregistrer plusieurs fois la destruction
}

// Constructeur de piéton
func NewPedestrian(id utils.ObjectID, syncChan chan int, gridChan chan messaging.GridRequest) *Pedestrian {
	return &Pedestrian{
		id:       id,
		syncChan: syncChan,
		endChan:  make(chan int),
		gridChan: gridChan,
		// Initialiser un offset aléatoire léger (+/- 5px
		// va permetre d'éviter des comportements trop uniformes (file indienne genre)
		WanderOffset: *utils.NewVector(
			utils.RandomFloatInRange(-5, 5),
			utils.RandomFloatInRange(-5, 5),
		),
	}
}

func (p Pedestrian) ObjectID() utils.ObjectID   { return p.id }
func (p Pedestrian) AgentID() utils.AgentID     { return utils.AgentID(p.id) }
func (p Pedestrian) SyncChan() chan int         { return p.syncChan }
func (p Pedestrian) EndChan() chan int          { return p.endChan }
func (p Pedestrian) PositionX() float64         { return p.Coords.X() }
func (p Pedestrian) PositionY() float64         { return p.Coords.Y() }
func (p Pedestrian) Direction() utils.Direction { return utils.North } // Placeholder
func (p Pedestrian) TheAngle() float64          { return 0.0 }         // Placeholder
func (p Pedestrian) State() string              { return "walking" }   // Placeholder
func (p Pedestrian) Type() string               { return "pieton" }    // Placeholder
func (p Pedestrian) Alive() bool {

	return !p.OutOfBounds()
} // Placeholder

// Vérifie si le piéton est hors des limites de la zone de simulation.
func (p *Pedestrian) OutOfBounds() bool {

	if time.Since(p.StartTime) > 45*time.Second {
		if p.spawnRecorded && !p.destroyRecorded {
			stats.RecordPedestrianDestroy(uint(p.id))
			p.destroyRecorded = true
		}
		return true
	}
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
	op.GeoM.Scale(0.5, 0.5)
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
		p.StartTime = time.Now()
		for {
			select {
			case step = <-p.syncChan:
				p.Percept(*env)
				p.Deliberate()

				p.syncChan <- step
				<-p.syncChan

				p.Act()

				p.gridChan <- messaging.UpdateObjectRequest{Object: p}
				p.syncChan <- step
			case <-p.endChan:
				close(p.endChan)
				close(p.syncChan)
				p.gridChan <- messaging.DeleteObjectRequest{Object: p}
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
	p.pedestriansAround = []*Pedestrian{}
	p.repulsionAreasUnder = []*environment.RepulsionArea{}

	for _, obj := range env {
		if ra, ok := obj.(*environment.RepulsionArea); ok {
			if ra.IsPointInside(p.Coords) {
				p.repulsionAreasUnder = append(p.repulsionAreasUnder, ra)
			}
		}
	}

	// Obtenir les piétons voisins via la grille
	replyChan := make(chan []utils.Object)
	p.gridChan <- messaging.GetNeighborsRequest{
		Coords:    p.Coords,
		Radius:    1,
		ReplyChan: replyChan,
	}
	neighbors := <-replyChan

	for _, obj := range neighbors {
		if otherPedestrian, ok := obj.(*Pedestrian); ok && otherPedestrian.id != p.id {
			diff := utils.NewVectorFromPoints(otherPedestrian.Coords, p.Coords)
			distance := diff.Norm()
			if distance > p.DesiredSeparation {
				continue
			}

			p.pedestriansAround = append(p.pedestriansAround, otherPedestrian)
		}
	}
}

// Délibération du piéton
func (p *Pedestrian) Deliberate() {
	p.velocity = utils.VectorZero
	target := p.Coords

	if len(p.Path) > 0 {
		// Si on n'est pas au dernier noeud, on vérifie si on est trop loin du checkpoint actuel
		if p.CurrentNodeIndex < len(p.Path)-1 {
			currentTarget := p.Path[p.CurrentNodeIndex]
			distToTarget := utils.NewVectorFromPoints(p.Coords, currentTarget).Norm()

			// Seuil de 18 pixels
			if distToTarget > 18.0 {
				closestIdx := p.CurrentNodeIndex
				minDist := distToTarget

				// On cherche le noeud le plus proche parmi les noeuds futurs uniquement
				for i := p.CurrentNodeIndex + 1; i < len(p.Path)-1; i++ {
					d := utils.NewVectorFromPoints(p.Coords, p.Path[i]).Norm()
					if d < minDist {
						minDist = d
						closestIdx = i
					}
				}
				// On "saute" les checkpoints obsolètes
				p.CurrentNodeIndex = closestIdx
			}
		}
		// Calcul de la cible finale ou intermédiaire
		if p.CurrentNodeIndex >= len(p.Path)-1 && p.FinalDestLine != nil {
			// Direction vers la sortie de l'écran
			target = p.Coords.ProjectOntoLine(*p.FinalDestLine)
		} else if p.CurrentNodeIndex < len(p.Path) {
			// Cible avec offset pour éviter la file indienne
			baseTarget := p.Path[p.CurrentNodeIndex]
			target = baseTarget.Translate(p.WanderOffset)
		}

		// Force d'attraction vers la cible choisie
		desiredDir := utils.NewVectorFromPoints(p.Coords, target)
		dist := desiredDir.Norm()

		if dist > 0 {
			steering := desiredDir.Normalize().Scale(p.MaxSpeed)
			p.velocity = p.velocity.Add(steering)
		}

		// Passage au noeud suivant si on est assez proche (15px)
		if dist < 10.0 && p.CurrentNodeIndex < len(p.Path)-1 {
			p.CurrentNodeIndex++
		}
	}

	// le debug (Visualisation)
	p.DebugTarget = target

	// Le piéton s'éloigne des autres piétons autour de lui
	for _, otherPedestrian := range p.pedestriansAround {
		diff := utils.NewVectorFromPoints(otherPedestrian.Coords, p.Coords)
		distance := diff.Norm()

		// La force de répulsion est telle que les piétons mettent
		// plus d'effort à s'éloigner qu'à se diriger vers leur objectif
		repulsionFactor := 1.8 // ajustable
		repulsionDirection := diff.Normalize()
		if distance < 0.1 {
			distance = 0.1
		} // Éviter division par zéro
		forceStrength := (p.DesiredSeparation / distance) - 1

		if forceStrength > 0 {
			repulsion := repulsionDirection.Scale(forceStrength * repulsionFactor)
			p.velocity = p.velocity.Add(repulsion)
		}
	}

	// Appliquer les forces de l'environnement (zones de répulsion)
	// for _, ra := range p.repulsionAreasUnder {
	// 	force := ra.ForceAtPoint(p.Coords)
	// 	p.velocity = p.velocity.Add(force)
	// }

	// vérifier que la correction soit bonne

	// Répulsion Environnementale (Feux/Murs)
	for _, ra := range p.repulsionAreasUnder {
		force := ra.ForceAtPoint(p.Coords)
		// Les murs ont priorité absolue sur le pathfinding
		p.velocity = p.velocity.Add(force.Scale(3.0))
	}

	//Sécurité d'arrêt complet
	if p.velocity.Norm() < 0.1 {
		p.velocity = utils.VectorZero
	}

	// Limiter la vitesse
	if p.velocity.Norm() > p.MaxSpeed {
		p.velocity = p.velocity.Normalize().Scale(p.MaxSpeed)
	}
}

// Action du piéton
func (p *Pedestrian) Act() {
	// Met à jour la position du piéton en fonction de sa vélocité
	p.OldCoords = p.Coords
	p.Coords = p.Coords.Translate(p.velocity)
}
