package pedestrian

import (
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/pedestrian_lights"
)

// Pedestrian agent struct
type Pedestrian struct {
	X, Y             float64 // world position
	TargetX, TargetY float64 // destination point
	Speed            float64 // magnitude of speed (e.g., 1.2 m/s)
	State            string  // "waiting", "walking"
}

// Method to be implemented by Game struct, gives the pedestrians the necessary information for his movement
type PedestrianEnvironment interface {
	nearbyPedestrianLights(x, y float64) []pedestrian_lights.PedestrianLight
}

// Method where pedestrian percepts the environment then acts accordingly
func (p *Pedestrian) Act(env PedestrianEnvironment) {
	p.X += 2 //rand.Float64()*4 - 2 // move randomly within ±2px per frame
	//p.Y += rand.Float64()*4 - 2
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
}
