package utils

// Declaration du type AgentID
type AgentID uint

// Declaration du type de la requête pour ajouter un object dans l'environnement
type ObjectAddRequest struct {
	Object   Object
	RespChan chan int
}

// Direction enum pour voiture/pietons
type Direction int

const (
	North Direction = iota // monte (y diminue)
	South                  // descend (y augmente)
	East                   // va à droite (x augmente)
	West                   // va à gauche (x diminue)
)

// Interface standard d'un agent (voiture/pieton/feu voiture/feu pieton)
type Agent interface {
	Start(*[]Object, chan<- ObjectAddRequest)
	Percept(*[]Object)
	Deliberate()
	Act(chan<- ObjectAddRequest)
	AgtID() AgentID
	Alive() bool
	EndChan() chan int
}

// Interface standard d'un objet, permettant de dessiner l'objet en question par Draw() de ubiten
type Object interface {
	Position() (float64, float64) // world position
	Direction() Direction
	Type() string  //to know what item to draw
	State() string //pour savoir quel etat dessiner (lumière éteinte/allumée par exemple)
}
