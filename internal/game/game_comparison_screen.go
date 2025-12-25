// Ce fichier gère l'affichage amélioré de l'écran de comparaison
// avec graphiques, histogrammes et analyses statistiques détaillées
package game

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/stats"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	"golang.org/x/image/font/basicfont"
)

// DrawComparisonScreen affiche l'écran de comparaison des résultats avec graphiques
func (g *Game) DrawComparisonScreen(screen *ebiten.Image) {
	// Fond sombre
	screen.Fill(color.RGBA{20, 25, 35, 255})

	// Vérifier si on a des données
	if g.comparisonResult == nil {
		g.drawErrorMessage(screen, "Aucune donnée disponible")
		return
	}

	result := g.comparisonResult

	// ═══════════════════════════════════════════════════════════
	// EN-TÊTE
	// ═══════════════════════════════════════════════════════════

	// Titre principal
	titleText := "COMPARAISON"
	titleX := (1400 - len(titleText)*7) / 2
	text.Draw(screen, titleText, basicfont.Face7x13,
		titleX, 25, color.RGBA{0, 220, 255, 255})

	// Sous-titre
	subtitle := "Shibuya vs French"
	subtitleX := (1400 - len(subtitle)*7) / 2
	text.Draw(screen, subtitle, basicfont.Face7x13,
		subtitleX, 40, color.RGBA{150, 160, 180, 255})

	// Ligne de séparation
	separator := ebiten.NewImage(1320, 1)
	separator.Fill(color.RGBA{0, 220, 255, 80})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(40, 48)
	screen.DrawImage(separator, op)

	// ═══════════════════════════════════════════════════════════
	// BOUTON REFRESH (en haut à droite)
	// ═══════════════════════════════════════════════════════════

	g.drawRefreshButton(screen)

	currentY := 60.0

	// ═══════════════════════════════════════════════════════════
	// STATISTIQUES RESUMÉES
	// ═══════════════════════════════════════════════════════════

	if !result.ShibuyaAvailable && !result.FrenchAvailable {
		g.drawErrorMessage(screen, "Aucune donnée disponible.\nLancez au moins une simulation de chaque type.")
		return
	}

	// Cartes de statistiques (Shibuya vs French)
	g.drawStatsCards(screen, result, currentY)
	currentY += 115

	// ═══════════════════════════════════════════════════════════
	// ANALYSE COMPARATIVE (si on a les 2 types)
	// ═══════════════════════════════════════════════════════════

	if result.ShibuyaAvailable && result.FrenchAvailable {
		g.drawComparativeAnalysis(screen, result, currentY)
		currentY += 65
	}

	// ═══════════════════════════════════════════════════════════
	// GRAPHIQUES
	// ═══════════════════════════════════════════════════════════

	// Histogramme de comparaison des temps moyens
	if result.ShibuyaAvailable && result.FrenchAvailable {
		g.drawComparisonHistogram(screen, result, 40, currentY, 640, 190)
		g.drawTotalAgentsChart(screen, result, 720, currentY, 640, 190)
		currentY += 205
	}

	// ═══════════════════════════════════════════════════════════
	// CONCLUSION
	// ═══════════════════════════════════════════════════════════

	if result.ShibuyaAvailable && result.FrenchAvailable {
		g.drawConclusion(screen, result, currentY)
	}

	// ═══════════════════════════════════════════════════════════
	// INSTRUCTIONS
	// ═══════════════════════════════════════════════════════════

	instruction := "Cliquez n'importe ou pour fermer"
	instX := (1400 - len(instruction)*7) / 2
	text.Draw(screen, instruction, basicfont.Face7x13,
		instX, 880, color.RGBA{120, 130, 150, 255})
}

