package utils

// Declaration du type AgentID
type AgentID uint

type ObjectAddRequest struct {
	Object   Object
	RespChan chan int
}

// Direction enum pour la voiture
type Direction int

const (
	North Direction = iota // monte (y diminue)
	South                  // descend (y augmente)
	East                   // va à droite (x augmente)
	West                   // va à gauche (x diminue)
)

// Interface standard d'un agent, peu importe lequel
type Agent interface {
	Start(*[]Object, chan<- ObjectAddRequest)
	Percept(*[]Object)
	Deliberate()
	Act(chan<- ObjectAddRequest)
	AgtID() AgentID
	Alive() bool
}

type Object interface {
	Position() (float64, float64) // world position
	Direction() Direction
	Type() string  //to know what item to draw
	State() string //pour savoir quel etat dessiner (lumière éteinte/allumée par exemple)
}
