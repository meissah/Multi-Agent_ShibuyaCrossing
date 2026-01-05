package pathfinding

import (
	"math"

	astar "github.com/beefsack/go-astar"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Constantes de la grille
const (
	WorldSize = 850
	CellSize  = 5
	GridDim   = 172
	Offset    = 50
)

// Tile représente une cellule de la grille pour A*
type Tile struct {
	X, Y         int
	Walkable     bool
	MovementCost float64
	grid         *Grid
}

// PathNeighbors définit les déplacements possibles (4 directions)
func (t *Tile) PathNeighbors() []astar.Pather {
	var neighbors []astar.Pather
	// Directions : Haut, Droite, Bas, Gauche
	dirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	for _, d := range dirs {
		nx, ny := t.X+d[0], t.Y+d[1]
		if nx >= 0 && nx < GridDim && ny >= 0 && ny < GridDim {
			node := t.grid.Tiles[ny][nx]
			if node.Walkable {
				neighbors = append(neighbors, node)
			}
		}
	}
	return neighbors
}

func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	return to.(*Tile).MovementCost
}

func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	toT := to.(*Tile)
	// Distance de Manhattan (optimale pour une grille orthogonale)
	return math.Abs(float64(t.X-toT.X)) + math.Abs(float64(t.Y-toT.Y))
}

// Grid contient la structure de navigation
type Grid struct {
	Tiles [GridDim][GridDim]*Tile
}

var NavGrid *Grid

// InitNavigationGrid initialise la carte des zones marchables
func InitNavigationGrid(isShibuya bool) {
	NavGrid = &Grid{}

	// Liste des zones autorisées (Trottoirs + Passages + Intersection)
	walkableRects := []struct{ x, y, w, h float64 }{
		// TROTTOIRS
		{245, -50, 55, 315}, {245, 435, 55, 415}, // Verticaux (G)
		{450, -50, 55, 315}, {450, 435, 55, 415}, // Verticaux (D)
		{-50, 245, 315, 55}, {435, 245, 415, 55}, // Horizontaux (H)
		{-50, 450, 315, 55}, {435, 450, 415, 55}, // Horizontaux (B)

		{290, 245, 170, 55}, // Nord (Lien NW <-> NE)
		{290, 450, 170, 55}, // Sud
		{245, 290, 55, 170}, // Ouest
		{450, 290, 55, 170}, // Est
	}

	// En mode Shibuya, on ajoute l'intersection centrale comme zone marchable
	if isShibuya {
		walkableRects = append(walkableRects, struct{ x, y, w, h float64 }{300, 300, 150, 150})
	}

	for y := 0; y < GridDim; y++ {
		for x := 0; x < GridDim; x++ {
			// Centre de la cellule en pixels
			px := float64(x*CellSize+CellSize/2) - Offset
			py := float64(y*CellSize+CellSize/2) - Offset

			isWalkable := false
			for _, r := range walkableRects {
				if px >= r.x && px <= r.x+r.w && py >= r.y && py <= r.y+r.h {
					isWalkable = true
					break
				}
			}

			NavGrid.Tiles[y][x] = &Tile{
				X:            x,
				Y:            y,
				Walkable:     isWalkable,
				MovementCost: 1.0,
				grid:         NavGrid,
			}
		}
	}
}

// FindPath calcule le chemin entre deux points pixels et retourne une liste de points
func FindPath(start, goal utils.Point) ([]utils.Point, bool) {
	if NavGrid == nil {
		InitNavigationGrid(false)
	}

	sx, sy := int(start.X()+50)/CellSize, int(start.Y()+50)/CellSize
	gx, gy := int(goal.X()+50)/CellSize, int(goal.Y()+50)/CellSize

	// Sécurité bornes
	if sx < 0 || sx >= GridDim || sy < 0 || sy >= GridDim ||
		gx < 0 || gx >= GridDim || gy < 0 || gy >= GridDim {
		return nil, false
	}

	startNode := NavGrid.Tiles[sy][sx]
	goalNode := NavGrid.Tiles[gy][gx]

	// Calcul A*
	path, _, found := astar.Path(startNode, goalNode)
	if !found {
		return nil, false
	}

	// Conversion en utils.Point et inversion (astar.Path renvoie de la fin vers le début)
	res := make([]utils.Point, len(path))
	for i := range path {
		t := path[len(path)-1-i].(*Tile)
		worldX := float64(t.X*CellSize+CellSize/2) - Offset
		worldY := float64(t.Y*CellSize+CellSize/2) - Offset
		res[i] = *utils.NewPoint(worldX, worldY)
	}

	return res, true
}
