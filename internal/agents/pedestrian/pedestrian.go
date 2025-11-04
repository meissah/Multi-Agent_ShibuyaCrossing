package pedestrian

import (
	"math/rand/v2"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/pedestrian_lights"
)

type Pedestrian struct {
	X, Y             float64 // world position
	TargetX, TargetY float64 // destination point
	Speed            float64 // magnitude of speed (e.g., 1.2 m/s)
	State            string  // "waiting", "walking"
}

type PedestrianEnvironment interface {
	nearbyPedestrianLights() []pedestrian_lights.PedestrianLight
}

func (p *Pedestrian) Act(env PedestrianEnvironment) {
	p.X += rand.Float64()*4 - 2 // move randomly within ±2px per frame
	p.Y += rand.Float64()*4 - 2
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
