package vehicule

import (
	"math/rand/v2"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/vehicule_lights"
)

type Vehicule struct {
	X, Y             float64
	Speed            float64 // e.g., 8–15 m/s (30–50 km/h)
	State            string  // "driving", "waiting", "turning"
	TargetX, TargetY float64
}

type VehiculeEnvironment interface {
	nearbyVehiculeLights() []vehicule_lights.VehiculeLight
}

func (c *Vehicule) Act(env VehiculeEnvironment) {
	c.X += rand.Float64()*6 - 3 // move randomly within ±3px per frame
	c.Y += rand.Float64()*6 - 3
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.X > 750 {
		c.X = 750
	}
	if c.Y > 750 {
		c.Y = 750
	}
}
