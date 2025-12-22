// Ce fichier gère tout le rendu graphique (Draw) du jeu.
//
// Il contient :
//   - init() : chargement des images et ressources
//   - Draw() : router principal qui dirige vers le menu ou la simulation
//   - DrawMenu() : affichage de l'écran de configuration
//   - DrawSimulation() : affichage de la simulation (route + agents)
//   - Helpers : fonctions pour dessiner boutons, voitures, crossings, etc.
package game

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"golang.org/x/image/font/basicfont"

	trafficlight "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/traffic_light"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/environment"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Variables globales pour les ressources graphiques
var (
	imgRoad       *ebiten.Image
	carImages     []*ebiten.Image
	carColorNames []string
	pedImages     []*ebiten.Image
	pedColorNames []string
	imgShibuya    *ebiten.Image
	imgFrench     *ebiten.Image
)

// init charge toutes les images au démarrage de l'application
func init() {
	var err error

	// Charger l'image de la route
	imgRoad, _, err = ebitenutil.NewImageFromFile("internal/assets/road.png")
	if err != nil {
		log.Fatal(err)
	}

	// Charger les images des voitures (une par couleur)
	carColorNames = []string{"nil", "blue", "green", "purple", "lightblue", "yellow"}
	for _, color := range carColorNames {
		filename := fmt.Sprintf("internal/assets/cars/car_%s.png", color)
		img, _, err := ebitenutil.NewImageFromFile(filename)
		if err != nil {
			log.Fatalf("Erreur: Impossible de charger %s: %v", filename, err)
		}
		carImages = append(carImages, img)
		log.Printf("✓ Image chargée: %s", filename)
	}

	// Charger les images des pietons (une par couleur)
	pedColorNames = []string{"nil", "zero"}
	for _, color := range pedColorNames {
		filename := fmt.Sprintf("internal/assets/ped_%s.png", color)
		img, _, err := ebitenutil.NewImageFromFile(filename)
		if err != nil {
			log.Fatalf("Erreur: Impossible de charger %s: %v", filename, err)
		}
		pedImages = append(pedImages, img)
		log.Printf("✓ Image chargée: %s", filename)
	}

	// Charger les images pour le menu (optionnel)
	imgShibuya, _, _ = ebitenutil.NewImageFromFile("internal/assets/menu/shibuya.png")
	imgFrench, _, _ = ebitenutil.NewImageFromFile("internal/assets/menu/french.png")
}

// Draw est appelé autant de fois que l'écran le permet
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.State {
	case config.StateMenu:
		g.DrawMenu(screen)
	case config.StateSimulation:
		g.DrawSimulation(screen)
	}
}

// ───────────────────────────────────────────────────────────────────
// MENU : Affichage de l'écran de configuration
// ───────────────────────────────────────────────────────────────────

