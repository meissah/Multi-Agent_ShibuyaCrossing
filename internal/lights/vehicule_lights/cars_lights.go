package vehicule_lights

// Vehicule light struct
type VehiculeLight struct {
	X, Y  float64 // world position
	State string  // "red" or "green" or "Yellow"
}

// IsRed retourne vrai si le feu est rouge
func (carl *VehiculeLight) IsRed() bool {
	return carl.State == "red"
}

// IsGreen retourne vrai si le feu est vert
func (carl *VehiculeLight) IsGreen() bool {
	return carl.State == "green"
}

// IsYellow retourne vrai si le feu est orange/jaune
func (carl *VehiculeLight) IsYellow() bool {
	return carl.State == "yellow"
}
