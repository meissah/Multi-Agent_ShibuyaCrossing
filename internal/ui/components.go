package ui

import "math"

// Package ui fournit les composants d'interface graphique réutilisables.
// Ce package contient les structures de base pour créer des éléments
// interactifs (boutons, zones cliquables) utilisés dans le menu de confif et intreface de la simu

// Rectangle pour détecter les clics et définir des zones
type Rectangle struct {
	X, Y, W, H float64
}

// Contains vérifie si un point (x,y) est à l'intérieur du rectangle
func (r Rectangle) Contains(x, y float64) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

// Button représente un bouton cliquable
type Button struct {
	Rect        Rectangle
	Text        string
	IsMouseOver bool   // souris au dessus "survolé"
	Action      func() // Fonction optionnelle à exécuter au clic
}

// NewButton crée un nouveau bouton
func NewButton(x, y, w, h float64, text string) Button {
	return Button{
		Rect: Rectangle{X: x, Y: y, W: w, H: h},
		Text: text,
	}
}

// UpdateHover met à jour l'état hover du bouton selon la position de la souris
func (b *Button) UpdateHover(mouseX, mouseY float64) {
	b.IsMouseOver = b.Rect.Contains(mouseX, mouseY)
}

// IsClicked vérifie si le bouton est cliqué (helper function)
func (b *Button) IsClicked(mouseX, mouseY float64, clicked bool) bool {
	return b.IsMouseOver && clicked
}

// Slider représente une barre de défilement pour choisir une valeur
type Slider struct {
	X, Y          float64 // Position du slider
	Width, Height float64 // Dimensions
	Min, Max      float64 // Valeurs min et max
	Value         float64 // Valeur actuelle
	IsDragging    bool    // true si on est en train de glisser
}

// NewSlider crée un nouveau slider
func NewSlider(x, y, width, height, min, max, initialValue float64) Slider {
	return Slider{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Min:    min,
		Max:    max,
		Value:  initialValue,
	}
}

// Update met à jour le slider selon la souris
func (s *Slider) Update(mouseX, mouseY float64, mousePressed bool) {
	if mousePressed {
		if mouseX >= s.X && mouseX <= s.X+s.Width &&
			mouseY >= s.Y && mouseY <= s.Y+s.Height {
			s.IsDragging = true
		}
	} else {
		s.IsDragging = false
	}

	if s.IsDragging {
		relativeX := mouseX - s.X
		relativeX = math.Max(0, math.Min(s.Width, relativeX))
		ratio := relativeX / s.Width
		s.Value = s.Min + ratio*(s.Max-s.Min)
	}
}

// GetHandleX retourne la position X du curseur du slider
func (s *Slider) GetHandleX() float64 {
	ratio := (s.Value - s.Min) / (s.Max - s.Min)
	// Limiter ratio entre 0 et 1 pour éviter les dépassements visuels
	ratio = math.Max(0, math.Min(1, ratio))
	return s.X + ratio*s.Width
}
