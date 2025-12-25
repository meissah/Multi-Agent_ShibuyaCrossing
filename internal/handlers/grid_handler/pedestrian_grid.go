package grid_handler

import (
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

type PedestrianCell []*pedestrian.Pedestrian

// Ajoute un piéton à la cellule.
func (pc *PedestrianCell) Add(p *pedestrian.Pedestrian) {
	*pc = append(*pc, p)
}

// Retire un piéton de la cellule.
func (pc *PedestrianCell) Remove(p *pedestrian.Pedestrian) {
	for i, pedestrian := range *pc {
		if pedestrian == p {
			*pc = append((*pc)[:i], (*pc)[i+1:]...)
			break
		}
	}
}

// Grille gérant la position des piétons pour des requêtes efficaces.
type PedestrianGrid struct {
	cellSize    float64
	cells       [][]PedestrianCell
	requestChan chan Request
}

// Constructeur de la grille des piétons.
func NewPedestrianGrid(width, height, cellSize float64) *PedestrianGrid {
	numCellsX := int(width / cellSize)
	numCellsY := int(height / cellSize)

	cells := make([][]PedestrianCell, numCellsX)
	for i := range cells {
		cells[i] = make([]PedestrianCell, numCellsY)
		for j := range cells[i] {
			cells[i][j] = PedestrianCell{}
		}
	}

	return &PedestrianGrid{
		cellSize:    cellSize,
		cells:       cells,
		requestChan: make(chan Request),
	}
}

// Récupère la taille des cellules.
func (grid PedestrianGrid) CellSize() float64 {
	return grid.cellSize
}

// Récupère la cellule aux indices donnés.
func (grid PedestrianGrid) CellAt(cellX, cellY int) *PedestrianCell {
	if cellX < 0 || cellX >= len(grid.cells) || cellY < 0 || cellY >= len(grid.cells[0]) {
		return nil
	}

	return &grid.cells[cellX][cellY]
}

// Convertit des coordonnées en indices de cellule.
func (grid PedestrianGrid) CoordsToCellIdx(p utils.Point) (int, int) {
	cellX := int(p.X() / grid.cellSize)
	cellY := int(p.Y() / grid.cellSize)
	return cellX, cellY
}

// Démarre le gestionnaire de la grille des piétons.
func (grid *PedestrianGrid) Start() {

	go func() {
		for req := range grid.requestChan {
			req.handle(grid)
		}
	}()
}

// Requête agissant sur la grille des piétons.
type Request interface {
	handle(*PedestrianGrid)
}

// Requête pour mettre à jour la position d'un piéton dans la grille.
type UpdatePedestrianRequest struct {
	Request

	Pedestrian *pedestrian.Pedestrian
}

// Met à jour la position du piéton dans la grille.
func (req UpdatePedestrianRequest) handle(grid *PedestrianGrid) {
	cellX, cellY := grid.CoordsToCellIdx(req.Pedestrian.Coords)
	oldCellX, oldCellY := grid.CoordsToCellIdx(req.Pedestrian.OldCoords)

	// Changer de cellule si nécessaire
	if cellX != oldCellX || cellY != oldCellY {
		grid.CellAt(oldCellX, oldCellY).Remove(req.Pedestrian)
		grid.CellAt(cellX, cellY).Add(req.Pedestrian)
	}
}
