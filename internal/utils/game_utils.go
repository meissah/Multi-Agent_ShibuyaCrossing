package utils

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Declaration du type AgentID
type AgentID ObjectID

type ObjectID uint

// Declaration asset (nom du fichier image)
type Asset int

// Direction enum pour voiture/pietons
type Direction int

const (
	North Direction = iota
	South
	East
	West
)

// Interface standard d'un agent (voiture/pieton/feu voiture/feu pieton)
type Agent interface {
	Object
	Start(*[]Object)
	Percept([]Object)
	Deliberate()
	Act()
	AgentID() AgentID
	Alive() bool
	EndChan() chan int
}

// Interface standard d'un objet, permettant de dessiner l'objet ou de le percept
type Object interface {
	ObjectID() ObjectID
	Type() string
	Direction() Direction
	TheAngle() float64
	PositionX() float64
	PositionY() float64
	Draw() (Asset, *ebiten.DrawImageOptions) //pour dessiner l'objet

}

type Drawable interface {
	Draw() (Asset, *ebiten.DrawImageOptions)
}
