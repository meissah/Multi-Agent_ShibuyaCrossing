package utils

type Request interface{}

type GetNeighborsRequest struct {
	Coords    Point
	ReplyChan chan []Object
}

type AddObjectRequest struct {
	Object Object
}

type UpdateObjectRequest struct {
	Object Object
}
