package vehicule_lights

// Vehicule light struct
type CarsLights struct {
	X, Y  float64 // world position
	State string  // "red" or "green" or "Yellow"
}

// IsRed retourne vrai si le feu est rouge
func (cl *CarsLight) IsRed() bool {
	return cl.State == "red"
}

// IsGreen retourne vrai si le feu est vert
func (cl *CarsLight) IsGreen() bool {
	return cl.State == "green"
}

// IsYellow retourne vrai si le feu est orange/jaune
func (cl *CarsLight) IsYellow() bool {
	return cl.State == "yellow"
}