// drawStatsCards dessine les cartes de statistiques pour Shibuya et French
func (g *Game) drawStatsCards(screen *ebiten.Image, result *stats.ComparisonResult, startY float64) {
	cardWidth := 640.0
	cardHeight := 100.0
	spacing := 40.0
	marginLeft := 40.0

	// Carte Shibuya (gauche)
	if result.ShibuyaAvailable {
		g.drawStatCard(screen, "SHIBUYA", result.ShibuyaStats,
			marginLeft, startY, cardWidth, cardHeight, color.RGBA{0, 200, 255, 255})
	}

	// Carte French (droite)
	if result.FrenchAvailable {
		g.drawStatCard(screen, "FRENCH", result.FrenchStats,
			marginLeft+cardWidth+spacing, startY, cardWidth, cardHeight, color.RGBA{160, 100, 255, 255})
	}
}

// drawStatCard dessine une carte de statistiques individuelle
func (g *Game) drawStatCard(screen *ebiten.Image, title string, stats stats.AnalysisStats, x, y, w, h float64, accentColor color.RGBA) {
	// Fond de la carte
	cardBg := ebiten.NewImage(int(w), int(h))
	cardBg.Fill(color.RGBA{35, 40, 50, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(cardBg, op)

	// Bordure
	g.drawRectBorder(screen, ui.Rectangle{X: x, Y: y, W: w, H: h}, 2, accentColor)

	// Titre
	text.Draw(screen, title, basicfont.Face7x13,
		int(x+10), int(y+20), accentColor)

	// Nombre de simulations
	simText := fmt.Sprintf("(%d simulations)", stats.SimulationCount)
	text.Draw(screen, simText, basicfont.Face7x13,
		int(x+10), int(y+40), color.RGBA{150, 160, 170, 255})

	// Statistiques en colonnes - layout horizontal compact
	col1X := x + 12
	col2X := x + w/2 + 5
	statsY := y + 42

	// Colonne 1: Voitures
	text.Draw(screen, "Voitures:", basicfont.Face7x13,
		int(col1X), int(statsY), color.RGBA{180, 180, 180, 255})
	text.Draw(screen, fmt.Sprintf("%.1fs moy", stats.AvgCarLifetime), basicfont.Face7x13,
		int(col1X), int(statsY+15), color.RGBA{0, 220, 255, 255})
	text.Draw(screen, fmt.Sprintf("%d total", stats.TotalCars), basicfont.Face7x13,
		int(col1X), int(statsY+28), color.RGBA{120, 200, 255, 255})

	// Colonne 2: Piétons
	text.Draw(screen, "Pietons:", basicfont.Face7x13,
		int(col2X), int(statsY), color.RGBA{180, 180, 180, 255})
	text.Draw(screen, fmt.Sprintf("%.1fs moy", stats.AvgPedLifetime), basicfont.Face7x13,
		int(col2X), int(statsY+15), color.RGBA{100, 255, 100, 255})
	text.Draw(screen, fmt.Sprintf("%d total", stats.TotalPeds), basicfont.Face7x13,
		int(col2X), int(statsY+28), color.RGBA{120, 255, 150, 255})
}

// drawComparativeAnalysis dessine l'analyse comparative
func (g *Game) drawComparativeAnalysis(screen *ebiten.Image, result *stats.ComparisonResult, startY float64) {
	// Panneau d'analyse
	panelW, panelH := 1320.0, 55.0
	panelX, panelY := 40.0, startY

	panelBg := ebiten.NewImage(int(panelW), int(panelH))
	panelBg.Fill(color.RGBA{40, 50, 60, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(panelX, panelY)
	screen.DrawImage(panelBg, op)

	g.drawRectBorder(screen, ui.Rectangle{X: panelX, Y: panelY, W: panelW, H: panelH},
		2, color.RGBA{255, 200, 0, 255})

	// Titre
	text.Draw(screen, "ANALYSE COMPARATIVE", basicfont.Face7x13,
		int(panelX+10), int(panelY+20), color.RGBA{255, 200, 0, 255})

	// Différences - layout horizontal
	diffY := panelY + 35
	col1X := panelX + 30
	col2X := panelX + 350
	col3X := panelX + 670
	col4X := panelX + 1020

	// Voitures Shibuya
	text.Draw(screen, "Voitures:", basicfont.Face7x13,
		int(col1X), int(diffY), color.RGBA{180, 180, 180, 255})
	text.Draw(screen, fmt.Sprintf("%.1fs", result.ShibuyaStats.AvgCarLifetime), basicfont.Face7x13,
		int(col1X), int(diffY+15), color.RGBA{0, 220, 255, 255})

	// Diff voitures
	carDiffColor := color.RGBA{100, 255, 100, 255}
	if result.CarDiffPercent > 0 {
		carDiffColor = color.RGBA{255, 100, 100, 255}
	}
	text.Draw(screen, fmt.Sprintf("%+.1f%%", result.CarDiffPercent), basicfont.Face7x13,
		int(col2X), int(diffY+7), carDiffColor)

	// Piétons Shibuya
	text.Draw(screen, "Pietons:", basicfont.Face7x13,
		int(col3X), int(diffY), color.RGBA{180, 180, 180, 255})
	text.Draw(screen, fmt.Sprintf("%.1fs", result.ShibuyaStats.AvgPedLifetime), basicfont.Face7x13,
		int(col3X), int(diffY+15), color.RGBA{100, 255, 100, 255})

	// Diff piétons
	pedDiffColor := color.RGBA{100, 255, 100, 255}
	if result.PedDiffPercent > 0 {
		pedDiffColor = color.RGBA{255, 100, 100, 255}
	}
	text.Draw(screen, fmt.Sprintf("%+.1f%%", result.PedDiffPercent), basicfont.Face7x13,
		int(col4X), int(diffY+7), pedDiffColor)
}

// drawComparisonHistogram dessine un histogramme de comparaison
func (g *Game) drawComparisonHistogram(screen *ebiten.Image, result *stats.ComparisonResult, x, y, w, h float64) {
	// Fond du graphique
	chartBg := ebiten.NewImage(int(w), int(h))
	chartBg.Fill(color.RGBA{35, 40, 50, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(chartBg, op)

	g.drawRectBorder(screen, ui.Rectangle{X: x, Y: y, W: w, H: h}, 2, color.RGBA{100, 200, 255, 255})

	// Titre
	text.Draw(screen, "TEMPS MOYENS DE VIE", basicfont.Face7x13,
		int(x+w/2-70), int(y+20), color.RGBA{100, 200, 255, 255})

	// Marges pour le graphique
	marginLeft := 60.0
	chartAreaY := y + 40
	chartAreaH := h - 60

	// Trouver la valeur maximale
	maxVal := 0.0
	if result.ShibuyaStats.AvgCarLifetime > maxVal {
		maxVal = result.ShibuyaStats.AvgCarLifetime
	}
	if result.FrenchStats.AvgCarLifetime > maxVal {
		maxVal = result.FrenchStats.AvgCarLifetime
	}
	if result.ShibuyaStats.AvgPedLifetime > maxVal {
		maxVal = result.ShibuyaStats.AvgPedLifetime
	}
	if result.FrenchStats.AvgPedLifetime > maxVal {
		maxVal = result.FrenchStats.AvgPedLifetime
	}

	if maxVal == 0 {
		maxVal = 1
	}

	// Dessiner les barres
	barWidth := 60.0

	// Groupe 1: Voitures
	group1X := x + marginLeft + 50

	// Shibuya Cars
	shibuyaCarH := (result.ShibuyaStats.AvgCarLifetime / maxVal) * chartAreaH
	g.drawBar(screen, group1X, chartAreaY+chartAreaH-shibuyaCarH, barWidth, shibuyaCarH,
		color.RGBA{0, 200, 255, 255})
	text.Draw(screen, fmt.Sprintf("%.1f", result.ShibuyaStats.AvgCarLifetime), basicfont.Face7x13,
		int(group1X+10), int(chartAreaY+chartAreaH-shibuyaCarH-5), color.RGBA{0, 200, 255, 255})

	// French Cars
	frenchCarH := (result.FrenchStats.AvgCarLifetime / maxVal) * chartAreaH
	g.drawBar(screen, group1X+barWidth+10, chartAreaY+chartAreaH-frenchCarH, barWidth, frenchCarH,
		color.RGBA{160, 100, 255, 255})
	text.Draw(screen, fmt.Sprintf("%.1f", result.FrenchStats.AvgCarLifetime), basicfont.Face7x13,
		int(group1X+barWidth+20), int(chartAreaY+chartAreaH-frenchCarH-5), color.RGBA{160, 100, 255, 255})

	// Label Voitures
	text.Draw(screen, "Voitures", basicfont.Face7x13,
		int(group1X+20), int(y+h-15), color.RGBA{180, 180, 180, 255})

	// Groupe 2: Piétons
	group2X := x + marginLeft + 250

	// Shibuya Peds
	shibuyaPedH := (result.ShibuyaStats.AvgPedLifetime / maxVal) * chartAreaH
	g.drawBar(screen, group2X, chartAreaY+chartAreaH-shibuyaPedH, barWidth, shibuyaPedH,
		color.RGBA{100, 255, 100, 255})
	text.Draw(screen, fmt.Sprintf("%.1f", result.ShibuyaStats.AvgPedLifetime), basicfont.Face7x13,
		int(group2X+10), int(chartAreaY+chartAreaH-shibuyaPedH-5), color.RGBA{100, 255, 100, 255})

	// French Peds
	frenchPedH := (result.FrenchStats.AvgPedLifetime / maxVal) * chartAreaH
	g.drawBar(screen, group2X+barWidth+10, chartAreaY+chartAreaH-frenchPedH, barWidth, frenchPedH,
		color.RGBA{255, 150, 50, 255})
	text.Draw(screen, fmt.Sprintf("%.1f", result.FrenchStats.AvgPedLifetime), basicfont.Face7x13,
		int(group2X+barWidth+20), int(chartAreaY+chartAreaH-frenchPedH-5), color.RGBA{255, 150, 50, 255})

	// Label Piétons
	text.Draw(screen, "Pietons", basicfont.Face7x13,
		int(group2X+25), int(y+h-15), color.RGBA{180, 180, 180, 255})

	// Légende
	legendY := y + h - 30
	legendX := x + w - 180
	text.Draw(screen, "Shibuya", basicfont.Face7x13,
		int(legendX+20), int(legendY), color.RGBA{0, 200, 255, 255})
	g.drawBar(screen, legendX, legendY-10, 15, 15, color.RGBA{0, 200, 255, 255})

	text.Draw(screen, "French", basicfont.Face7x13,
		int(legendX+20), int(legendY+20), color.RGBA{160, 100, 255, 255})
	g.drawBar(screen, legendX, legendY+10, 15, 15, color.RGBA{160, 100, 255, 255})
}

// drawTotalAgentsChart dessine un graphique du total d'agents
func (g *Game) drawTotalAgentsChart(screen *ebiten.Image, result *stats.ComparisonResult, x, y, w, h float64) {
	// Fond du graphique
	chartBg := ebiten.NewImage(int(w), int(h))
	chartBg.Fill(color.RGBA{35, 40, 50, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(chartBg, op)

	g.drawRectBorder(screen, ui.Rectangle{X: x, Y: y, W: w, H: h}, 2, color.RGBA{100, 255, 100, 255})

	// Titre
	text.Draw(screen, "NOMBRE TOTAL D'AGENTS", basicfont.Face7x13,
		int(x+w/2-80), int(y+20), color.RGBA{100, 255, 100, 255})

	chartAreaY := y + 40
	chartAreaH := h - 60
	marginLeft := 60.0

	// Trouver le max
	maxVal := 0
	if result.ShibuyaStats.TotalCars > maxVal {
		maxVal = result.ShibuyaStats.TotalCars
	}
	if result.FrenchStats.TotalCars > maxVal {
		maxVal = result.FrenchStats.TotalCars
	}
	if result.ShibuyaStats.TotalPeds > maxVal {
		maxVal = result.ShibuyaStats.TotalPeds
	}
	if result.FrenchStats.TotalPeds > maxVal {
		maxVal = result.FrenchStats.TotalPeds
	}

	if maxVal == 0 {
		maxVal = 1
	}

	barWidth := 60.0
	group1X := x + marginLeft + 50
	group2X := x + marginLeft + 250

	// Voitures
	shibuyaCarH := (float64(result.ShibuyaStats.TotalCars) / float64(maxVal)) * chartAreaH
	g.drawBar(screen, group1X, chartAreaY+chartAreaH-shibuyaCarH, barWidth, shibuyaCarH,
		color.RGBA{0, 200, 255, 255})
	text.Draw(screen, fmt.Sprintf("%d", result.ShibuyaStats.TotalCars), basicfont.Face7x13,
		int(group1X+15), int(chartAreaY+chartAreaH-shibuyaCarH-5), color.RGBA{0, 200, 255, 255})

	frenchCarH := (float64(result.FrenchStats.TotalCars) / float64(maxVal)) * chartAreaH
	g.drawBar(screen, group1X+barWidth+10, chartAreaY+chartAreaH-frenchCarH, barWidth, frenchCarH,
		color.RGBA{160, 100, 255, 255})
	text.Draw(screen, fmt.Sprintf("%d", result.FrenchStats.TotalCars), basicfont.Face7x13,
		int(group1X+barWidth+20), int(chartAreaY+chartAreaH-frenchCarH-5), color.RGBA{160, 100, 255, 255})

	text.Draw(screen, "Voitures", basicfont.Face7x13,
		int(group1X+20), int(y+h-15), color.RGBA{180, 180, 180, 255})

	// Piétons
	shibuyaPedH := (float64(result.ShibuyaStats.TotalPeds) / float64(maxVal)) * chartAreaH
	g.drawBar(screen, group2X, chartAreaY+chartAreaH-shibuyaPedH, barWidth, shibuyaPedH,
		color.RGBA{100, 255, 100, 255})
	text.Draw(screen, fmt.Sprintf("%d", result.ShibuyaStats.TotalPeds), basicfont.Face7x13,
		int(group2X+15), int(chartAreaY+chartAreaH-shibuyaPedH-5), color.RGBA{100, 255, 100, 255})

	frenchPedH := (float64(result.FrenchStats.TotalPeds) / float64(maxVal)) * chartAreaH
	g.drawBar(screen, group2X+barWidth+10, chartAreaY+chartAreaH-frenchPedH, barWidth, frenchPedH,
		color.RGBA{255, 150, 50, 255})
	text.Draw(screen, fmt.Sprintf("%d", result.FrenchStats.TotalPeds), basicfont.Face7x13,
		int(group2X+barWidth+20), int(chartAreaY+chartAreaH-frenchPedH-5), color.RGBA{255, 150, 50, 255})

	text.Draw(screen, "Pietons", basicfont.Face7x13,
		int(group2X+25), int(y+h-15), color.RGBA{180, 180, 180, 255})
}

// drawBar dessine une barre simple
func (g *Game) drawBar(screen *ebiten.Image, x, y, w, h float64, col color.RGBA) {
	if h <= 0 {
		return
	}
	bar := ebiten.NewImage(int(w), int(h))
	bar.Fill(col)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(bar, op)
}

// drawConclusion dessine la conclusion
func (g *Game) drawConclusion(screen *ebiten.Image, result *stats.ComparisonResult, startY float64) {
	panelW, panelH := 1320.0, 48.0
	panelX, panelY := 40.0, startY

	panelBg := ebiten.NewImage(int(panelW), int(panelH))
	panelBg.Fill(color.RGBA{20, 80, 100, 100})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(panelX, panelY)
	screen.DrawImage(panelBg, op)

	g.drawRectBorder(screen, ui.Rectangle{X: panelX, Y: panelY, W: panelW, H: panelH},
		2, color.RGBA{0, 200, 255, 255})

	// Titre
	text.Draw(screen, "CONCLUSION", basicfont.Face7x13,
		int(panelX+10), int(panelY+20), color.RGBA{0, 220, 255, 255})

	// Texte de conclusion - centré et wrap si nécessaire
	conclusionColor := color.RGBA{220, 220, 220, 255}
	conclusionText := result.Conclusion
	if len(conclusionText) > 150 {
		conclusionText = conclusionText[:147] + "..."
	}
	text.Draw(screen, conclusionText, basicfont.Face7x13,
		int(panelX+15), int(panelY+30), conclusionColor)
}

// drawRefreshButton dessine le bouton de rafraîchissement
func (g *Game) drawRefreshButton(screen *ebiten.Image) {
	btn := g.refreshButton

	// Couleurs
	var bgColor, textColor color.Color
	if btn.IsMouseOver {
		bgColor = color.RGBA{220, 60, 60, 255}
		textColor = color.RGBA{255, 255, 255, 255}
	} else {
		bgColor = color.RGBA{180, 50, 50, 255}
		textColor = color.RGBA{255, 255, 255, 255}
	}

	// Fond
	btnImg := ebiten.NewImage(int(btn.Rect.W), int(btn.Rect.H))
	btnImg.Fill(bgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(btn.Rect.X, btn.Rect.Y)
	screen.DrawImage(btnImg, op)

	// Bordure
	borderColor := color.RGBA{255, 100, 100, 255}
	g.drawRectBorder(screen, btn.Rect, 2, borderColor)

	// Texte
	textX := int(btn.Rect.X + btn.Rect.W/2 - float64(len(btn.Text)*7)/2)
	textY := int(btn.Rect.Y + btn.Rect.H/2 + 5)
	text.Draw(screen, btn.Text, basicfont.Face7x13, textX, textY, textColor)
}

// drawErrorMessage dessine un message d'erreur centré
func (g *Game) drawErrorMessage(screen *ebiten.Image, message string) {
	lines := strings.Split(message, "\n")
	startY := 400

	for i, line := range lines {
		textX := (1400 - len(line)*7) / 2
		text.Draw(screen, line, basicfont.Face7x13,
			textX, startY+i*20, color.RGBA{255, 100, 100, 255})
	}

	// Message pour retourner
	instruction := "Cliquez n'importe ou pour fermer"
	instX := (1400 - len(instruction)*7) / 2
	text.Draw(screen, instruction, basicfont.Face7x13,
		instX, 860, color.RGBA{120, 130, 150, 255})
}

// HandleComparisonClick gère les clics dans l'écran de comparaison
func (g *Game) HandleComparisonClick(mouseX, mouseY float64) {
	// Vérifier le clic sur le bouton refresh
	if g.refreshButton.IsMouseOver && !g.refreshButtonPress {
		g.refreshButtonPress = true
		// Vider le dossier metrics
		if err := stats.ClearMetricsFolder(); err != nil {
			fmt.Printf("Erreur lors de la suppression des métriques: %v\n", err)
		} else {
			fmt.Println("✓ Dossier de métriques vidé avec succès")
			// Recharger la comparaison (qui sera vide maintenant)
			g.comparisonResult = &stats.ComparisonResult{}
		}
	} else if !g.refreshButton.IsMouseOver {
		g.refreshButtonPress = false
	}

	// Clic ailleurs = fermer
	if !g.refreshButton.Rect.Contains(mouseX, mouseY) {
		// Attendre un petit délai pour éviter de fermer immédiatement après avoir ouvert
		g.showComparison = false
	}
}

// UpdateComparisonHover met à jour l'état hover du bouton refresh
func (g *Game) UpdateComparisonHover(mouseX, mouseY float64) {
	g.refreshButton.UpdateHover(mouseX, mouseY)
}
