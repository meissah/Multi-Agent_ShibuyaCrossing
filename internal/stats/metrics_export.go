package stats

// fichier qui gère l'exportation des métriques sous forme de graphiques PNG
import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"time"
)

// GenerateAndSaveCharts génère et sauvegarde tous les graphiques
func (sm *SimulationMetrics) GenerateAndSaveCharts() error {
	sm.CalculateStatistics()

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	baseFilename := fmt.Sprintf("%s_cars%.1f_peds%.1f_%s",
		sm.CrossingType,
		sm.CarSpawnRate,
		sm.PedestrianSpawnRate,
		timestamp,
	)

	// Créer le dossier metrics s'il n'existe pas
	metricsDir := "metrics_data"
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return fmt.Errorf("erreur création dossier: %v", err)
	}

	// 1. Histogramme des temps de vie des voitures
	if len(sm.CarMetrics) > 0 {
		carDurations := sm.GetLifetimeDistribution("car")
		if len(carDurations) > 0 {
			img := generateHistogram(carDurations, "Temps de vie des voitures (secondes)",
				color.RGBA{100, 200, 255, 255})
			filename := filepath.Join(metricsDir, fmt.Sprintf("histogram_cars_%s.png", baseFilename))
			if err := saveImage(img, filename); err != nil {
				return err
			}
			fmt.Printf("✓ Histogramme voitures sauvegardé: %s\n", filename)
		}
	}

	// 2. Histogramme des temps de vie des piétons
	if len(sm.PedestrianMetrics) > 0 {
		pedDurations := sm.GetLifetimeDistribution("pedestrian")
		if len(pedDurations) > 0 {
			img := generateHistogram(pedDurations, "Temps de vie des pietons (secondes)",
				color.RGBA{100, 255, 100, 255})
			filename := filepath.Join(metricsDir, fmt.Sprintf("histogram_peds_%s.png", baseFilename))
			if err := saveImage(img, filename); err != nil {
				return err
			}
			fmt.Printf("✓ Histogramme piétons sauvegardé: %s\n", filename)
		}
	}

	// 3. Graphique de comparaison
	if sm.AvgCarLifetime > 0 && sm.AvgPedestrianLife > 0 {
		img := generateComparisonChart(sm)
		filename := filepath.Join(metricsDir, fmt.Sprintf("comparison_%s.png", baseFilename))
		if err := saveImage(img, filename); err != nil {
			return err
		}
		fmt.Printf("✓ Graphique comparaison sauvegardé: %s\n", filename)
	}

	// 4. Distribution par direction
	img := generateDirectionChart(sm)
	filename := filepath.Join(metricsDir, fmt.Sprintf("directions_%s.png", baseFilename))
	if err := saveImage(img, filename); err != nil {
		return err
	}
	fmt.Printf("✓ Graphique directions sauvegardé: %s\n", filename)

	return nil
}

// generateHistogram crée un histogramme
func generateHistogram(data []float64, title string, barColor color.RGBA) *image.RGBA {
	width, height := 800, 600
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fond blanc
	drawRect(img, 0, 0, width, height, color.RGBA{255, 255, 255, 255})

	// Marges
	marginLeft, marginRight := 80, 40
	marginTop, marginBottom := 60, 80
	chartWidth := width - marginLeft - marginRight
	chartHeight := height - marginTop - marginBottom

	// Titre
	drawText(img, title, width/2-len(title)*4, 30, color.RGBA{0, 0, 0, 255})

	// Créer les bins
	numBins := 20
	if len(data) < 20 {
		numBins = len(data)
	}

	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal-minVal < 0.01 {
		maxVal = minVal + 1
	}

	// Créer les bins
	bins := make([]int, numBins)
	binWidth := (maxVal - minVal) / float64(numBins)

	for _, v := range data {
		binIdx := int((v - minVal) / binWidth)
		if binIdx >= numBins {
			binIdx = numBins - 1
		}
		bins[binIdx]++
	}

	// Trouver le max
	maxCount := 0
	for _, count := range bins {
		if count > maxCount {
			maxCount = count
		}
	}

	if maxCount == 0 {
		maxCount = 1
	}

	// Dessiner les axes
	drawLine(img, marginLeft, marginTop, marginLeft, height-marginBottom,
		color.RGBA{0, 0, 0, 255})
	drawLine(img, marginLeft, height-marginBottom, width-marginRight, height-marginBottom,
		color.RGBA{0, 0, 0, 255})

	// Dessiner les barres
	barW := float64(chartWidth) / float64(numBins)
	for i, count := range bins {
		barH := (float64(count) / float64(maxCount)) * float64(chartHeight)
		x := marginLeft + int(float64(i)*barW)
		y := height - marginBottom - int(barH)
		drawRect(img, x, y, int(barW)-2, int(barH), barColor)
	}

	// Labels axes
	drawText(img, fmt.Sprintf("%.1fs", minVal), marginLeft-20, height-marginBottom+20,
		color.RGBA{0, 0, 0, 255})
	drawText(img, fmt.Sprintf("%.1fs", maxVal), width-marginRight-30, height-marginBottom+20,
		color.RGBA{0, 0, 0, 255})
	drawText(img, fmt.Sprintf("%d", maxCount), marginLeft-40, marginTop,
		color.RGBA{0, 0, 0, 255})

	return img
}

