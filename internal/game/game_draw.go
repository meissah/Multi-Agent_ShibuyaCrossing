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

	"golang.org/x/image/font/basicfont"

	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/config"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/ui"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// Variables globales pour les ressources graphiques
var (
	imgRoad       *ebiten.Image
	carImages     []*ebiten.Image
	carColorNames []string
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
	carColorNames = []string{"blue", "green", "purple", "lightblue", "yellow"}
	for _, color := range carColorNames {
		filename := fmt.Sprintf("internal/assets/cars/car_%s.png", color)
		img, _, err := ebitenutil.NewImageFromFile(filename)
		if err != nil {
			log.Fatalf("Erreur: Impossible de charger %s: %v", filename, err)
		}
		carImages = append(carImages, img)
		log.Printf("✓ Image chargée: %s", filename)
	}

	// Charger les images pour le menu (optionnel)
	imgShibuya, _, _ = ebitenutil.NewImageFromFile("internal/assets/menu/shibuya.png")
	imgFrench, _, _ = ebitenutil.NewImageFromFile("internal/assets/menu/french.png")
}

// Draw est appelé 60 fois parfdrawSlider seconde par Ebiten pour afficher le jeu
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
	// Fond sombre
	screen.Fill(color.RGBA{40, 40, 50, 255})

	// Titre principal
	text.Draw(screen, "Shibuya Crossing Simulator", basicfont.Face7x13, 250, 50, color.White)
	text.Draw(screen, "Parametres de simulation", basicfont.Face7x13, 50, 100, color.White)

	/// Slider voitures
	text.Draw(screen, "Voitures par seconde:", basicfont.Face7x13, 50, 130, color.RGBA{255, 255, 255, 255})
	g.drawSlider(screen, g.carSpawnSlider)
	carValue := fmt.Sprintf("%d/s", int(g.carSpawnSlider.Value)) // arrondi en int
	text.Draw(screen, carValue, basicfont.Face7x13, 520, 130, color.RGBA{200, 200, 200, 255})

	// Slider piétons
	text.Draw(screen, "Pietons par seconde:", basicfont.Face7x13, 50, 170, color.RGBA{255, 255, 255, 255})
	g.drawSlider(screen, g.pedSpawnSlider)
	pedValue := fmt.Sprintf("%d/s", int(g.pedSpawnSlider.Value)) // arrondi en int
	text.Draw(screen, pedValue, basicfont.Face7x13, 520, 170, color.RGBA{200, 200, 200, 255})
	// Instruction de sélection
	text.Draw(screen, "Choisissez un type de crossing:", basicfont.Face7x13, 50, 250, color.White)

	// Dessiner les deux options de crossing
	g.drawCrossingOption(screen, "Shibuya", g.shibuyaRect, g.selectedCrossing == "shibuya", imgShibuya)
	g.drawCrossingOption(screen, "French", g.frenchRect, g.selectedCrossing == "french", imgFrench)

	// Dessiner le bouton de validation
	g.drawButton(screen, g.validateButton, g.selectedCrossing != "")
}

