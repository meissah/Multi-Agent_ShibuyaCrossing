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
	Shape         utils.Shape
	PushVector    utils.Vector // Direction vers laquelle la zone éjecte

	minX, maxX, minY, maxY float64
	// borne pour le visuel
	vMinX, vMaxX, vMinY, vMaxY float64
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

	ra.Shape = utils.NewRectangle(ra.minX, ra.minY, ra.maxX-ra.minX, ra.maxY-ra.minY)

	ra.vMinX, ra.vMaxX = ra.minX, ra.maxX
	ra.vMinY, ra.vMaxY = ra.minY, ra.maxY

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

func (ra *RepulsionArea) ExpandCollision(margin float64) {
	ra.minX -= margin
	ra.maxX += margin
	ra.minY -= margin
	ra.maxY += margin
	// Met à jour la forme physique utilisée par IsPointInside
	ra.Shape = utils.NewRectangle(ra.minX, ra.minY, ra.maxX-ra.minX, ra.maxY-ra.minY)
}

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
		float32(ra.vMinX),
		float32(ra.vMinY),
		float32(ra.vMaxX-ra.vMinX),
		float32(ra.vMaxY-ra.vMinY),
		drawColor,
		true,
	)
}

func (ra *RepulsionArea) ExpandSide(side string, margin float64) {
	switch side {
	case "N":
		ra.minY -= margin
	case "S":
		ra.maxY += margin
	case "W":
		ra.minX -= margin
	case "E":
		ra.maxX += margin
	}

	ra.Shape = utils.NewRectangle(ra.minX, ra.minY, ra.maxX-ra.minX, ra.maxY-ra.minY)
}

func NewTriangleRepulsionArea(id utils.ObjectID, a, b, c utils.Point, push utils.Vector, strength float64) *RepulsionArea {
	return &RepulsionArea{
		id:            id,
		Shape:         utils.NewTriangle(a, b, c),
		PushVector:    push,
		forceStrength: strength,
		active:        true,
		isPedestrian:  true,
		visible:       false,
		topLeft:       a, // Point de référence
		bottomRight:   c, // Point de référence
	}
}

func (ra RepulsionArea) IsPointInside(p utils.Point) bool {
	if ra.Shape == nil {
		return false
	}
	return ra.Shape.IsPointInside(p)
}

func (ra RepulsionArea) ForceAtPoint(p utils.Point) utils.Vector {
	if !ra.active || !ra.IsPointInside(p) {
		return utils.VectorZero
	}
	// CAS triangle
	if ra.PushVector.Norm() > 0 {
		return ra.PushVector.Scale(ra.forceStrength)
	}

	// CAS rectangle (Passage piéton)
	projection := p.ProjectOntoLine(ra.side)
	repulsion := *utils.NewVectorFromPoints(p, projection)
	distance := repulsion.Norm()

	if distance < 0.1 {
		return utils.VectorZero
	}

	// Calcul de la force selon la distance (plus on est loin du trottoir, plus on est poussé)
	force := repulsion.Normalize().Scale(ra.forceStrength)
	return force
}

func (ra *RepulsionArea) SetAttractorLine(p1, p2 utils.Point) {
	ra.side = *utils.NewLineFromPoints(p1, p2)
}
