# 🚦 Shibuya Crossing Simulator

**🇫🇷 Français** · [🇬🇧 English version](#english-version)

Shibuya Crossing Simulator est une simulation multi-agents reproduisant le fonctionnement d'un carrefour urbain inspiré du célèbre passage piéton de Shibuya à Tokyo.

Le système de gestion des flux du Shibuya Crossing se distingue par une organisation séquentielle en trois phases distinctes :
- Phase de circulation véhiculaire verticale (trafic piétonnier suspendu)
- Phase de circulation véhiculaire horizontale (trafic piétonnier suspendu)
- Phase de circulation piétonnière omnidirectionnelle, autorisant les traversées diagonales

Ce projet vise à étudier et comparer différentes configurations de carrefours urbains (modèle Shibuya et modèle français traditionnel) afin d'analyser les dynamiques de circulation des véhicules et des piétons. La simulation est développée en Go et s'appuie sur le moteur de jeu 2D [Ebiten](https://ebiten.org/).

<div align="center">
    <img src="docs/demo_shibuya.gif" width="100%" alt="Simulation en mode Shibuya" />
    <p><em>Un cycle complet de feux en mode Shibuya (vidéo accélérée)</em></p>
</div>

<div align="center">
    <img src="docs/screen_menu.png" width="70%" alt="Menu" />
</div>


## Principales Technologies utilisées

| Technologie | Description |
|-------------|-------------|
| **Forces Piétons/Voitures** | Répartition spatiale des différents agents, respectant une zone de distance |
| **Système de Grille des Piétons** | Algorithme d'optimisation du requêtage des piétons voisins |
| **A* Pathfinding** | Algorithme de recherche de chemin pour les déplacements des piétons |
| **Goroutines et channels** | Les feux, les spawns, la grille des piétons et l'arrêt des voitures tournent comme des microservices indépendants |
| **Serveur Web** | Affichage des graphes dynamiques (HTML, CSS, JavaScript, Chart.js, html2canvas, jsPDF) |
| **Ebitengine** | Moteur de la simulation et de l'architecture globale de la boucle de jeu |

## Fonctionnalités

- **Simulation en temps réel** : Visualisation dynamique des mouvements de véhicules et de piétons
- **Deux types de carrefours** : Comparaison entre le système Shibuya (tous piétons simultanément) et le système French (passages piétons traditionnels)
- **Métriques détaillées** :
  - Temps de traversée moyen
  - Vitesses moyennes
  - Nombre d'agents actifs et détruits
  - Durée de simulation
- **Dashboard web** : Interface de visualisation comparative avec graphiques et export PDF
- **API REST** : Récupération des métriques en temps réel
- **Configuration flexible** : Paramètres de simulation ajustables (nombre de voitures et de piétons par seconde, type de carrefour)

## Résultats

Pour comparer les deux carrefours, nous avons lancé chaque modèle pendant 60 secondes dans trois situations différentes. Voici les résultats présentés lors de notre soutenance :

| Scénario | Voitures entrées<br/>(Shibuya / French) | Temps moyen voitures<br/>(Shibuya / French) | Temps moyen piétons<br/>(Shibuya / French) |
|----------|:---:|:---:|:---:|
| Beaucoup de piétons (4 voitures/s, 50 piétons/s) | 119 / 164 | 21,8 s / 16,7 s | 25,3 s / 25,7 s |
| Beaucoup de voitures (50 voitures/s, 10 piétons/s) | 185 / 252 | 19,5 s / 17,0 s | 21,5 s / 21,7 s |
| Les deux (50 voitures/s, 50 piétons/s) | 182 / 233 | 20,1 s / 18,9 s | 24,8 s / 25,8 s |

Le carrefour français laisse passer plus de voitures, et plus rapidement, alors que pour les piétons les temps restent quasiment identiques. En revanche, le modèle Shibuya est plus sûr puisque les voitures et les piétons ne circulent jamais en même temps. En résumé, le French Crossing est plus efficace, tandis que Shibuya privilégie la sécurité.

## Prérequis

Avant de commencer, il faut s'assurer d'avoir installé :

- [Go](https://golang.org/dl/) version 1.25.0 ou supérieure
- Un navigateur web moderne (pour le dashboard)

## Installation

**Cloner le dépôt** :

```bash
git clone https://github.com/meissah/Multi-Agent_ShibuyaCrossing.git
cd Multi-Agent_ShibuyaCrossing
```

Les dépendances seront automatiquement téléchargées au premier lancement.

## Lancer la simulation

### Démarrage de la simulation

Depuis la racine du projet :

```bash
go run ./cmd/simulation
```

Dans le menu, on choisit le nombre de voitures et de piétons par seconde ainsi que le type de carrefour, puis on clique sur **Valider**. Une simulation dure au maximum 3 minutes.

### Accès aux interfaces

Une fois la simulation lancée, vous aurez accès à :

- **Fenêtre de simulation** : Interface graphique de la simulation (Ebiten), avec les métriques en temps réel à droite
- **API REST** : `http://localhost:8080`
- **Dashboard web** : Accessible via le bouton "Dashboard" du menu, une fois qu'une simulation Shibuya et une simulation French ont été sauvegardées avec le bouton "Sauvegarder les métriques"

<div align="center">
    <img src="docs/screen_simu.png" width="100%" alt="Simu" />
</div>

## Limites et améliorations possibles

- Le carrefour est simplifié et ne reprend pas vraiment la disposition du vrai Shibuya Crossing
- Les voitures ne peuvent pas encore tourner à droite (il faudrait gérer les conflits de trajectoires)
- Les piétons cherchent parfois à traverser au feu rouge, ce qui est compensé par les zones de répulsion
- La durée des feux n'est pas encore configurable depuis le menu

## Structure du projet

```
Multi-Agent_ShibuyaCrossing/
├── cmd/
│   └── simulation/
│       └── main.go              # Point d'entrée de la simulation
├── internal/
│   ├── agents/                  # Agents (piétons, véhicules, feux)
│   ├── api/                     # Serveur API REST
│   ├── assets/                  # Images utilisées par la simulation
│   ├── config/                  # Configuration de simulation
│   ├── environment/             # Zones de spawn et de répulsion
│   ├── game/                    # Logique principale du jeu
│   ├── handlers/                # Gestionnaires d'agents
│   ├── pathfinding/             # Calcul des trajets optimaux pour les piétons
│   ├── stats/                   # Calcul et comparaison des métriques
│   └── ui/                      # Composants d'interface
├── docs/                        # GIF, captures d'écran et présentation
├── dashboard.html               # Dashboard de visualisation
├── dashboard.css                # Styles du dashboard
└── README.md                    # Ce fichier
```

Le dossier `metrics_data/` est créé automatiquement lorsqu'on sauvegarde les métriques d'une simulation.


## 👥 Auteurs

Projet réalisé dans le cadre du cours AI30 à l'UTC.
- Pierrick Clause : gestion des feux de circulation, des passages piétons et du pathfinding
- Meïssa H'mida : gestion des voitures (virages avec des courbes de Bézier, spawn, force de répulsion), interface graphique et statistiques, API serveur et dashboard
- Ahmad Misselmani : interface visuelle de l'intersection, gestion des agents et des ID, microservices de spawn et d'arrêt des voitures, boucle du jeu
- Enzo Maros : gestion des piétons (interactions entre piétons, zones de répulsion, grille de recherche)

La présentation de notre soutenance est disponible ici : [presentation_AI30.pdf](docs/presentation_AI30.pdf) (template Slidesgo).


---

# English version

**🇬🇧 English** · [🇫🇷 Version française](#-shibuya-crossing-simulator)

Shibuya Crossing Simulator is a multi-agent simulation of an urban intersection, inspired by the famous Shibuya pedestrian crossing in Tokyo.

Traffic at Shibuya Crossing is managed in three distinct, sequential phases:
- Vertical vehicle traffic (pedestrians wait)
- Horizontal vehicle traffic (pedestrians wait)
- Pedestrians cross in every direction at once, including diagonally

The goal of this project is to study and compare different intersection layouts (the Shibuya model and the traditional French model) in order to analyse how vehicles and pedestrians move through them. The simulation is written in Go and built on the 2D game engine [Ebiten](https://ebiten.org/).

<div align="center">
    <img src="docs/demo_shibuya.gif" width="100%" alt="Simulation in Shibuya mode" />
    <p><em>A full traffic light cycle in Shibuya mode (sped up)</em></p>
</div>

<div align="center">
    <img src="docs/screen_menu.png" width="70%" alt="Menu" />
</div>


## Main technologies

| Technology | Description |
|------------|-------------|
| **Pedestrian/Car forces** | Spatial distribution of the agents, keeping a safe distance between them |
| **Pedestrian grid system** | Optimisation for querying neighbouring pedestrians |
| **A* Pathfinding** | Path-finding algorithm for pedestrian movement |
| **Goroutines and channels** | Traffic lights, spawning, the pedestrian grid and car stopping run as independent microservices |
| **Web server** | Dynamic charts (HTML, CSS, JavaScript, Chart.js, html2canvas, jsPDF) |
| **Ebitengine** | Simulation engine and overall game loop architecture |

## Features

- **Real-time simulation**: Live view of vehicle and pedestrian movements
- **Two intersection types**: Comparison between the Shibuya system (all pedestrians cross at once) and the French system (traditional crosswalks)
- **Detailed metrics**:
  - Average crossing time
  - Average speeds
  - Number of active and destroyed agents
  - Simulation duration
- **Web dashboard**: Side-by-side comparison with charts and PDF export
- **REST API**: Real-time access to the metrics
- **Flexible configuration**: Adjustable parameters (cars and pedestrians per second, intersection type)

## Results

To compare the two intersections, we ran each model for 60 seconds in three different situations. These are the results we presented at our final review:

| Scenario | Cars that got in<br/>(Shibuya / French) | Average car time<br/>(Shibuya / French) | Average pedestrian time<br/>(Shibuya / French) |
|----------|:---:|:---:|:---:|
| Lots of pedestrians (4 cars/s, 50 pedestrians/s) | 119 / 164 | 21.8 s / 16.7 s | 25.3 s / 25.7 s |
| Lots of cars (50 cars/s, 10 pedestrians/s) | 185 / 252 | 19.5 s / 17.0 s | 21.5 s / 21.7 s |
| Both (50 cars/s, 50 pedestrians/s) | 182 / 233 | 20.1 s / 18.9 s | 24.8 s / 25.8 s |

The French intersection lets more cars through, and faster, while pedestrian times stay almost the same. On the other hand, the Shibuya model is safer, since cars and pedestrians never move at the same time. In short, the French Crossing is more efficient, while Shibuya puts safety first.

## Prerequisites

Before getting started, make sure you have installed:

- [Go](https://golang.org/dl/) version 1.25.0 or higher
- A modern web browser (for the dashboard)

## Installation

**Clone the repository**:

```bash
git clone https://github.com/meissah/Multi-Agent_ShibuyaCrossing.git
cd Multi-Agent_ShibuyaCrossing
```

Dependencies are downloaded automatically on the first run.

## Running the simulation

### Starting the simulation

From the project root:

```bash
go run ./cmd/simulation
```

In the menu, choose the number of cars and pedestrians per second and the type of intersection, then click **Valider** (confirm). A simulation lasts at most 3 minutes.

### Accessing the interfaces

Once the simulation is running, you have access to:

- **Simulation window**: Graphical interface of the simulation (Ebiten), with live metrics on the right
- **REST API**: `http://localhost:8080`
- **Web dashboard**: Available through the "Dashboard" button in the menu, once a Shibuya run and a French run have been saved with the "Sauvegarder les métriques" (save metrics) button

<div align="center">
    <img src="docs/screen_simu.png" width="100%" alt="Simulation" />
</div>

## Limitations and possible improvements

- The intersection is simplified and doesn't really match the layout of the real Shibuya Crossing
- Cars can't turn right yet (trajectory conflicts would need to be handled)
- Pedestrians sometimes try to cross on a red light, which is compensated by the repulsion areas
- Traffic light durations can't be changed from the menu yet

## Project structure

```
Multi-Agent_ShibuyaCrossing/
├── cmd/
│   └── simulation/
│       └── main.go              # Simulation entry point
├── internal/
│   ├── agents/                  # Agents (pedestrians, vehicles, traffic lights)
│   ├── api/                     # REST API server
│   ├── assets/                  # Images used by the simulation
│   ├── config/                  # Simulation configuration
│   ├── environment/             # Spawn and repulsion areas
│   ├── game/                    # Main game logic
│   ├── handlers/                # Agent handlers
│   ├── pathfinding/             # Optimal path computation for pedestrians
│   ├── stats/                   # Metrics computation and comparison
│   └── ui/                      # Interface components
├── docs/                        # GIF, screenshots and slides
├── dashboard.html               # Visualisation dashboard
├── dashboard.css                # Dashboard styles
└── README.md                    # This file
```

The `metrics_data/` folder is created automatically when you save the metrics of a simulation.


## 👥 Authors

Project carried out as part of the AI30 course at UTC (Université de Technologie de Compiègne).
- Pierrick Clause: traffic lights, crosswalks and pathfinding
- Meïssa H'mida: cars (turns using Bézier curves, spawning, repulsion force), graphical interface and statistics, API server and dashboard
- Ahmad Misselmani: intersection visuals, agent and ID management, spawning and car-stopping microservices, game loop
- Enzo Maros: pedestrians (pedestrian interactions, repulsion areas, search grid)

Our final presentation is available here: [presentation_AI30.pdf](docs/presentation_AI30.pdf) (in French, Slidesgo template).
