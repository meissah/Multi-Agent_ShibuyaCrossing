package utils

type Request interface{}

type GetNeighborsRequest struct {
	Coords    Point
	ReplyChan chan []Object
}

type UpdateObjectRequest struct {
	Object Object
}
