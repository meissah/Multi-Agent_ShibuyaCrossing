package vehicule

import (
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/lights/vehicule_lights"
)

// Struct of a vehicule agent
type Vehicule struct {
	X, Y             float64
	Speed            float64 // e.g., 8–15 m/s (30–50 km/h)
	State            string  // "driving", "waiting", "turning"
	TargetX, TargetY float64
}

// Method to be implemented by Game struct, gives the necessary information of the environment to Vehicule
type VehiculeEnvironment interface {
	nearbyVehiculeLights(x, y float64) []vehicule_lights.VehiculeLight
}

// Method where vehicule percepts the environment then acts accordingly
func (c *Vehicule) Act(env VehiculeEnvironment) {
	//c.X += rand.Float64()*6 - 3 // move randomly within ±3px per frame
	c.Y += 1 //rand.Float64()*6 - 3
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.X > 750 {
		c.X = 750
	}
	if c.Y > 700 {
		c.Y = 700
	}
}
