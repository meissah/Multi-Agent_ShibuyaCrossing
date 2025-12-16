package environment

import (
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

	topLeft       utils.Point
	bottomRight   utils.Point
	forceStrength float64
	bufferSize    float64
	visible       bool
	active        bool
}

func NewRepulsionArea(topLeft, bottomRight utils.Point, forceStrength, bufferSize float64) *RepulsionArea {
	return &RepulsionArea{
		topLeft:       topLeft,
		bottomRight:   bottomRight,
		forceStrength: forceStrength,
		bufferSize:    bufferSize,
		visible:       true,
		active:        true,
	}
}

func (ra *RepulsionArea) Hide()       { ra.visible = false }
func (ra *RepulsionArea) Show()       { ra.visible = true }
func (ra *RepulsionArea) Activate()   { ra.active = true }
func (ra *RepulsionArea) Deactivate() { ra.active = false }

func (ra RepulsionArea) Direction() utils.Direction { return utils.North }      // placeholder
func (ra RepulsionArea) State() string              { return "active" }         // placeholder
func (ra RepulsionArea) Type() string               { return "repulsion_area" } // placeholder

func (ra RepulsionArea) Position() (float64, float64) {
	centerX := (ra.topLeft.X() + ra.bottomRight.X()) / 2
	centerY := (ra.topLeft.Y() + ra.bottomRight.Y()) / 2
	return centerX, centerY
}

// Dessine la zone de répulsion si elle est marquée comme visible.
func (ra RepulsionArea) Draw(screen *ebiten.Image) {
	if !ra.visible {
		return
	}

	colorRed := color.RGBA{255, 0, 0, 100}
	vector.FillRect(
		screen,
		float32(math.Min(ra.topLeft.X(), ra.bottomRight.X())),
		float32(math.Min(ra.topLeft.Y(), ra.bottomRight.Y())),
		float32(math.Abs(ra.bottomRight.X()-ra.topLeft.X())),
		float32(math.Abs(ra.bottomRight.Y()-ra.topLeft.Y())),
		colorRed,
		true,
	)
}

// Vérifie si un point donné se trouve à l'intérieur de la zone de répulsion.
func (ra RepulsionArea) IsPointInside(p utils.Point) bool {
	isWithinX := p.X() >= ra.topLeft.X() && p.X() <= ra.bottomRight.X()
	isWithinY := p.Y() >= ra.topLeft.Y() && p.Y() <= ra.bottomRight.Y()
	return isWithinX && isWithinY
}

// Calcule la force de répulsion exercée sur un point donné.
func (ra RepulsionArea) ForceAtPoint(p utils.Point) utils.Vector {
	if !ra.active || !ra.IsPointInside(p) {
		return utils.VectorZero
	}

	topLeft := ra.topLeft
	topRight := utils.NewPoint(ra.bottomRight.X(), ra.topLeft.Y())
	side := utils.NewLineFromPoints(topLeft, *topRight)

	// Le point le plus proche sur le côté désigné est son projeté orthogonal.
	// On l'utilise pour calculer la distance et la direction de la force de répulsion.
	projection := p.ProjectOntoLine(*side)
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
