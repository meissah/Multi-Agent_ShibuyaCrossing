package utils

import (
	"sync/atomic"

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
	Start(*atomic.Pointer[[]Object])
	Percept(*atomic.Pointer[[]Object])
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
	Draw() (Asset, *ebiten.DrawImageOptions) //pour dessiner l'objet
	PhysicalCharacteristics() PhysicalCharacteristics
}

type PhysicalCharacteristics struct {
	PositionX float64 // world position
	PositionY float64
	Direction Direction
	State     string //pour connaiter l'etat du feu (par exemple)
	Angle     float64
}
