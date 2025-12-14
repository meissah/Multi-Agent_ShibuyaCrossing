package trafficlight

import (
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// TrafficLightObject est une représentation visuelle simple du feu, utilisée
// par le renderer (`game.Draw`). Elle implémente l'interface `utils.Object`.
type TrafficLightObject struct {
	utils.Object
	X, Y  float64
	dir   utils.Direction
	state string
}

func NewTrafficLightObject(x, y float64, dir utils.Direction, state string) *TrafficLightObject {
	return &TrafficLightObject{X: x, Y: y, dir: dir, state: state}
}

func (t TrafficLightObject) ObjectID() utils.ObjectID   { return 0 }
func (t TrafficLightObject) Type() string               { return "traffic_light" }
func (t TrafficLightObject) Direction() utils.Direction { return t.dir }
func (t TrafficLightObject) TheAngle() float64          { return 0.0 }
func (t TrafficLightObject) PositionX() float64         { return t.X }
func (t TrafficLightObject) PositionY() float64         { return t.Y }
func (t TrafficLightObject) State() string              { return t.state }

// SetState permet à l'agent de mettre à jour l'état visuel du feu.
func (t *TrafficLightObject) SetState(s string) { t.state = s }
