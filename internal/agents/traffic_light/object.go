package trafficlight

import (
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// TrafficLightObject est une représentation visuelle simple du feu, utilisée
// par le renderer (`game.Draw`). Elle implémente l'interface `utils.Object`.
type TrafficLightObject struct {
	utils.Object
	ID    string
	X, Y  float64
	W, H  float64
	dir   utils.Direction
	state string
	kind  string // "car" ou "pedestrian"
}

func NewTrafficLightObject(id string, x, y, w, h float64, dir utils.Direction, state string, kind string) *TrafficLightObject {
	return &TrafficLightObject{
		ID:    id,
		X:     x,
		Y:     y,
		W:     w,
		H:     h,
		dir:   dir,
		state: state,
		kind:  kind,
	}
}

func (t TrafficLightObject) ObjectID() utils.ObjectID   { return 0 }
func (t TrafficLightObject) Type() string               { return "traffic_light_" + t.kind }
func (t TrafficLightObject) Direction() utils.Direction { return t.dir }
func (t TrafficLightObject) TheAngle() float64          { return 0.0 }
func (t TrafficLightObject) PositionX() float64         { return t.X }
func (t TrafficLightObject) PositionY() float64         { return t.Y }
func (t TrafficLightObject) State() string              { return t.state }
func (t *TrafficLightObject) Size() (float64, float64)  { return t.W, t.H }

// SetState permet à l'agent de mettre à jour l'état visuel du feu.
func (t *TrafficLightObject) SetState(s string) { t.state = s }
