package messaging

import "github.com/meissah/Multi-Agent_ShibuyaCrossing/internal/utils"

// Grille stockant des objets quelconques.
type Grid interface {
	AddObject(obj utils.Object)
	UpdateObject(obj utils.Object)
	DeleteObject(obj utils.Object)
	GetNeighbors(coords utils.Point, radius int) []utils.Object
	GetInRect(topLeft utils.Point, bottomRight utils.Point) []utils.Object
}

// Requête agissant sur une grille.
type GridRequest interface {
	Handle(grid Grid)
}

// Requête d'ajout d'un objet à la grille.
type AddObjectRequest struct {
	GridRequest
	Object utils.Object
}

func (req AddObjectRequest) Handle(grid Grid) {
	grid.AddObject(req.Object)
}

// Requête de mise à jour d'un objet dans la grille.
type UpdateObjectRequest struct {
	GridRequest
	Object utils.Object
}

func (req UpdateObjectRequest) Handle(grid Grid) {
	grid.UpdateObject(req.Object)
}

// Requête de suppression d'un objet de la grille.
type DeleteObjectRequest struct {
	GridRequest
	Object utils.Object
}

func (req DeleteObjectRequest) Handle(grid Grid) {
	grid.DeleteObject(req.Object)
}

// Requête de récupération des voisins autour de coordonnées données.
type GetNeighborsRequest struct {
	GridRequest
	Coords    utils.Point
	Radius    int
	ReplyChan chan []utils.Object
}

func (req GetNeighborsRequest) Handle(grid Grid) {
	neighbors := grid.GetNeighbors(req.Coords, req.Radius)
	req.ReplyChan <- neighbors
}

// Requête de récupération des objets dans un rectangle défini par deux points.
type GetInRectRequest struct {
	GridRequest
	TopLeft     utils.Point
	BottomRight utils.Point
	ReplyChan   chan []utils.Object
}

func (req GetInRectRequest) Handle(grid Grid) {
	objects := grid.GetInRect(req.TopLeft, req.BottomRight)
	req.ReplyChan <- objects
}