// drawCrossingOption dessine une option de crossing (cadre + image + label)
func (g *Game) drawCrossingOption(screen *ebiten.Image, name string, rect ui.Rectangle, selected bool, img *ebiten.Image) {
	// Choisir la couleur de bordure selon la sélection
	borderColor := color.RGBA{100, 100, 100, 255}
	if selected {
		borderColor = color.RGBA{23, 182, 128, 1} // Vert si sélectionné
	}

	// Dessiner le fond du cadre
	bgImg := ebiten.NewImage(int(rect.W), int(rect.H))
	bgImg.Fill(color.RGBA{60, 60, 70, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(rect.X, rect.Y)
	screen.DrawImage(bgImg, op)

	// Dessiner l'image du crossing à l'intérieur du cadre
	if img != nil {
		op := &ebiten.DrawImageOptions{}

		// Calculer l'échelle pour remplir le rectangle (avec marge de 10px)
		targetW := rect.W - 20 // marge de 10px de chaque côté
		targetH := rect.H - 20
		imgW, imgH := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())

		// Prendre l'échelle minimale pour garder les proportions
		scaleX := targetW / imgW
		scaleY := targetH / imgH
		scale := math.Min(scaleX, scaleY)

		// Centrer l'image
		scaledW := imgW * scale
		scaledH := imgH * scale
		offsetX := (rect.W - scaledW) / 2
		offsetY := (rect.H - scaledH) / 2

		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(rect.X+offsetX, rect.Y+offsetY)
		screen.DrawImage(img, op)
	} else {
		// Placeholder si pas d'image
		text.Draw(screen, "[Image]", basicfont.Face7x13, int(rect.X+50), int(rect.Y+70), color.RGBA{255, 255, 255, 255})
	}

	// Dessiner la bordure
	g.drawRectBorder(screen, rect, 3, borderColor)

	// Label en dessous du cadre
	text.Draw(screen, name, basicfont.Face7x13, int(rect.X+40), int(rect.Y+rect.H+20), color.White)
}

// drawButton dessine un bouton avec état hover (bouton validé dans le menu par exemple)
func (g *Game) drawButton(screen *ebiten.Image, btn ui.Button, enabled bool) {
	bgColor := color.RGBA{80, 80, 90, 255} // Gris par défaut (désactivé)
	textColor := color.RGBA{150, 150, 150, 255}

	if enabled {
		if btn.IsMouseOver {
			bgColor = color.RGBA{0, 200, 0, 255} // Vert clair (hover)
		} else {
			bgColor = color.RGBA{0, 150, 0, 255} // Vert normal
		}
		textColor = color.RGBA{255, 255, 255, 255} // Blanc
	}

	// Dessiner le fond du bouton
	btnImg := ebiten.NewImage(int(btn.Rect.W), int(btn.Rect.H))
	btnImg.Fill(bgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(btn.Rect.X, btn.Rect.Y)
	screen.DrawImage(btnImg, op)

	// Dessiner le texte centré
	text.Draw(screen, btn.Text, basicfont.Face7x13,
		int(btn.Rect.X+30), int(btn.Rect.Y+25), textColor)
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

// drawSlider dessine un slider avec sa barre et son curseur
func (g *Game) drawSlider(screen *ebiten.Image, slider ui.Slider) {
	// Vérifier que les dimensions sont valides
	if slider.Width <= 0 || slider.Height <= 0 {
		fmt.Printf("Slider invalide: W=%.1f, H=%.1f\n", slider.Width, slider.Height)
		return
	}

	// Barre de fond (gris foncé)
	track := ebiten.NewImage(int(slider.Width), int(slider.Height))
	track.Fill(color.RGBA{60, 60, 60, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(slider.X, slider.Y)
	screen.DrawImage(track, op)

	// Curseur (cercle blanc)
	handleSize := 12.0
	handle := ebiten.NewImage(int(handleSize), int(handleSize))
	handle.Fill(color.RGBA{255, 255, 255, 255})
	op = &ebiten.DrawImageOptions{}
	handleX := slider.GetHandleX() - handleSize/2
	handleY := slider.Y + slider.Height/2 - handleSize/2
	op.GeoM.Translate(handleX, handleY)
	screen.DrawImage(handle, op)
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
	case "Shibuya":
		g.drawShibuyaCrossing(screen)
	case "French":
		g.drawFrenchCrossing(screen)
	}

	// Dessiner tous les objets (voitures, piétons, etc.)
	for _, obj := range g.Objts {
		switch obj.Type() {
		case "car":
			g.drawCar(screen, obj)
			// case "pedestrian":
			//     g.drawPedestrian(screen, obj)
		}
	}

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
	// A faire : Implémenter les spécificités du French crossing -> apres avec le behavore ca va changer mais c'est juste que j'ai aps la vision mtn
	// Pour l'instant, même chose que Shibuya
	g.drawShibuyaCrossing(screen)
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
func (g *Game) drawCar(screen *ebiten.Image, c Object) {
	car, ok := c.(*vehicule.Car)
	if !ok {
		return
	}

	// Récupérer l'image de la voiture selon sa couleur
	colorIndex := car.CarColorIndex()
	carImg := carImages[colorIndex]
	bounds := carImg.Bounds()
	imgW, imgH := float64(bounds.Dx()), float64(bounds.Dy())

	op := &ebiten.DrawImageOptions{}
	var w, h int
	var scaleX, scaleY float64
	angle := c.Angle()

	// Appliquer la rotation de base selon la direction
	switch c.Direction() {
	case South:
		w, h = 35, 50
		op.GeoM.Rotate(math.Pi)
		op.GeoM.Translate(imgW, imgH)
		scaleX = float64(w) / imgW
		scaleY = float64(h) / imgH
	case North:
		w, h = 35, 50
		scaleX = float64(w) / imgW
		scaleY = float64(h) / imgH
	case West:
		w, h = 50, 35
		op.GeoM.Rotate((3 * math.Pi) / 2)
		op.GeoM.Translate(0, imgW)
		scaleY = float64(h) / imgW
		scaleX = float64(w) / imgH
	case East:
		w, h = 50, 35
		op.GeoM.Rotate(math.Pi / 2)
		op.GeoM.Translate(imgH, 0)
		scaleY = float64(h) / imgW
		scaleX = float64(w) / imgH
	}

	// Appliquer le scale et centrer
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	// Récupérer la position
	x, y := c.Position()

	// Appliquer la rotation du virage si nécessaire
	if angle != 0 {
		if c.Direction() == West || c.Direction() == South {
			op.GeoM.Rotate(angle + math.Pi)
		} else {
			op.GeoM.Rotate(angle)
		}
	}

	// Positionner à la position finale
	op.GeoM.Translate(x, y)

	// Dessiner
	screen.DrawImage(carImg, op)
}