// generateComparisonChart crée un graphique de comparaison
func generateComparisonChart(sm *SimulationMetrics) *image.RGBA {
	width, height := 800, 600
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fond blanc
	drawRect(img, 0, 0, width, height, color.RGBA{255, 255, 255, 255})

	// Titre
	title := "Comparaison Voitures vs Pietons"
	drawText(img, title, width/2-len(title)*4, 30, color.RGBA{0, 0, 0, 255})

	// Barres de comparaison
	barWidth := 120
	centerY := height / 2
	spacing := 200

	// Voitures (gauche)
	carX := width/2 - spacing
	maxHeight := 300.0
	carBarHeight := int((sm.AvgCarLifetime / math.Max(sm.AvgCarLifetime, sm.AvgPedestrianLife)) * maxHeight)
	drawRect(img, carX-barWidth/2, centerY-carBarHeight, barWidth, carBarHeight,
		color.RGBA{100, 200, 255, 255})
	drawText(img, "Voitures", carX-30, centerY+30, color.RGBA{0, 0, 0, 255})
	drawText(img, fmt.Sprintf("%.2fs", sm.AvgCarLifetime), carX-25, centerY+50,
		color.RGBA{0, 0, 0, 255})

	// Piétons (droite)
	pedX := width/2 + spacing
	pedBarHeight := int((sm.AvgPedestrianLife / math.Max(sm.AvgCarLifetime, sm.AvgPedestrianLife)) * maxHeight)
	drawRect(img, pedX-barWidth/2, centerY-pedBarHeight, barWidth, pedBarHeight,
		color.RGBA{100, 255, 100, 255})
	drawText(img, "Pietons", pedX-25, centerY+30, color.RGBA{0, 0, 0, 255})
	drawText(img, fmt.Sprintf("%.2fs", sm.AvgPedestrianLife), pedX-25, centerY+50,
		color.RGBA{0, 0, 0, 255})

	return img
}

// generateDirectionChart crée un graphique des directions
func generateDirectionChart(sm *SimulationMetrics) *image.RGBA {
	width, height := 800, 600
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fond blanc
	drawRect(img, 0, 0, width, height, color.RGBA{255, 255, 255, 255})

	// Titre
	title := "Distribution par direction"
	drawText(img, title, width/2-len(title)*4, 30, color.RGBA{0, 0, 0, 255})

	// Créer des barres pour chaque direction
	directions := []string{"North", "South", "East", "West"}
	colors := []color.RGBA{
		{255, 100, 100, 255}, // North - Rouge
		{100, 100, 255, 255}, // South - Bleu
		{100, 255, 100, 255}, // East - Vert
		{255, 255, 100, 255}, // West - Jaune
	}

	barWidth := 80
	startX := 150
	spacing := 150
	maxBarHeight := 400

	// Trouver le max pour normaliser
	maxCount := 0
	for _, dir := range directions {
		count := sm.CarsByDirection[dir] + sm.PedestriansByDirection[dir]
		if count > maxCount {
			maxCount = count
		}
	}

	if maxCount == 0 {
		maxCount = 1
	}

	// Dessiner les barres
	for i, dir := range directions {
		count := sm.CarsByDirection[dir] + sm.PedestriansByDirection[dir]
		barHeight := int((float64(count) / float64(maxCount)) * float64(maxBarHeight))
		x := startX + i*spacing
		y := height - 100 - barHeight

		drawRect(img, x, y, barWidth, barHeight, colors[i])
		drawText(img, dir, x+10, height-70, color.RGBA{0, 0, 0, 255})
		drawText(img, fmt.Sprintf("%d", count), x+20, height-50, color.RGBA{0, 0, 0, 255})
	}

	return img
}

// Fonctions helper pour dessiner

func drawRect(img *image.RGBA, x, y, w, h int, col color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			if i >= 0 && i < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
				img.Set(i, j, col)
			}
		}
	}
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x1, y1, col)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func drawText(img *image.RGBA, text string, x, y int, col color.RGBA) {
	// Dessin de texte simple (8x8 pixels par caractère)
	for i, char := range text {
		drawChar(img, char, x+i*8, y, col)
	}
}

func drawChar(img *image.RGBA, char rune, x, y int, col color.RGBA) {
	// Dessiner un rectangle simple pour chaque caractère
	drawRect(img, x, y, 6, 8, col)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func saveImage(img *image.RGBA, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}
