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
	if !FloatIsFinite(x) || !FloatIsFinite(y) {
		return &VectorZero
	}

	return &Vector{x, y}
}

// Constructeur pour créer un vecteur à partir de deux points
func NewVectorFromPoints(from, to Point) *Vector {
	x := to.X() - from.X()
	y := to.Y() - from.Y()
	return NewVector(x, y)
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

// Représente un point en deux dimensions, défini par ses coordonnées x et y.
//
// Le point est conçu pour être read-only, aucune méthode ne modifie ses
// coordonnées après création. À la place, les opérations retournent de nouveaux
// points.
type Point struct {
	x, y float64
}

// Constructeur pour créer un nouveau point
func NewPoint(x, y float64) *Point {
	if !FloatIsFinite(x) || !FloatIsFinite(y) {
		return nil
	}

	return &Point{x, y}
}

func (p Point) X() float64 { return p.x }
func (p Point) Y() float64 { return p.y }

// Translation du point par un vecteur.
//
// Retourne un nouveau point résultant du déplacement du point courant selon
// le vecteur donné.
func (p Point) Translate(v Vector) Point {
	newX := p.x + v.x
	newY := p.y + v.y
	return Point{newX, newY}
}

// Projection orthogonale du point sur une ligne.
//
// Retourne le point projeté sur la ligne donnée.
func (p Point) ProjectOntoLine(l Line) Point {
	lineVector := l.Vector()
	pointToStartVector := NewVectorFromPoints(l.StartPoint(), p)

	t := pointToStartVector.Dot(lineVector)
	newX := l.x1 + t*lineVector.x
	newY := l.y1 + t*lineVector.y
	return Point{newX, newY}
}

// Représente une ligne droite définie par deux points (x1, y1) et (x2, y2).
type Line struct {
	x1, y1 float64
	x2, y2 float64
}

// Constructeur pour créer une nouvelle ligne à partir de deux points.
// Retourne nil si les deux points sont identiques.
func NewLineFromPoints(p1, p2 Point) *Line {
	if p1.X() == p2.X() && p1.Y() == p2.Y() {
		return nil
	}

	return &Line{p1.X(), p1.Y(), p2.X(), p2.Y()}
}

// Constructeur pour créer une nouvelle ligne à partir d'un point et d'un
// vecteur directionnel.
func NewLineFromPointAndVector(p Point, v Vector) *Line {
	p2 := p.Translate(v)
	return &Line{p.X(), p.Y(), p2.X(), p2.Y()}
}

// Vecteur directionnel de la ligne.
func (l Line) Vector() Vector {
	vector := Vector{l.x2 - l.x1, l.y2 - l.y1}
	return vector.Normalize()
}

// Point de départ de la ligne.
func (l Line) StartPoint() Point {
	return Point{l.x1, l.y1}
}

// Point d'arrivée de la ligne.
func (l Line) EndPoint() Point {
	return Point{l.x2, l.y2}
}

// Forme géométrique quelconque.
type Shape interface {
	IsPointInside(p Point) bool
}

// Forme géométrique rectangulaire définie par son coin supérieur gauche,
// sa largeur et sa hauteur.
type Rectangle struct {
	Shape

	x, y          float64 // Coordonnées du coin supérieur gauche
	width, height float64 // Largeur et hauteur du rectangle
}

// Constructeur pour créer un nouveau rectangle
func NewRectangle(x, y, width, height float64) *Rectangle {
	if !FloatIsFinite(x) || !FloatIsFinite(y) || !FloatIsFinite(width) || !FloatIsFinite(height) {
		return nil
	}

	return &Rectangle{
		x:      x,
		y:      y,
		width:  width,
		height: height,
	}
}

func (r Rectangle) X() float64      { return r.x }
func (r Rectangle) Y() float64      { return r.y }
func (r Rectangle) Width() float64  { return r.width }
func (r Rectangle) Height() float64 { return r.height }

// Vérifie si un point donné se trouve à l'intérieur du rectangle.
func (r Rectangle) IsPointInside(p Point) bool {
	isWithinX := p.X() >= r.x && p.X() <= r.x+r.width
	isWithinY := p.Y() >= r.y && p.Y() <= r.y+r.height
	return isWithinX && isWithinY
}

type Triangle struct {
	Shape

	a, b, c Point // Sommets du triangle
}

func NewTriangle(a, b, c Point) *Triangle {
	return &Triangle{a: a, b: b, c: c}
}

// Vérifie si un point donné se trouve à l'intérieur du triangle en utilisant
// les determinants.
func (t Triangle) IsPointInside(p Point) bool {
	det := func(p1, p2, p3 Point) float64 {
		return (p2.X()-p1.X())*(p3.Y()-p1.Y()) - (p3.X()-p1.X())*(p2.Y()-p1.Y())
	}

	d1 := det(p, t.a, t.b)
	d2 := det(p, t.b, t.c)
	d3 := det(p, t.c, t.a)

	allPositive := (d1 > 0) && (d2 > 0) && (d3 > 0)
	allNegative := (d1 < 0) && (d2 < 0) && (d3 < 0)

	return allPositive || allNegative
}
