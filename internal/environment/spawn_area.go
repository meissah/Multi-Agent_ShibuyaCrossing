package environment

import "github.com/meissah/Multi-Agent_ShibuyaCrossing/internal/utils"

// Définition d'une zone de spawn rectangulaire
type SpawnArea struct {
	X, Y, Width, Height float64
}

// Collection de zones de spawn
type SpawnAreaCollection []*SpawnArea

// Génération d'un point aléatoire à l'intérieur de la zone de spawn
func (sa SpawnArea) RandomPoint() utils.Point {
	x := sa.X + utils.RandomFloatInRange(0, sa.Width)
	y := sa.Y + utils.RandomFloatInRange(0, sa.Height)
	return *utils.NewPoint(x, y)
}

// Génération d'un point aléatoire à l'intérieur de l'une des zones de spawn
func (sac SpawnAreaCollection) RandomPoint() utils.Point {
	areaIndex := utils.RandomIntInRange(0, len(sac)-1)
	return sac[areaIndex].RandomPoint()
}
