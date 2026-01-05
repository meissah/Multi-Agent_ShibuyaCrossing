package grid_handler

import (
	"fmt"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/pedestrian"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/messaging"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Cellule contenant des piétons.
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
	requestChan chan messaging.GridRequest
	endChan     chan int
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
		requestChan: make(chan messaging.GridRequest),
	}
}

// Channel permettant de communiquer avec la grille, via des requêtes.
func (grid PedestrianGrid) RequestChan() chan messaging.GridRequest {
	return grid.requestChan
}

// Taille des cellules.
func (grid PedestrianGrid) CellSize() float64 {
	return grid.cellSize
}

// Cellule aux indices donnés.
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
		grid.endChan = make(chan int)
		for {
			select {
			case req := <-grid.requestChan:
				req.Handle(grid)
				fmt.Println("Bouclé-ing dans la grille")
			case <-grid.endChan:
				fmt.Println("Arrivée au end de pdGrid")
				return
			}
		}
	}()
}

func (grid *PedestrianGrid) Stop() {
	fmt.Println("Arrivée au beg de stop de pdGrid")
	grid.endChan <- 1
	fmt.Println("Arrivée au end de stop de pdGrid")
}

// Ajoute un piéton à la grille.
func (grid *PedestrianGrid) AddObject(obj utils.Object) {
	ped, ok := obj.(*pedestrian.Pedestrian)
	if !ok {
		return
	}

	cellX, cellY := grid.CoordsToCellIdx(ped.Coords)
	cell := grid.CellAt(cellX, cellY)
	if cell == nil {
		return
	}

	cell.Add(ped)
}

// Met à jour la position du piéton dans la grille.
func (grid *PedestrianGrid) UpdateObject(obj utils.Object) {
	ped, ok := obj.(*pedestrian.Pedestrian)
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

// Supprime un piéton de la grille.
func (grid *PedestrianGrid) DeleteObject(obj utils.Object) {
	ped, ok := obj.(*pedestrian.Pedestrian)
	if !ok {
		return
	}

	cellX, cellY := grid.CoordsToCellIdx(ped.Coords)
	cell := grid.CellAt(cellX, cellY)
	if cell == nil {
		return
	}

	cell.Remove(ped)
}

// Récupère les piétons voisins aux coordonnées données.
func (grid PedestrianGrid) GetNeighbors(coords utils.Point, radius int) []utils.Object {
	cellX, cellY := grid.CoordsToCellIdx(coords)
	neighbors := []utils.Object{}

	// Parcourir les cellules voisines (jusqu'au rayon spécifié)
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			neighborCell := grid.CellAt(cellX+dx, cellY+dy)
			if neighborCell == nil {
				continue
			}

			for _, ped := range *neighborCell {
				neighbors = append(neighbors, ped)
			}
		}
	}

	return neighbors
}

// Récupère les piétons dans un rectangle défini par deux points.
func (grid PedestrianGrid) GetInRect(topLeft utils.Point, bottomRight utils.Point) []utils.Object {
	startCellX, startCellY := grid.CoordsToCellIdx(topLeft)
	endCellX, endCellY := grid.CoordsToCellIdx(bottomRight)
	objects := []utils.Object{}

	// Assurer que les indices sont dans le bon ordre
	minCellX := utils.IntMin(startCellX, endCellX)
	maxCellX := utils.IntMax(startCellX, endCellX)
	minCellY := utils.IntMin(startCellY, endCellY)
	maxCellY := utils.IntMax(startCellY, endCellY)

	for cellX := minCellX; cellX <= maxCellX; cellX++ {
		for cellY := minCellY; cellY <= maxCellY; cellY++ {
			cell := grid.CellAt(cellX, cellY)
			if cell == nil {
				continue
			}

			for _, ped := range *cell {
				objects = append(objects, ped)
			}
		}
	}

	return objects
}
