// Ce fichier gère l'affichage de la comparaison des statistiques
// entre les différentes simulations (Shibuya vs French crossing).
package game

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	"golang.org/x/image/font/basicfont"
)

// drawCompareButton dessine le bouton de comparaison dans le menu principal
func (g *Game) drawCompareButton(screen *ebiten.Image, btn ui.Button) {
	var bgColor, textColor color.Color

	if btn.IsMouseOver {
		// Hover - orange vif
		bgColor = color.RGBA{255, 140, 0, 255}
		textColor = color.RGBA{255, 255, 255, 255}

		// Effet de lueur
		glowBtn := ebiten.NewImage(int(btn.Rect.W+10), int(btn.Rect.H+10))
		glowBtn.Fill(color.RGBA{255, 140, 0, 60})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(btn.Rect.X-5, btn.Rect.Y-5)
		screen.DrawImage(glowBtn, op)
	} else {
		// Normal - orange foncé
		bgColor = color.RGBA{255, 100, 0, 255}
		textColor = color.RGBA{255, 255, 255, 255}
	}

	// Fond du bouton
	btnImg := ebiten.NewImage(int(btn.Rect.W), int(btn.Rect.H))
	btnImg.Fill(bgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(btn.Rect.X, btn.Rect.Y)
	screen.DrawImage(btnImg, op)

	// Bordure
	borderColor := color.RGBA{255, 165, 0, 255}
	g.drawRectBorder(screen, btn.Rect, 3, borderColor)

	// Texte centré
	textX := int(btn.Rect.X + btn.Rect.W/2 - float64(len(btn.Text)*7)/2)
	textY := int(btn.Rect.Y + btn.Rect.H/2 + 7)
	text.Draw(screen, btn.Text, basicfont.Face7x13, textX, textY, textColor)
}

// DrawComparisonScreen affiche l'écran de comparaison des résultats
func (g *Game) DrawComparisonScreen(screen *ebiten.Image) {
	// Fond semi-transparent
	screen.Fill(color.RGBA{20, 25, 35, 255})

	// Titre
	titleText := "COMPARAISON DES RESULTATS"
	titleX := (750 - len(titleText)*7) / 2
	text.Draw(screen, titleText, basicfont.Face7x13,
		titleX, 40, color.RGBA{255, 200, 0, 255})

	// Ligne de séparation
	separator := ebiten.NewImage(700, 2)
	separator.Fill(color.RGBA{255, 140, 0, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(25, 55)
	screen.DrawImage(separator, op)

	// Panneau principal avec fond
	panelBg := ebiten.NewImage(700, 550)
	panelBg.Fill(color.RGBA{35, 40, 50, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(25, 80)
	screen.DrawImage(panelBg, op)

	// Bordure du panneau
	g.drawRectBorder(screen, ui.Rectangle{X: 25, Y: 80, W: 700, H: 550}, 2,
		color.RGBA{255, 140, 0, 255})

	// Afficher le texte du rapport
	g.drawWrappedText(screen, g.comparisonText, 50, 100, 650,
		color.RGBA{220, 220, 220, 255})

	// Message en bas
	instruction := "Cliquez n'importe où pour fermer"
	instX := (750 - len(instruction)*7) / 2
	text.Draw(screen, instruction, basicfont.Face7x13,
		instX, 680, color.RGBA{150, 160, 180, 255})
}

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
