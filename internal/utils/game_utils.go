package utils

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type ChangeCharacRequest struct {
	Object             Object
	NewCharacteristics *PhysicalCharacteristics
}

type ReceiveCharacRequest struct {
	Object        Object
	ReceiveCharac chan *PhysicalCharacteristics
}

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
	Start(<-chan []Object, chan<- ChangeCharacRequest, <-chan []PhysicalCharacteristics)
	Percept([]PhysicalCharacteristics)
	Deliberate()
	Act(chan<- ChangeCharacRequest)
	AgentID() AgentID
	Alive() bool
	EndChan() chan int
}

// Interface standard d'un objet, permettant de dessiner l'objet ou de le percept
type Object interface {
	ObjectID() ObjectID
	Type() string
	Draw(chan<- ReceiveCharacRequest) (Asset, *ebiten.DrawImageOptions) //pour dessiner l'objet
	PhysicalCharacteristics() *PhysicalCharacteristics
	ChangePhysicalCharac(*PhysicalCharacteristics)
}

type PhysicalCharacteristics struct {
	PositionX float64 // world position
	PositionY float64
	Direction Direction
	State     string //pour connaiter l'etat du feu (par exemple)
	Angle     float64
	Type      string // type d'objet: car, pedestrian, trafficlight, ...
}
