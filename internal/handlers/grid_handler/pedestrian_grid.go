package grid_handler

import (
	"fmt"

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
	requestChan chan utils.Request
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

	fmt.Println(cells)

	return &PedestrianGrid{
		cellSize:    cellSize,
		cells:       cells,
		requestChan: make(chan utils.Request),
	}
}

func (grid PedestrianGrid) RequestChan() chan utils.Request {
	return grid.requestChan
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
	cellX := int((p.X() + 50) / grid.cellSize)
	cellY := int((p.Y() + 50) / grid.cellSize)
	return cellX, cellY
}

// Démarre le gestionnaire de la grille des piétons.
func (grid *PedestrianGrid) Start() {
	go func() {
		for req := range grid.requestChan {
			if updateReq, ok := req.(utils.UpdateObjectRequest); ok {
				grid.handleUpdateRequest(updateReq)
			} else if neighborsReq, ok := req.(utils.GetNeighborsRequest); ok {
				grid.handleNeighborRequest(neighborsReq)
			}
		}
	}()
}

// Met à jour la position du piéton dans la grille.
func (grid *PedestrianGrid) handleUpdateRequest(req utils.UpdateObjectRequest) {
	// Extraire le piéton de la requête
	ped, ok := req.Object.(*pedestrian.Pedestrian)
	if !ok {
		return
	}

	newCellX, newCellY := grid.CoordsToCellIdx(ped.Coords)
	oldCellX, oldCellY := grid.CoordsToCellIdx(ped.OldCoords)

	// Changer de cellule si nécessaire
	if newCellX != oldCellX || newCellY != oldCellY {
		oldCell := grid.CellAt(oldCellX, oldCellY)
		newCell := grid.CellAt(newCellX, newCellY)
		if oldCell == nil || newCell == nil {
			return
		}

		oldCell.Remove(ped)
		newCell.Add(ped)
	}
}

// Récupère les piétons voisins aux coordonnées données.
func (grid *PedestrianGrid) handleNeighborRequest(req utils.GetNeighborsRequest) {
	cellX, cellY := grid.CoordsToCellIdx(req.Coords)
	neighbors := []utils.Object{}

	// Parcourir les cellules voisines
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			neighborCell := grid.CellAt(cellX+dx, cellY+dy)
			if neighborCell == nil {
				continue
			}

			for _, ped := range *neighborCell {
				neighbors = append(neighbors, ped)
			}
		}
	}

	req.ReplyChan <- neighbors
}
