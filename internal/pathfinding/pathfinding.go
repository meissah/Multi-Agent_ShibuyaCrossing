package pathfinding

import (
	"math"

	astar "github.com/beefsack/go-astar"
)

// Vec2 représente une position en coordonnées "monde" (float64),
// utilisée pour retourner des centres de cellules (waypoints).
type Vec2 struct {
	X, Y float64
}

// Tile représente une case de la grille et implémente l'interface astar.Pather.
// Chaque Tile connaît sa position (X,Y), si elle est franchissable (Walkable)
// et le coût de déplacement pour entrer dans cette case (MovementCost).
type Tile struct {
	X, Y         int
	Walkable     bool
	MovementCost float64
	grid         *Grid
}

// PathNeighbors retourne les voisins orthogonaux valides (haut/droite/bas/gauche).
func (t *Tile) PathNeighbors() []astar.Pather {
	var out []astar.Pather
	dirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for _, d := range dirs {
		nx := t.X + d[0]
		ny := t.Y + d[1]
		if nx < 0 || ny < 0 || nx >= t.grid.Width || ny >= t.grid.Height {
			continue
		}
		nt := t.grid.Tiles[ny][nx]
		if nt != nil && nt.Walkable {
			out = append(out, nt)
		}
	}
	return out
}

// PathNeighborCost retourne le coût exact pour aller vers le voisin `to`.
// Ici on retourne simplement le MovementCost de la case de destination.
func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	if other, ok := to.(*Tile); ok {
		return other.MovementCost
	}
	return 1.0
}

// PathEstimatedCost est l'heuristique (coût estimé) entre deux cases.
// Nous utilisons la distance de Manhattan (admissible pour mouvements orthogonaux).
func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	if other, ok := to.(*Tile); ok {
		dx := float64(t.X - other.X)
		dy := float64(t.Y - other.Y)
		return math.Abs(dx) + math.Abs(dy) // Manhattan
	}
	return 0
}

// Grid contient la grille de Tile. Tiles est indexée [y][x].
type Grid struct {
	Width, Height int
	Tiles         [][]*Tile // indexed [y][x]
	CellSize      int
}

// Rect représente un rectangle en coordonnées pixels (X,Y,W,H).
type Rect struct {
	X, Y, W, H int
}

// Default world / cell constants (utilisées dans FindPath si rien d'autre)
const DefaultWorldSize = 750
const DefaultCellSize = 10

// AllowedRects définit les rectangles où le déplacement est autorisé.
// Si vide, tout est autorisé.
var AllowedRects []Rect

// SpawnRects définit les rectangles où les piétons peuvent apparaître/avoir leur destination.
var SpawnRects []Rect

// SetAllowedRects définit les zones franchissables pour le pathfinding.
func SetAllowedRects(rects []Rect) { AllowedRects = rects }

// SetSpawnRects définit les zones utilisables pour l'apparition/destination.
func SetSpawnRects(rects []Rect) { SpawnRects = rects }

// WorldToCell convertit des coordonnées monde (pixels) en indices de cellule
// en utilisant DefaultCellSize.
func WorldToCell(x, y float64) (int, int) {
	return int(x) / DefaultCellSize, int(y) / DefaultCellSize
}

// PointInRect vérifie si une coordonnée monde (pixels) est à l'intérieur d'un Rect.
func PointInRect(x, y float64, r Rect) bool {
	return int(x) >= r.X && int(x) < r.X+r.W && int(y) >= r.Y && int(y) < r.Y+r.H
}

