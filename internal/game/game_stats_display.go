// Ce fichier gère l'affichage des statistiques
// Il gère également l'affichage du panneau de métriques en temps réel ( a droite de la simu ).
package game

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/meissah/Multi-Agent_ShibuyaCrossing/internal/stats"
	"golang.org/x/image/font/basicfont"
)

// drawWrappedText dessine du texte avec retour à la ligne automatique
func (g *Game) drawWrappedText(screen *ebiten.Image, content string, x, y float64, maxWidth int, col color.Color) {
	currentY := int(y)
	lineHeight := 18

	lines := splitIntoLines(content, maxWidth/7) // ~7 pixels par caractère

	for _, line := range lines {
		text.Draw(screen, line, basicfont.Face7x13, int(x), currentY, col)
		currentY += lineHeight
	}
}

// splitIntoLines découpe un texte en lignes selon la largeur max
func splitIntoLines(text string, maxChars int) []string {
	var lines []string
	currentLine := ""

	words := strings.Fields(text)

	for _, word := range words {
		// Si le mot contient un saut de ligne
		if strings.Contains(word, "\n") {
			parts := strings.Split(word, "\n")
			for i, part := range parts {
				if i > 0 {
					lines = append(lines, currentLine)
					currentLine = ""
				}
				if len(currentLine)+len(part)+1 <= maxChars {
					if currentLine != "" {
						currentLine += " "
					}
					currentLine += part
				} else {
					if currentLine != "" {
						lines = append(lines, currentLine)
					}
					currentLine = part
				}
			}
		} else {
			// Mot normal
			if len(currentLine)+len(word)+1 <= maxChars {
				if currentLine != "" {
					currentLine += " "
				}
				currentLine += word
			} else {
				if currentLine != "" {
					lines = append(lines, currentLine)
				}
				currentLine = word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// ----------------------------------------------------------------
// Affichage du panneau de métriques a droite de la simulation
// ----------------------------------------------------------------

// DrawMetricsPanel dessine le panneau de métriques à droite de la simulation
func (g *Game) DrawMetricsPanel(screen *ebiten.Image) {
	if stats.GlobalMetrics == nil {
		return
	}

	// Position du panneau (à droite de la simulation 750x750)
	panelX := 750.0
	panelY := 0.0
	panelWidth := 400.0
	panelHeight := 750.0

	// Fond du panneau avec transparence
	panel := ebiten.NewImage(int(panelWidth), int(panelHeight))
	panel.Fill(color.RGBA{20, 25, 35, 240})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(panelX, panelY)
	screen.DrawImage(panel, op)

	// Bordure du panneau
	g.drawPanelBorder(screen, panelX, panelY, panelWidth, panelHeight)

	// Titre du panneau
	title := "METRIQUES EN TEMPS REEL"
	titleX := panelX + (panelWidth-float64(len(title)*7))/2
	text.Draw(screen, title, basicfont.Face7x13,
		int(titleX), int(panelY+25), color.RGBA{100, 200, 255, 255})

	// Ligne de séparation
	separator := ebiten.NewImage(int(panelWidth-20), 2)
	separator.Fill(color.RGBA{50, 140, 200, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(panelX+10, panelY+35)
	screen.DrawImage(separator, op)

	// Récupérer les stats actuelles
	currentStats := stats.GlobalMetrics.GetCurrentStats()

	// Position Y courante pour l'affichage
	currentY := panelY + 65

	// Section: Informations générales
	currentY = g.drawMetricSection(screen, panelX+15, currentY, "INFORMATIONS GENERALES")

	simDuration := currentStats["simulation_duration"].(float64)
	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Duree simulation:", fmt.Sprintf("%.1f s", simDuration), color.RGBA{200, 200, 200, 255})

	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Type carrefour:", g.Config.CrossingType, color.RGBA{200, 200, 200, 255})

	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Freq. voitures:", fmt.Sprintf("%.1f/s", 1.0/g.Config.CarSpawnRate.Seconds()),
		color.RGBA{200, 200, 200, 255})

	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Freq. pietons:", fmt.Sprintf("%.1f/s", 1.0/g.Config.PedestrianSpawnRate.Seconds()),
		color.RGBA{200, 200, 200, 255})

	currentY += 15

	// Section: Voitures
	currentY = g.drawMetricSection(screen, panelX+15, currentY, "VOITURES")

	totalCars := currentStats["total_cars"].(int)
	completedCars := currentStats["completed_cars"].(int)
	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Total spawned:", fmt.Sprintf("%d", totalCars), color.RGBA{100, 200, 255, 255})

	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Detruites:", fmt.Sprintf("%d", completedCars), color.RGBA{255, 150, 50, 255})

	activeCars := totalCars - completedCars
	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Actives:", fmt.Sprintf("%d", activeCars), color.RGBA{50, 255, 100, 255})

	avgCarLife := currentStats["avg_car_lifetime"].(float64)
	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Temps moy. vie:", fmt.Sprintf("%.2f s", avgCarLife), color.RGBA{255, 255, 100, 255})

	// Vitesse moyenne des voitures
	avgCarSpeed := currentStats["avg_car_speed"].(float64)
	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Vitesse moyenne:", fmt.Sprintf("%.2f px/s", avgCarSpeed),
		color.RGBA{255, 200, 100, 255})

	currentY += 15 // Espacement avant la section piétons

	// Section: Piétons
	currentY = g.drawMetricSection(screen, panelX+15, currentY, "PIETONS")

	totalPeds := currentStats["total_pedestrians"].(int)
	completedPeds := currentStats["completed_pedestrians"].(int)
	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Total spawned:", fmt.Sprintf("%d", totalPeds), color.RGBA{100, 200, 255, 255})

	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Detruits:", fmt.Sprintf("%d", completedPeds), color.RGBA{255, 150, 50, 255})

	activePeds := totalPeds - completedPeds
	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Actifs:", fmt.Sprintf("%d", activePeds), color.RGBA{50, 255, 100, 255})

	avgPedLife := currentStats["avg_ped_lifetime"].(float64)
	currentY = g.drawMetricLine(screen, panelX+20, currentY,
		"Temps moy. vie:", fmt.Sprintf("%.2f s", avgPedLife), color.RGBA{255, 255, 100, 255})

	// Bouton de sauvegarde en bas
	g.drawSaveButton(screen, panelX+10, panelY+panelHeight-50, panelWidth-20, 40)
}

// drawMetricSection dessine un titre de section
func (g *Game) drawMetricSection(screen *ebiten.Image, x, y float64, title string) float64 {
	text.Draw(screen, title, basicfont.Face7x13,
		int(x), int(y), color.RGBA{255, 200, 100, 255})

	// Ligne sous le titre
	line := ebiten.NewImage(340, 1)
	line.Fill(color.RGBA{255, 200, 100, 100})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y+5)
	screen.DrawImage(line, op)

	return y + 20
}

// drawMetricLine dessine une ligne de métrique
func (g *Game) drawMetricLine(screen *ebiten.Image, x, y float64, label, value string, col color.Color) float64 {
	// Label à gauche
	text.Draw(screen, label, basicfont.Face7x13,
		int(x), int(y), color.RGBA{180, 180, 180, 255})

	// Valeur à droite (alignée)
	valueX := x + 200
	text.Draw(screen, value, basicfont.Face7x13,
		int(valueX), int(y), col)

	return y + 18
}

// drawPanelBorder dessine la bordure du panneau
func (g *Game) drawPanelBorder(screen *ebiten.Image, x, y, w, h float64) {
	borderColor := color.RGBA{50, 140, 200, 255}
	thickness := 2.0

	// Bordures horizontales
	hBorder := ebiten.NewImage(int(w), int(thickness))
	hBorder.Fill(borderColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(hBorder, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y+h-thickness)
	screen.DrawImage(hBorder, op)

	// Bordures verticales
	vBorder := ebiten.NewImage(int(thickness), int(h))
	vBorder.Fill(borderColor)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(vBorder, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x+w-thickness, y)
	screen.DrawImage(vBorder, op)
}

// drawSaveButton dessine le bouton de sauvegarde
func (g *Game) drawSaveButton(screen *ebiten.Image, x, y, w, h float64) {
	// Vérifier si la souris est sur le bouton
	mx, my := ebiten.CursorPosition()
	mouseX, mouseY := float64(mx), float64(my)
	isHover := mouseX >= x && mouseX <= x+w && mouseY >= y && mouseY <= y+h

	// Couleur selon hover
	bgColor := color.RGBA{0, 150, 80, 255}
	if isHover {
		bgColor = color.RGBA{0, 200, 100, 255}
	}

	// Fond du bouton
	btn := ebiten.NewImage(int(w), int(h))
	btn.Fill(bgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(btn, op)

	// Bordure
	borderColor := color.RGBA{0, 255, 150, 255}
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h),
		2, borderColor, false)

	// Texte
	btnText := "SAUVEGARDER LES METRIQUES"
	textX := x + (w-float64(len(btnText)*7))/2
	text.Draw(screen, btnText, basicfont.Face7x13,
		int(textX), int(y+h/2+5), color.RGBA{255, 255, 255, 255})

	// Gérer le clic
	if isHover && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// Sauvegarder les métriques
		if err := stats.GlobalMetrics.SaveToFile(); err != nil {
			fmt.Printf("Erreur sauvegarde métriques: %v\n", err)
		}
	}
}
