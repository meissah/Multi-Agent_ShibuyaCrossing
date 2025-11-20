package pedestrian

import (
	"math"
	rand "math/rand/v2"
	"time"

	// "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/pedestrian_lights"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/pathfinding"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

type Pedestrian struct {
	ID               AgentID
	X, Y             float64 // world position
	TargetX, TargetY float64 // destination point
	Speed            float64 // magnitude of speed (pixels per frame)
	PState           string  // "waiting", "walking"
	Path             []pathfinding.Vec2
	PathIndex        int
	WaitUntil        time.Time
	Done             bool // true après avoir atteint la destination, prêt à être retiré après WaitUntil

	AmIAlive bool
	syncChan chan int
	endChan  chan int
}

func NewPieton(id AgentID, x, y float64, syncChan chan int) *Pedestrian {
	return &Pedestrian{
		X:        x,
		Y:        y,
		ID:       id,
		syncChan: syncChan,
		AmIAlive: true,
		endChan:  make(chan int),
	}
}

func (ag Pedestrian) AgtID() AgentID {
	return ag.ID
}
func (ag Pedestrian) EndChan() chan int {
	return ag.endChan
}

// lance la goroutine de l'agent car
func (ag *Pedestrian) Start(env *[]Object, ObjectCh chan<- ObjectAddRequest) {

	go func() {
		var step int
		for {
			select {
			case step = <-ag.syncChan:

				ag.Percept(env)
				ag.Deliberate()
				ag.Act(ObjectCh)
				step++
				ag.syncChan <- step
			case <-ag.endChan:
				return
			}
		}
	}()

}

func (c *Pedestrian) Position() (float64, float64) {
	return c.X, c.Y
}

func (c *Pedestrian) Type() string {
	return "pieton"
}

func (c *Pedestrian) State() string {
	return c.PState
}

func (c *Pedestrian) Direction() Direction {
	return North
}

func (c *Pedestrian) Alive() bool {
	return c.AmIAlive
}

func (c *Pedestrian) Percept(env *[]Object) {}

func (c *Pedestrian) Deliberate() {}

// Method to be implemented by Game struct, gives the pedestrians the necessary information for his movement
// type PedestrianEnvironment interface {
// 	nearbyPedestrianLights(x, y float64) []pedestrian_lights.PedestrianLight
// }

// Act est appelé chaque frame pour mettre à jour le comportement du piéton.
// Comportement résumé :
//   - Si pas de Target défini, on choisit une cible aléatoire.
//   - Si WaitUntil est dans le futur, le piéton reste en attente (petit jitter visuel).
//   - Si pas de chemin calculé, on calcule le chemin A* vers la Target (FindPath).
//   - Si un chemin existe, on suit les waypoints un par un.
//   - Lorsque le piéton atteint le dernier waypoint, on met WaitUntil = now+3s,
//     on choisit immédiatement une nouvelle cible aléatoire et on calcule son chemin
//     pour pouvoir afficher la croix à l'endroit exact (mais le piéton n'avance
//     qu'après l'expiration de WaitUntil).
func (p *Pedestrian) Act(ActChn chan<- ObjectAddRequest) { /*env PedestrianEnvironment*/ // ) {
	if p.Speed == 0 {
		p.Speed = 2.0
	}

	// assign a random target if none
	if p.TargetX == 0 && p.TargetY == 0 {
		// preferer choisir la cible uniquement dans les rectangles de spawn (trottoirs rouges)
		minDist := 40.0
		placed := false
		// si SpawnRects est non vide, on tire dedans
		if len(pathfinding.SpawnRects) > 0 {
			for attempt := 0; attempt < 50; attempt++ {
				rx, ry, ok := pathfinding.RandomPointInRects(pathfinding.SpawnRects, rand.Float64)
				if !ok {
					break
				}
				if math.Hypot(rx-p.X, ry-p.Y) < minDist {
					continue
				}
				sx, sy := pathfinding.WorldToCell(p.X, p.Y)
				gx, gy := pathfinding.WorldToCell(rx, ry)
				if sx == gx && sy == gy {
					continue
				}
				p.TargetX = rx
				p.TargetY = ry
				placed = true
				break
			}
			// si après tentatives on n'a rien trouvé, forcer la cible au centre du premier spawn rect
			if !placed {
				r := pathfinding.SpawnRects[0]
				p.TargetX = float64(r.X + r.W/2)
				p.TargetY = float64(r.Y + r.H/2)
				placed = true
			}
		}
		// if SpawnRects was empty, fall back to selecting anywhere (previous behavior)
		if !placed && len(pathfinding.SpawnRects) == 0 {
			for attempt := 0; attempt < 50; attempt++ {
				rx := rand.Float64() * 750
				ry := rand.Float64() * 750
				if math.Hypot(rx-p.X, ry-p.Y) < minDist {
					continue
				}
				sx, sy := pathfinding.WorldToCell(p.X, p.Y)
				gx, gy := pathfinding.WorldToCell(rx, ry)
				if sx == gx && sy == gy {
					continue
				}
				p.TargetX = rx
				p.TargetY = ry
				placed = true
				break
			}
			if !placed {
				p.TargetX = rand.Float64() * 750
				p.TargetY = rand.Float64() * 750
			}
		}
	}

	// si le piéton est marqué Done, il attend jusqu'à WaitUntil avant d'être retiré
	//now := time.Now()
	if p.Done {
		// if p.WaitUntil.After(now) {
		//  // petit jitter visuel pendant l'attente
		//  p.X += rand.Float64()*0.2 - 0.1
		//  p.Y += rand.Float64()*0.2 - 0.1
		// }
		// // pas d'autre comportement (la suppression est gérée par Game.Update)
		p.AmIAlive = false
	}

	// compute a path if none (but if a path was already computed earlier, keep it)
	if len(p.Path) == 0 {
		// attempt to choose a target that yields a valid path. We will try candidates
		// drawn from SpawnRects (if available) and verify FindPath succeeds before
		// accepting the target. This prevents piétons d'apparaitre dans des zones
		// interdites ou avec des cibles inatteignables.
		if p.TargetX != 0 || p.TargetY != 0 {
			// if target already set (from earlier fallback), try to compute path once
			if path, found := pathfinding.FindPath(p.X, p.Y, p.TargetX, p.TargetY); found && len(path) > 0 {
				p.Path = path
				p.PathIndex = 0
				p.PState = "walking"
				// FIX: Force p.X, p.Y to the path start point to correct initial placement
				p.X = path[0].X
				p.Y = path[0].Y
				last := path[len(path)-1]
				p.TargetX = last.X
				p.TargetY = last.Y
			}
		} else {
			// pick candidates from SpawnRects (if any) and verify A* finds a path
			if len(pathfinding.SpawnRects) > 0 {
				for attempt := 0; attempt < 80; attempt++ {
					rx, ry, ok := pathfinding.RandomPointInRects(pathfinding.SpawnRects, rand.Float64)
					if !ok {
						break
					}
					// avoid same cell
					sx, sy := pathfinding.WorldToCell(p.X, p.Y)
					gx, gy := pathfinding.WorldToCell(rx, ry)
					if sx == gx && sy == gy {
						continue
					}
					// ensure the candidate is on a walkable cell
					if !pathfinding.IsWorldPointWalkable(rx, ry) {
						continue
					}
					// try to compute a path to that candidate
					if path, found := pathfinding.FindPath(p.X, p.Y, rx, ry); found && len(path) > 0 {
						p.Path = path
						p.PathIndex = 0
						p.PState = "walking"
						// FIX: Force p.X, p.Y to the path start point to correct initial placement
						p.X = path[0].X
						p.Y = path[0].Y
						last := path[len(path)-1]
						p.TargetX = last.X
						p.TargetY = last.Y
						break
					}
				}
				// if after attempts we still have no path, try centers of spawn rects deterministically
				if len(p.Path) == 0 {
					for _, r := range pathfinding.SpawnRects {
						rx := float64(r.X + r.W/2)
						ry := float64(r.Y + r.H/2)
						if !pathfinding.IsWorldPointWalkable(rx, ry) {
							continue
						}
						if path, found := pathfinding.FindPath(p.X, p.Y, rx, ry); found && len(path) > 0 {
							p.Path = path
							p.PathIndex = 0
							p.PState = "walking"
							// FIX: Force p.X, p.Y to the path start point to correct initial placement
							p.X = path[0].X
							p.Y = path[0].Y
							last := path[len(path)-1]
							p.TargetX = last.X
							p.TargetY = last.Y
							break
						}
					}
				}
			} else {
				// fallback: pick random world points (rare)
				for attempt := 0; attempt < 50 && len(p.Path) == 0; attempt++ {
					rx := rand.Float64() * 750
					ry := rand.Float64() * 750
					if pathfinding.IsWorldPointWalkable(rx, ry) {
						if path, found := pathfinding.FindPath(p.X, p.Y, rx, ry); found && len(path) > 0 {
							p.Path = path
							p.PathIndex = 0
							p.PState = "walking"
							// FIX: Force p.X, p.Y to the path start point to correct initial placement
							p.X = path[0].X
							p.Y = path[0].Y
							last := path[len(path)-1]
							p.TargetX = last.X
							p.TargetY = last.Y
							break
						}
					}
				}
			}
		}
	}

	// follow path
	if len(p.Path) > 0 && p.PathIndex < len(p.Path) {
		// REMOVED: Initial waypoint skip logic is now redundant as we explicitly set
		// p.X/p.Y to path[0] upon finding the path. The movement loop handles the
		// immediate advance to path[1] by setting dist=0.

		// Add safety check before accessing
		if p.PathIndex < len(p.Path) {
			next := p.Path[p.PathIndex]
			dx := next.X - p.X
			dy := next.Y - p.Y
			dist := math.Hypot(dx, dy)
			step := p.Speed
			if dist <= step || dist == 0 {
				p.X = next.X
				p.Y = next.Y
				p.PathIndex++
				if p.PathIndex >= len(p.Path) {
					// reached destination: marquer comme Done, attendre 3s avant suppression
					p.Path = nil
					p.PathIndex = 0
					p.WaitUntil = time.Now().Add(3 * time.Second)
					p.PState = "waiting"
					p.Done = true
					// conserver TargetX/TargetY sur le dernier waypoint pour l'affichage de la croix
				}
			} else {
				p.X += dx / dist * step
				p.Y += dy / dist * step
			}
		}
	} else {
		// small jitter to avoid being completely static
		p.X += rand.Float64()*0.5 - 0.25
		p.Y += rand.Float64()*0.5 - 0.25
	}

	// clamp to window
	if p.X < 0 {
		p.X = 0
	}
	if p.Y < 0 {
		p.Y = 0
	}
	if p.X > 750 {
		p.X = 750
	}
	if p.Y > 750 {
		p.Y = 750
	}

	RespChan := make(chan int)
	ActChn <- ObjectAddRequest{
		Object:   p,
		RespChan: RespChan,
	}
	<-RespChan
}