// RandomPointInRects choisit aléatoirement un point à l'intérieur d'un des rectangles donnés.
// Si la liste est vide, renvoie (0,0,false).
func RandomPointInRects(rects []Rect, rnd func() float64) (float64, float64, bool) {
	if len(rects) == 0 {
		return 0, 0, false
	}
	// choisir un rectangle aléatoire
	idx := int(rnd() * float64(len(rects)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(rects) {
		idx = len(rects) - 1
	}
	r := rects[idx]
	x := float64(r.X) + rnd()*float64(r.W)
	y := float64(r.Y) + rnd()*float64(r.H)
	return x, y, true
}

// NewGrid crée une grille Width x Height avec toutes les cases franchissables
// et un MovementCost par défaut (1.0). CellSize définit la taille d'une case
// en pixels (utilisée pour convertir entre coordonnées monde et indices).
func NewGrid(width, height, cellSize int) *Grid {
	g := &Grid{Width: width, Height: height, CellSize: cellSize}
	g.Tiles = make([][]*Tile, height)
	for y := 0; y < height; y++ {
		row := make([]*Tile, width)
		for x := 0; x < width; x++ {
			row[x] = &Tile{X: x, Y: y, Walkable: true, MovementCost: 1.0, grid: g}
		}
		g.Tiles[y] = row
	}
	return g
}

// NewGridFromRects crée une grille où seules les cases dont le centre tombe
// dans au moins un des rects fournis sont franchissables.
func NewGridFromRects(width, height, cellSize int, rects []Rect) *Grid {
	g := &Grid{Width: width, Height: height, CellSize: cellSize}
	g.Tiles = make([][]*Tile, height)
	for y := 0; y < height; y++ {
		row := make([]*Tile, width)
		for x := 0; x < width; x++ {
			// centre de la cellule
			cx := float64(x*cellSize + cellSize/2)
			cy := float64(y*cellSize + cellSize/2)
			walk := true
			if len(rects) > 0 {
				walk = false
				for _, r := range rects {
					if PointInRect(cx, cy, r) {
						walk = true
						break
					}
				}
			}
			row[x] = &Tile{X: x, Y: y, Walkable: walk, MovementCost: 1.0, grid: g}
		}
		g.Tiles[y] = row
	}
	return g
}

// ToWorld retourne le centre (en pixels) de la case (x,y) dans l'espace monde.
func (g *Grid) ToWorld(x, y int) Vec2 {
	cx := float64(x*g.CellSize + g.CellSize/2)
	cy := float64(y*g.CellSize + g.CellSize/2)
	return Vec2{X: cx, Y: cy}
}

// FindPath convertit des positions monde (startX,startY) et (goalX,goalY)
// en indices de grille, lance l'algorithme A* et retourne une slice de Vec2
// correspondant aux centres des cases du chemin trouvé (ou nil,false si pas de chemin).
func FindPath(startX, startY, goalX, goalY float64) ([]Vec2, bool) {
	// paramètres par défaut: monde 750x750 et taille de cellule 10 -> grille 75x75
	width := DefaultWorldSize / DefaultCellSize
	height := DefaultWorldSize / DefaultCellSize
	// si des zones autorisées ont été fournies, crée la grille à partir d'elles
	var g *Grid
	if len(AllowedRects) > 0 {
		g = NewGridFromRects(width, height, DefaultCellSize, AllowedRects)
	} else {
		g = NewGrid(width, height, DefaultCellSize)
	}

	sx := int(startX) / DefaultCellSize
	sy := int(startY) / DefaultCellSize
	gx := int(goalX) / DefaultCellSize
	gy := int(goalY) / DefaultCellSize
	if sx < 0 || sy < 0 || gx < 0 || gy < 0 || sx >= width || sy >= height || gx >= width || gy >= height {
		return nil, false
	}
	start := g.Tiles[sy][sx]
	goal := g.Tiles[gy][gx]
	path, _, found := astar.Path(start, goal)
	if !found {
		return nil, false
	}
	out := make([]Vec2, 0, len(path))
	for _, p := range path {
		if t, ok := p.(*Tile); ok {
			out = append(out, g.ToWorld(t.X, t.Y))
		}
	}
	return out, true
}

// IsWorldPointWalkable vérifie si une position monde (pixels) se trouve
// sur une case franchissable selon les rectangles autorisés (AllowedRects).
// Utile pour s'assurer que les spawns/destinations sont sur des cellules valides.
func IsWorldPointWalkable(x, y float64) bool {
	width := DefaultWorldSize / DefaultCellSize
	height := DefaultWorldSize / DefaultCellSize
	g := NewGridFromRects(width, height, DefaultCellSize, AllowedRects)
	cx, cy := WorldToCell(x, y)
	if cx < 0 || cy < 0 || cx >= g.Width || cy >= g.Height {
		return false
	}
	t := g.Tiles[cy][cx]
	if t == nil {
		return false
	}
	return t.Walkable
}