// DrawMenu affiche l'écran de menu avec les options de configuration
func (g *Game) DrawMenu(screen *ebiten.Image) {
	// Fond dégradé  (bleu-gris foncé)
	screen.Fill(color.RGBA{25, 30, 40, 255})

	// ═══════════════════════════════════════════════════════════
	// TITRE PRINCIPAL centré
	// ═══════════════════════════════════════════════════════════

	titleText := "CROSSING SIMULATOR"
	titleWidth := len(titleText) * 7
	titleX := (750 - titleWidth) / 2

	// Dessiner l'ombre du titre (effet de profondeur)
	for i := 0; i < 3; i++ {
		text.Draw(screen, titleText, basicfont.Face7x13,
			titleX+i, 48+i, color.RGBA{0, 0, 0, 100})
	}
	// Dessiner le titre principal en blanc
	text.Draw(screen, titleText, basicfont.Face7x13,
		titleX, 48, color.RGBA{255, 255, 255, 255})

	// Sous-titre centré
	subtitleText := "Comparaison Shibuya vs French Crossing"
	subtitleWidth := len(subtitleText) * 7
	subtitleX := (750 - subtitleWidth) / 2
	text.Draw(screen, subtitleText, basicfont.Face7x13,
		subtitleX, 70, color.RGBA{150, 160, 180, 255})

	// Ligne de séparation centrée
	separatorLine := ebiten.NewImage(600, 2)
	separatorLine.Fill(color.RGBA{50, 140, 200, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(75, 85)
	screen.DrawImage(separatorLine, op)

	// ═══════════════════════════════════════════════════════════
	// SECTION PARAMÈTRES
	// ═══════════════════════════════════════════════════════════

	// Titre de section avec fond coloré
	sectionBg := ebiten.NewImage(650, 35)
	sectionBg.Fill(color.RGBA{40, 50, 65, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(50, 110)
	screen.DrawImage(sectionBg, op)

	// Titre de section centré
	sectionTitle := "PARAMETRES DE SIMULATION"
	sectionTitleWidth := len(sectionTitle) * 7
	sectionTitleX := (750 - sectionTitleWidth) / 2
	text.Draw(screen, sectionTitle, basicfont.Face7x13,
		sectionTitleX, 130, color.RGBA{100, 200, 255, 255})

	// Slider voitures centré avec icône
	labelCars := "Voitures par seconde:"
	text.Draw(screen, labelCars, basicfont.Face7x13,
		100, 175, color.RGBA{220, 220, 220, 255})
	g.drawSlider(screen, g.carSpawnSlider)
	carValue := fmt.Sprintf("%d/s", int(g.carSpawnSlider.Value))
	text.Draw(screen, carValue, basicfont.Face7x13,
		555, 175, color.RGBA{100, 200, 100, 255})

	// Slider piétons
	labelPed := "Pietons par seconde:"
	text.Draw(screen, labelPed, basicfont.Face7x13,
		100, 215, color.RGBA{220, 220, 220, 255})
	g.drawSlider(screen, g.pedSpawnSlider)
	pedValue := fmt.Sprintf("%d/s", int(g.pedSpawnSlider.Value))
	text.Draw(screen, pedValue, basicfont.Face7x13,
		555, 215, color.RGBA{100, 200, 100, 255})

	// ═══════════════════════════════════════════════════════════
	// SECTION CHOIX DU CROSSING
	// ═══════════════════════════════════════════════════════════

	// Titre de section
	sectionBg = ebiten.NewImage(650, 35)
	sectionBg.Fill(color.RGBA{40, 50, 65, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(50, 260)
	screen.DrawImage(sectionBg, op)

	// Titre centré
	crossingTitle := "CHOISISSEZ VOTRE TYPE DE CARREFOUR"
	crossingTitleWidth := len(crossingTitle) * 7
	crossingTitleX := (750 - crossingTitleWidth) / 2
	text.Draw(screen, crossingTitle, basicfont.Face7x13,
		crossingTitleX, 290, color.RGBA{100, 200, 255, 255})

	// Instructions centrées
	instructionText := "Cliquez sur une option ci-dessous"
	instructionWidth := len(instructionText) * 7
	instructionX := (750 - instructionWidth) / 2
	text.Draw(screen, instructionText, basicfont.Face7x13,
		instructionX, 320, color.RGBA{150, 160, 180, 255})

	// Dessiner les deux options de crossing côté à côté
	g.drawCrossingOption(screen, "Shibuya Crossing",
		g.shibuyaRect, // Utilise directement le rectangle de détection
		g.selectedCrossing == "shibuya", imgShibuya)

	g.drawCrossingOption(screen, "French Crossing",
		g.frenchRect, // Utilise directement le rectangle de détection
		g.selectedCrossing == "French", imgFrench)

	// ═══════════════════════════════════════════════════════════
	// BOUTON VALIDER (Actif seulement si un crossing est sélectionné)
	// ═══════════════════════════════════════════════════════════

	// Repositionner le bouton (centré)
	g.validateButton.Rect.X = 275
	g.validateButton.Rect.Y = 620
	g.validateButton.Rect.W = 200
	g.validateButton.Rect.H = 50

	g.drawButton(screen, g.validateButton, g.selectedCrossing != "")

	// Message si aucune sélection
	if g.selectedCrossing == "" {
		helpText := "Selectionnez un crossing pour continuer"
		helpWidth := len(helpText) * 7
		helpX := (750 - helpWidth) / 2
		text.Draw(screen, helpText, basicfont.Face7x13,
			helpX, 700, color.RGBA{255, 100, 100, 255})
	}
}

// drawCrossingOption dessine une option de crossing
func (g *Game) drawCrossingOption(screen *ebiten.Image, name string, rect ui.Rectangle, selected bool, img *ebiten.Image) {
	// Couleur de bordure selon la sélection
	var borderColor color.Color
	var bgColor color.Color

	if selected {
		borderColor = color.RGBA{0, 200, 120, 255} // Vert brillant
		bgColor = color.RGBA{30, 50, 45, 255}      // Fond légèrement vert
	} else {
		borderColor = color.RGBA{80, 90, 100, 255} // Gris
		bgColor = color.RGBA{40, 45, 55, 255}      // Fond sombre
	}

	// Fond du cadre
	bgImg := ebiten.NewImage(int(rect.W), int(rect.H))
	bgImg.Fill(bgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(rect.X, rect.Y)
	screen.DrawImage(bgImg, op)

	// Image du crossing
	if img != nil {
		op := &ebiten.DrawImageOptions{}
		targetW := rect.W - 30
		targetH := rect.H - 30
		imgW, imgH := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())

		scaleX := targetW / imgW
		scaleY := targetH / imgH
		scale := math.Min(scaleX, scaleY)

		scaledW := imgW * scale
		scaledH := imgH * scale
		offsetX := (rect.W - scaledW) / 2
		offsetY := (rect.H - scaledH) / 2

		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(rect.X+offsetX, rect.Y+offsetY)
		screen.DrawImage(img, op)
	} else {
		placeholderText := "[Apercu]"
		text.Draw(screen, placeholderText, basicfont.Face7x13,
			int(rect.X+rect.W/2-35), int(rect.Y+rect.H/2),
			color.RGBA{120, 130, 140, 255})
	}

	// Bordure avec effet lumineux si sélectionné
	thickness := 2.0
	if selected {
		thickness = 4.0
		// Effet de lueur
		g.drawRectBorder(screen, rect, thickness+2, color.RGBA{0, 255, 150, 100})
	}
	g.drawRectBorder(screen, rect, thickness, borderColor)

	// Label centré en dessous
	labelColor := color.RGBA{180, 190, 200, 255}
	if selected {
		labelColor = color.RGBA{0, 255, 150, 255}
	}

	textX := int(rect.X + rect.W/2 - float64(len(name)*7)/2)
	text.Draw(screen, name, basicfont.Face7x13,
		textX, int(rect.Y+rect.H+25), labelColor)
}

// drawButton dessine un bouton avec état hover (bouton validé dans le menu par exemple)
func (g *Game) drawButton(screen *ebiten.Image, btn ui.Button, enabled bool) {
	var bgColor, textColor color.Color

	if !enabled {
		// Désactivé
		bgColor = color.RGBA{60, 65, 70, 255}
		textColor = color.RGBA{100, 105, 110, 255}
	} else if btn.IsMouseOver {
		// Hover
		bgColor = color.RGBA{0, 220, 130, 255}
		textColor = color.RGBA{255, 255, 255, 255}

		// Effet de lueur au survol
		glowBtn := ebiten.NewImage(int(btn.Rect.W+10), int(btn.Rect.H+10))
		glowBtn.Fill(color.RGBA{0, 255, 150, 40})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(btn.Rect.X-5, btn.Rect.Y-5)
		screen.DrawImage(glowBtn, op)
	} else {
		// Normal activé
		bgColor = color.RGBA{0, 180, 100, 255}
		textColor = color.RGBA{255, 255, 255, 255}
	}

	// Fond du bouton
	btnImg := ebiten.NewImage(int(btn.Rect.W), int(btn.Rect.H))
	btnImg.Fill(bgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(btn.Rect.X, btn.Rect.Y)
	screen.DrawImage(btnImg, op)

	// Bordure du bouton
	borderColor := color.RGBA{0, 100, 60, 255}
	if enabled {
		borderColor = color.RGBA{0, 255, 150, 255}
	}
	g.drawRectBorder(screen, btn.Rect, 2, borderColor)

	// Texte centré verticalement et horizontalement
	textX := int(btn.Rect.X + btn.Rect.W/2 - float64(len(btn.Text)*7)/2)
	textY := int(btn.Rect.Y + btn.Rect.H/2 + 5)
	text.Draw(screen, btn.Text, basicfont.Face7x13, textX, textY, textColor)
}

// drawSlider dessine un slider avec sa barre et son curseur
func (g *Game) drawSlider(screen *ebiten.Image, slider ui.Slider) {
	if slider.Width <= 0 || slider.Height <= 0 {
		return
	}

	// Barre de fond avec bordure
	track := ebiten.NewImage(int(slider.Width), int(slider.Height))
	track.Fill(color.RGBA{30, 35, 45, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(slider.X, slider.Y)
	screen.DrawImage(track, op)

	// Bordure de la barre
	g.drawRectBorder(screen, ui.Rectangle{
		X: slider.X, Y: slider.Y,
		W: slider.Width, H: slider.Height,
	}, 1, color.RGBA{60, 70, 80, 255})

	// Partie remplie (gradient)
	handleX := slider.GetHandleX()
	filledWidth := handleX - slider.X
	if filledWidth > 0 {
		filled := ebiten.NewImage(int(filledWidth), int(slider.Height))
		filled.Fill(color.RGBA{60, 130, 180, 255})
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(slider.X, slider.Y)
		screen.DrawImage(filled, op)
	}

	// Curseur (cercle) avec effet de profondeur
	handleSize := 12.0

	// Ombre du curseur
	shadow := ebiten.NewImage(int(handleSize+4), int(handleSize+4))
	shadow.Fill(color.RGBA{0, 0, 0, 100})
	op = &ebiten.DrawImageOptions{}
	shadowX := handleX - handleSize/2 + 2
	shadowY := slider.Y + slider.Height/2 - handleSize/2 + 2
	shadowX = math.Max(slider.X, math.Min(slider.X+slider.Width-handleSize, shadowX))
	op.GeoM.Translate(shadowX, shadowY)
	screen.DrawImage(shadow, op)

	// Curseur principal
	handle := ebiten.NewImage(int(handleSize), int(handleSize))
	handle.Fill(color.RGBA{255, 255, 255, 255})
	op = &ebiten.DrawImageOptions{}
	handleXPos := handleX - handleSize/2
	handleXPos = math.Max(slider.X, math.Min(slider.X+slider.Width-handleSize, handleXPos))
	handleY := slider.Y + slider.Height/2 - handleSize/2
	op.GeoM.Translate(handleXPos, handleY)
	screen.DrawImage(handle, op)

	// Bordure du curseur
	g.drawRectBorder(screen, ui.Rectangle{
		X: handleXPos, Y: handleY,
		W: handleSize, H: handleSize,
	}, 1.5, color.RGBA{0, 200, 255, 255})
}

// drawRectBorder dessine une bordure autour d'un rectangle
func (g *Game) drawRectBorder(screen *ebiten.Image, rect ui.Rectangle, thickness float64, col color.Color) {
	// Bordure haute et basse
	horizontal := ebiten.NewImage(int(rect.W), int(thickness))
	horizontal.Fill(col)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(rect.X, rect.Y)
	screen.DrawImage(horizontal, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(rect.X, rect.Y+rect.H-thickness)
	screen.DrawImage(horizontal, op)

	// Bordure gauche et droite
	vertical := ebiten.NewImage(int(thickness), int(rect.H))
	vertical.Fill(col)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(rect.X, rect.Y)
	screen.DrawImage(vertical, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(rect.X+rect.W-thickness, rect.Y)
	screen.DrawImage(vertical, op)
}

// ───────────────────────────────────────────────────────────────────
// SIMULATION : Affichage de la simulation (route + agents)
// ───────────────────────────────────────────────────────────────────

// DrawSimulation affiche l'état actuel de la simulation (route + agents)
func (g *Game) DrawSimulation(screen *ebiten.Image) {
	// Fond beige
	screen.Fill(color.RGBA{126, 109, 97, 255})

	// Dessiner le crossing selon le type choisi
	switch g.Config.CrossingType {
	case "shibuya":
		g.drawShibuyaCrossing(screen)
	case "French":
		g.drawFrenchCrossing(screen)
	}

	// Dessiner tous les objets (voitures, piétons, etc.)
	currentObjtSlice := g.Objets
	for _, obj := range currentObjtSlice {
		switch obj.Type() {
		case "car":
			g.drawCar(screen, obj.(Drawable))
		case "pieton":
			g.drawPedestrian(screen, obj.(Drawable))
		case "traffic_light_car":
			X, Y := obj.PositionX(), obj.PositionY()
			dir := obj.Direction()
			isHorizontal := dir == East || dir == West

			// Dimensions "Standard"
			boxW, boxH := 24, 48
			rDx, rDy := 0.0, -12.0
			gDx, gDy := 0.0, +12.0

			if isHorizontal {
				boxW, boxH = 48, 24
				rDx, rDy = -12.0, 0.0
				gDx, gDy = +12.0, 0.0
			}

			// Dessin Boitier
			box := ebiten.NewImage(boxW, boxH)
			box.Fill(color.RGBA{50, 50, 50, 255})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(X-float64(boxW)/2, Y-float64(boxH)/2)
			screen.DrawImage(box, op)

			// Dessin Cercles
			state := obj.(*trafficlight.TrafficLightObject).State()
			radius := float32(7) // Gros cercles

			// Couleurs
			onRed, onGreen := color.RGBA{255, 0, 0, 255}, color.RGBA{0, 200, 0, 255}
			off := color.RGBA{70, 70, 70, 255}

			rCol, gCol := off, off
			if state == "Red" {
				rCol = onRed
			}
			if state == "Green" {
				gCol = onGreen
			}

			vector.FillCircle(screen, float32(X+rDx), float32(Y+rDy), radius, rCol, false)
			vector.FillCircle(screen, float32(X+gDx), float32(Y+gDy), radius, gCol, false)

		case "traffic_light_pedestrian":
			// Pour accéder aux champs W et H, on doit "caster" l'interface Object
			// vers son type concret *trafficlight.TrafficLightObject
			tlObj, ok := obj.(*trafficlight.TrafficLightObject)
			if !ok {
				continue // Sécurité si le cast échoue
			}

			cX, cY := tlObj.PositionX(), tlObj.PositionY() // Centre
			w, h := tlObj.Size()                           // Dimensions

			// Définition des couleurs avec transparence (Alpha = 100 sur 255)
			// Cela permet de voir le passage piéton dessous.
			redTransparent := color.RGBA{255, 0, 0, 100}
			greenTransparent := color.RGBA{0, 255, 0, 100}

			var drawColor color.Color

			switch tlObj.State() {
			case "Red":
				drawColor = redTransparent
			case "Green":
				drawColor = greenTransparent
			default:
				// Si l'état est inconnu, on ne dessine rien (invisible)
				continue
			}

			// vector.DrawFilledRect prend les coordonnées du coin supérieur gauche.
			// On convertit depuis le centre :
			topLeftX := float32(cX - w/2)
			topLeftY := float32(cY - h/2)

			// On dessine le rectangle coloré semi-transparent
			vector.DrawFilledRect(screen, topLeftX, topLeftY, float32(w), float32(h), drawColor, false)
		case "repulsion_area":
			obj.(*environment.RepulsionArea).Draw(screen)
		}

	}
	// Dessiner le panneau de métriques (pour les statistiques en temps réel)
	g.DrawMetricsPanel(screen)

}

// drawShibuyaCrossing dessine le carrefour de type Shibuya (avec diagonale)
func (g *Game) drawShibuyaCrossing(screen *ebiten.Image) {
	// Route verticale
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Translate(300, 0)
	screen.DrawImage(imgRoad, op)

	// Route horizontale
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Rotate(math.Pi / 2)
	op.GeoM.Translate(750, 300)
	screen.DrawImage(imgRoad, op)

	// Intersection centrale
	img1 := ebiten.NewImage(260, 150)
	img1.Fill(color.RGBA{71, 71, 71, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(245, 300)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(150, 260)
	img1.Fill(color.RGBA{71, 71, 71, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(300, 245)
	screen.DrawImage(img1, op)

	// Bandes blanches verticales
	stripes := ebiten.NewImage(55, 5)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for y := 300.0; y < 445; y += 7.5 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(245, y)
		screen.DrawImage(stripes, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(450, y)
		screen.DrawImage(stripes, op)
	}

	// Bandes blanches horizontales
	stripes = ebiten.NewImage(5, 55)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for x := 300.0; x < 445; x += 7.5 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, 245)
		screen.DrawImage(stripes, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, 450)
		screen.DrawImage(stripes, op)
	}

	// Bandes blanches diagonales (spécifique Shibuya)
	stripes = ebiten.NewImage(5, 55)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for x := 0.0; x < 22; x++ {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Rotate(math.Pi / 4)
		op.GeoM.Translate(340.0+x*5.0, 300.0+x*5.0)
		screen.DrawImage(stripes, op)
	}

	// Trottoirs
	g.drawSidewalks(screen)
}

// drawFrenchCrossing dessine le carrefour de type French (sans diagonale)
func (g *Game) drawFrenchCrossing(screen *ebiten.Image) {
	// Route verticale
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Translate(300, 0)
	screen.DrawImage(imgRoad, op)

	// Route horizontale
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15075, .392465)
	op.GeoM.Rotate(math.Pi / 2)
	op.GeoM.Translate(750, 300)
	screen.DrawImage(imgRoad, op)

	// Intersection centrale
	img1 := ebiten.NewImage(260, 150)
	img1.Fill(color.RGBA{71, 71, 71, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(245, 300)
	screen.DrawImage(img1, op)

	img1 = ebiten.NewImage(150, 260)
	img1.Fill(color.RGBA{71, 71, 71, 255})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(300, 245)
	screen.DrawImage(img1, op)

	// Bandes blanches verticales
	stripes := ebiten.NewImage(55, 5)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for y := 300.0; y < 445; y += 7.5 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(245, y)
		screen.DrawImage(stripes, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(450, y)
		screen.DrawImage(stripes, op)
	}

	// Bandes blanches horizontales
	stripes = ebiten.NewImage(5, 55)
	stripes.Fill(color.RGBA{255, 255, 255, 255})
	for x := 300.0; x < 445; x += 7.5 {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, 245)
		screen.DrawImage(stripes, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, 450)
		screen.DrawImage(stripes, op)
	}

	// Trottoirs
	g.drawSidewalks(screen)
}

// drawSidewalks dessine les trottoirs autour du carrefour
func (g *Game) drawSidewalks(screen *ebiten.Image) {
	// Trottoirs verticaux
	sidewalk := ebiten.NewImage(55, 300)
	sidewalk.Fill(color.RGBA{128, 128, 128, 255})

	positions := [][2]float64{
		{245, 0}, {245, 450}, {450, 0}, {450, 450},
	}
	for _, pos := range positions {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pos[0], pos[1])
		screen.DrawImage(sidewalk, op)
	}

	// Trottoirs horizontaux
	sidewalk = ebiten.NewImage(300, 55)
	sidewalk.Fill(color.RGBA{128, 128, 128, 255})

	positions = [][2]float64{
		{0, 245}, {450, 245}, {0, 450}, {450, 450},
	}
	for _, pos := range positions {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pos[0], pos[1])
		screen.DrawImage(sidewalk, op)
	}
}

// drawCar dessine une voiture avec rotation selon sa direction et son virage
func (g *Game) drawCar(screen *ebiten.Image, c Drawable) {
	assetIndex, op := c.Draw()

	// Dessiner la voiture
	screen.DrawImage(carImages[assetIndex], op)

	// VISUALISATION DU SYSTÈME DE RALENTISSEMENT
	// Cast en *Car pour accéder aux méthodes spécifiques
	if car, ok := c.(*vehicule.Car); ok {
		// 1. Afficher la vitesse au-dessus de la voiture
		speed := car.GetSpeed()
		speedText := fmt.Sprintf("%.1f", speed)
		speedColor := color.RGBA{0, 255, 0, 255} // Vert par défaut

		// Changer la couleur selon la vitesse
		if speed < 1.0 {
			speedColor = color.RGBA{255, 0, 0, 255} // Rouge si très lent
		} else if speed < 1.5 {
			speedColor = color.RGBA{255, 165, 0, 255} // Orange si ralenti
		}

		text.Draw(screen, speedText, basicfont.Face7x13,
			int(car.PositionX())-10, int(car.PositionY())-15, speedColor)

		// 2. Dessiner une ligne vers la voiture détectée
		carAhead, distance := car.GetCarAhead()
		if carAhead != nil && distance > 0 {
			// Récupérer la position de la voiture devant
			// Couleur de la ligne selon la distance
			lineColor := color.RGBA{255, 255, 0, 180} // Jaune
			if distance < 70 {
				lineColor = color.RGBA{255, 0, 0, 200} // Rouge si trop proche
			}

			// Dessiner la ligne de détection
			ebitenutil.DrawLine(screen,
				car.PositionX(), car.PositionY(),
				carAhead.PositionX(), carAhead.PositionY(),
				lineColor)

			// Afficher la distance
			distText := fmt.Sprintf("%.0f", distance)
			midX := (car.PositionX() + carAhead.PositionX()) / 2
			midY := (car.PositionY() + carAhead.PositionY()) / 2
			text.Draw(screen, distText, basicfont.Face7x13,
				int(midX)-10, int(midY), color.RGBA{255, 255, 255, 255})
		}
	}
}

func (g *Game) drawPedestrian(screen *ebiten.Image, c Drawable) {
	assetIndex, op := c.Draw()

	// Centrer l'image sur la position
	op.GeoM.Translate(-float64(pedImages[assetIndex].Bounds().Dx())/2,
		-float64(pedImages[assetIndex].Bounds().Dy())/2)

	// Dessiner
	screen.DrawImage(pedImages[assetIndex], op)
}
