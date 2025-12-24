package environment

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Zone rectangulaire repoussant les agents qui la traversent vers un côté défini.
// La force de répulsion augmente plus l'agent est loin du côté désigné, pour
// l'inciter à s'en rapprocher rapidement. Une zone tampon (buffer) est définie
// pour éviter que l'agent ne soit trop brusquement repoussé.
type RepulsionArea struct {
	utils.Object

	id            utils.ObjectID
	topLeft       utils.Point
	bottomRight   utils.Point
	forceStrength float64
	bufferSize    float64
	visible       bool
	active        bool
	isPedestrian  bool
	side          utils.Line

	minX, maxX, minY, maxY float64
}

func NewRepulsionArea(id utils.ObjectID, topLeft, bottomRight utils.Point, forceStrength, bufferSize float64) *RepulsionArea {
	ra := &RepulsionArea{
		id:            id,
		topLeft:       topLeft,
		bottomRight:   bottomRight,
		forceStrength: forceStrength,
		bufferSize:    bufferSize,
		visible:       true,
		active:        true,
		isPedestrian:  false,
	}

	// Le côté désigné est le bord supérieur de la zone rectangulaire.
	// On prend ici en compte les cas où les points topLeft et bottomRight
	// impliquent une rotation de la zone.
	deltaX := bottomRight.X() - topLeft.X()
	deltaY := bottomRight.Y() - topLeft.Y()

	var topRight utils.Point
	if deltaX*deltaY >= 0 {
		topRight = *utils.NewPoint(bottomRight.X(), topLeft.Y())
	} else {
		topRight = *utils.NewPoint(topLeft.X(), bottomRight.Y())
	}

	ra.side = *utils.NewLineFromPoints(topLeft, topRight)
	fmt.Println(ra.side)

	// Calcul des bornes pour optimiser IsPointInside
	ra.minX = math.Min(topLeft.X(), bottomRight.X())
	ra.maxX = math.Max(topLeft.X(), bottomRight.X())
	ra.minY = math.Min(topLeft.Y(), bottomRight.Y())
	ra.maxY = math.Max(topLeft.Y(), bottomRight.Y())

	return ra
}

func (ra *RepulsionArea) SetPedestrian(val bool) { ra.isPedestrian = val }
func (ra *RepulsionArea) Hide()                  { ra.visible = false }
func (ra *RepulsionArea) Show()                  { ra.visible = true }
func (ra *RepulsionArea) Activate()              { ra.active = true }
func (ra *RepulsionArea) Deactivate()            { ra.active = false }

func (ra RepulsionArea) ObjectID() utils.ObjectID   { return ra.id }
func (ra RepulsionArea) PositionX() float64         { return ra.topLeft.X() }
func (ra RepulsionArea) PositionY() float64         { return ra.topLeft.Y() }
func (ra RepulsionArea) Direction() utils.Direction { return utils.North }      // placeholder
func (ra RepulsionArea) State() string              { return "active" }         // placeholder
func (ra RepulsionArea) Type() string               { return "repulsion_area" } // placeholder
func (ra RepulsionArea) TheAngle() float64          { return 0.0 }              // placeholder

// Dessine la zone de répulsion si elle est marquée comme visible.
func (ra RepulsionArea) Draw(screen *ebiten.Image) {
	if !ra.visible {
		return
	}

	// Définition des couleurs selon l'état 'active'
	var drawColor color.RGBA
	if ra.active {
		if ra.isPedestrian {
			drawColor = color.RGBA{128, 0, 128, 100} // Violet transparent pour les piétons
		} else {
			drawColor = color.RGBA{255, 0, 0, 80} // Rouge pour les murs/routes
		}
	} else {
		drawColor = color.RGBA{0, 255, 0, 80} // Vert pour le passage
	}

	vector.FillRect(
		screen,
		float32(math.Min(ra.topLeft.X(), ra.bottomRight.X())),
		float32(math.Min(ra.topLeft.Y(), ra.bottomRight.Y())),
		float32(math.Abs(ra.bottomRight.X()-ra.topLeft.X())),
		float32(math.Abs(ra.bottomRight.Y()-ra.topLeft.Y())),
		drawColor,
		true,
	)
}

// Vérifie si un point donné se trouve à l'intérieur de la zone de répulsion.
func (ra RepulsionArea) IsPointInside(p utils.Point) bool {
	isWithinX := p.X() >= ra.minX && p.X() <= ra.maxX
	isWithinY := p.Y() >= ra.minY && p.Y() <= ra.maxY
	return isWithinX && isWithinY
}

// Calcule la force de répulsion exercée sur un point donné.
func (ra RepulsionArea) ForceAtPoint(p utils.Point) utils.Vector {
	if !ra.active || !ra.IsPointInside(p) {
		return utils.VectorZero
	}

	// Le point le plus proche sur le côté désigné est son projeté orthogonal.
	// On l'utilise pour calculer la distance et la direction de la force de répulsion.
	projection := p.ProjectOntoLine(ra.side)
	repulsion := *utils.NewVectorFromPoints(p, projection)
	distance := repulsion.Norm()

	// La force de répulsion est moindre selon la distance au côté désigné.
	repulsion = repulsion.Normalize()
	if distance > ra.bufferSize {
		repulsion = repulsion.Scale(ra.forceStrength)
	} else {
		repulsion = repulsion.Scale(ra.forceStrength * (distance / ra.bufferSize))
	}

	return repulsion
}

func (ra *RepulsionArea) SetAttractorLine(p1, p2 utils.Point) {
	ra.side = *utils.NewLineFromPoints(p1, p2)
}
