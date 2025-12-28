package messaging

import "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"

// Grille stockant des objets quelconques.
type Grid interface {
	AddObject(obj utils.Object)
	UpdateObject(obj utils.Object)
	GetNeighbors(coords utils.Point, radius int) []utils.Object
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
