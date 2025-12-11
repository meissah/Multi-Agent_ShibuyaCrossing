package utils

import "math"

// Représente un vecteur en deux dimensions, défini par ses composantes x et y.
//
// Le vecteur est conçu pour être read-only, aucune méthode ne modifie ses
// composantes après création. À la place, les opérations retournent de nouveaux
// vecteurs.
type Vector struct {
	x, y float64
}

var (
	VectorZero  = Vector{0, 0}
	VectorUnitX = Vector{1, 0}
	VectorUnitY = Vector{0, 1}
)

// Constructeur pour créer un nouveau vecteur
func NewVector(x, y float64) *Vector {
	// Vérification des valeurs Not a Number (NaN) et infinies
	isNan := math.IsNaN(x) || math.IsNaN(y)
	isPosInf := math.IsInf(x, 1) || math.IsInf(y, 1)
	isNegInf := math.IsInf(x, -1) || math.IsInf(y, -1)

	if isNan || isPosInf || isNegInf {
		return &VectorZero
	}

	return &Vector{x, y}
}

func (v Vector) X() float64 { return v.x }
func (v Vector) Y() float64 { return v.y }

// Calcul de la norme (longueur) du vecteur
func (v Vector) Norm() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}

// Normalisation du vecteur.
//
// Retourne un nouveau vecteur unitaire dans la même direction que le vecteur
// d'origine. Si la norme est nulle, retourne le vecteur nul.
func (v Vector) Normalize() Vector {
	norm := v.Norm()
	if norm == 0 {
		return VectorZero
	}

	newX := v.x / norm
	newY := v.y / norm
	return Vector{newX, newY}
}

// Addition de deux vecteurs.
//
// Retourne un nouveau vecteur résultant de l'addition du vecteur courant et
// du vecteur passé en paramètre.
func (v Vector) Add(other Vector) Vector {
	newX := v.x + other.x
	newY := v.y + other.y
	return Vector{newX, newY}
}

// Mise à l'échelle du vecteur par un scalaire.
//
// Retourne un nouveau vecteur résultant de la multiplication des composantes
// du vecteur par le scalaire donné.
func (v Vector) Scale(scalar float64) Vector {
	newX := v.x * scalar
	newY := v.y * scalar
	return Vector{newX, newY}
}

// Produit scalaire entre deux vecteurs.
//
// Retourne le produit scalaire du vecteur courant avec le vecteur passé en
// paramètre.
func (v Vector) Dot(other Vector) float64 {
	return v.x*other.x + v.y*other.y
}

// Calcul de l'angle (en radians) entre le vecteur et l'axe X positif.
//
// Retourne l'angle en radians dans l'intervalle [0, 2π). Si la norme du
// vecteur est nulle, retourne 0.
func (v Vector) Direction() float64 {
	norm := v.Norm()
	if norm == 0 {
		return 0
	}

	// Utilisation de la formule : cos(theta) = (v . x) / |v|
	// où x est le vecteur unitaire dans la direction de l'axe X positif (1,0)
	cosTheta := v.Dot(VectorUnitX) / norm
	cosTheta = Clamp(cosTheta, -1.0, 1.0)
	theta := math.Acos(cosTheta)

	// Passe l'angle de [0, π[ à [0, 2π[ en fonction du signe de y
	if v.y < 0 {
		theta = 2*math.Pi - theta
	}

	return theta
}

// Modification de la direction du vecteur.
//
// Retourne un nouveau vecteur avec la même norme que le vecteur courant mais
// orienté selon l'angle donné (en radians) par rapport à l'axe X positif.
func (v Vector) SetDirection(angle float64) Vector {
	norm := v.Norm()
	newX := norm * math.Cos(angle)
	newY := norm * math.Sin(angle)
	return Vector{newX, newY}
}
