package utils

import (
	"math"
	"math/rand/v2"
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

// Génération de nombres entiers aléatoires dans un intervalle donné
func RandomIntInRange(min, max int) int {
	return rand.IntN(max-min+1) + min
}

// Génération de nombres réels aléatoires dans un intervalle donné
func RandomFloatInRange(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}
