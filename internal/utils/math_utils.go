package utils

import (
	"math"
	"math/rand"
)

// Restriction d'une valeur dans un intervalle donné [min, max]
func Clamp(value, min, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}

// Vérification des valeurs Not a Number (NaN) et infinies
func FloatIsFinite(value float64) bool {
	isNan := math.IsNaN(value)
	isPosInf := math.IsInf(value, 1)
	isNegInf := math.IsInf(value, -1)

	return !isNan && !isPosInf && !isNegInf
}

// Génération d'un nombre réel aléatoire dans une plage spécifique [min, max]
func RandomFloatInRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// Génération d'un entier aléatoire dans une plage spécifique [min, max]
func RandomIntInRange(min, max int) int {
	return rand.Intn(max-min+1) + min
}
