# 🚦 Shibuya Crossing Simulator

**🇫🇷 Français** · [🇬🇧 English version](#english-version)

Shibuya Crossing Simulator est une simulation multi-agents reproduisant le fonctionnement d'un carrefour urbain inspiré du célèbre passage piéton de Shibuya à Tokyo. 

Le système de gestion des flux du Shibuya Crossing se distingue par une organisation séquentielle en trois phases distinctes :
- Phase de circulation véhiculaire verticale (trafic piétonnier suspendu)
- Phase de circulation véhiculaire horizontale (trafic piétonnier suspendu)
- Phase de circulation piétonnière omnidirectionnelle, autorisant les traversées diagonales
  
Ce projet vise à étudier et comparer différentes configurations de carrefours urbains (modèle Shibuya et modèle français traditionnel) afin d'analyser les dynamiques de circulation des véhicules et des piétons. La simulation est développée en Go et s'appuie sur le moteur de jeu 2D [Ebiten](https://ebiten.org/).


<div align="center" style="margin-top: 20px">
    <img src="assets/screen/screen_menu.png" width="80%" alt="Menu" />
</div>


<div align="center">
    <img src="assets/screen/screen_simu.png" width="100%" alt="Simu" />
</div>



## Principales Technologies utilisées

| Technologie | Description |
|-------------|-------------|
| **Forces Piétons/Voitures** | Répartition spatiale des différents agents, respectant une zone de distance |
| **Système de Grille des Piétons** | Algorithme d'optimisation du requêtage des piétons voisins |
| **A* Pathfinding** | Algorithme de recherche de chemin pour les déplacements des piétons |
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
- **Configuration flexible** : Paramètres de simulation ajustables (taux de spawn, etc.)

## Prérequis

Avant de commencer, il faut s'assurer d'avoir installé :

- [Go](https://golang.org/dl/) version 1.25.0 ou supérieure
- Un navigateur web moderne (pour le dashboard)

## Installation

**Cloner le dépôt** :

```bash
git clone https://gitlab.utc.fr/amisselm/shibuyacrossing_ai30.git
cd shibuyacrossing_ai30
```

Les dépendances seront automatiquement téléchargées au premier lancement.

## Lancer la simulation

### Démarrage de la simulation

Pour lancer la simulation principale :

```bash
go run cmd/simulation/main.go
```

Ou depuis la racine du projet :

```bash
go run gitlab.utc.fr/amisselm/shibuyacrossing_ai30/cmd/simulation
```

### Accès aux interfaces

Une fois la simulation lancée, vous aurez accès à :

- **Fenêtre de simulation** : Interface graphique de la simulation (Ebiten)
- **API REST** : `http://localhost:8080`
- **Dashboard web** : Accessible via le bouton "Dashboard" dans l'interface de simulation


## Structure du projet

```
shibuyacrossing_ai30/
├── cmd/
│   └── simulation/
│       └── main.go              # Point d'entrée de la simulation
├── internal/
│   ├── agents/                  # Agents (piétons, véhicules, feux)
│   ├── api/                     # Serveur API REST
│   ├── config/                  # Configuration de simulation
│   ├── environment/             # Zones de spawn et de répulsion
│   ├── game/                    # Logique principale du jeu
│   ├── handlers/                # Gestionnaires d'agents
│   ├── pathfinding/             # Calcul des trajets optimaux pour les piétons
│   ├── stats/                   # Calcul et comparaison des métriques
│   └── ui/                      # Composants d'interface
├── metrics_data/                # Données de métriques sauvegardées
├── dashboard.html               # Dashboard de visualisation
├── dashboard.css                # Styles du dashboard
└── README.md                    # Ce fichier

```


## 👥 Auteurs

Projet réalisé dans le cadre du cours AI30 à l'UTC.
- Pierrick Clause
- Meïssa H'mida
- Ahmed Misselmani
- Enzo Maros


---

# English version

**🇬🇧 English** · [🇫🇷 Version française](#-shibuya-crossing-simulator)

Shibuya Crossing Simulator is a multi-agent simulation reproducing the behaviour of an urban intersection inspired by the famous Shibuya pedestrian scramble in Tokyo.

The Shibuya Crossing traffic management system stands out for its sequential organisation in three distinct phases:
- Vertical vehicle traffic phase (pedestrian traffic suspended)
- Horizontal vehicle traffic phase (pedestrian traffic suspended)
- Omnidirectional pedestrian phase, allowing diagonal crossings

This project aims to study and compare different urban intersection layouts (the Shibuya model and the traditional French model) in order to analyse vehicle and pedestrian traffic dynamics. The simulation is written in Go and built on the 2D game engine [Ebiten](https://ebiten.org/).


<div align="center" style="margin-top: 20px">
    <img src="assets/screen/screen_menu.png" width="80%" alt="Menu" />
</div>


<div align="center">
    <img src="assets/screen/screen_simu.png" width="100%" alt="Simulation" />
</div>



## Main technologies

| Technology | Description |
|------------|-------------|
| **Pedestrian/Car forces** | Spatial distribution of the agents, keeping a safety distance between them |
| **Pedestrian grid system** | Optimisation algorithm for querying neighbouring pedestrians |
| **A* Pathfinding** | Path-finding algorithm for pedestrian movement |
| **Web server** | Display of dynamic charts (HTML, CSS, JavaScript, Chart.js, html2canvas, jsPDF) |
| **Ebitengine** | Simulation engine and overall game-loop architecture |

## Features

- **Real-time simulation**: Dynamic visualisation of vehicle and pedestrian movements
- **Two intersection types**: Comparison between the Shibuya system (all pedestrians cross at the same time) and the French system (traditional pedestrian crossings)
- **Detailed metrics**:
  - Average crossing time
  - Average speeds
  - Number of active and destroyed agents
  - Simulation duration
- **Web dashboard**: Comparative visualisation interface with charts and PDF export
- **REST API**: Real-time metrics retrieval
- **Flexible configuration**: Adjustable simulation parameters (spawn rate, etc.)

## Prerequisites

Before getting started, make sure you have installed:

- [Go](https://golang.org/dl/) version 1.25.0 or higher
- A modern web browser (for the dashboard)

## Installation

**Clone the repository**:

```bash
git clone https://gitlab.utc.fr/amisselm/shibuyacrossing_ai30.git
cd shibuyacrossing_ai30
```

Dependencies are downloaded automatically on the first run.

## Running the simulation

### Starting the simulation

To start the main simulation:

```bash
go run cmd/simulation/main.go
```

Or from the project root:

```bash
go run gitlab.utc.fr/amisselm/shibuyacrossing_ai30/cmd/simulation
```

### Accessing the interfaces

Once the simulation is running, you have access to:

- **Simulation window**: Graphical interface of the simulation (Ebiten)
- **REST API**: `http://localhost:8080`
- **Web dashboard**: Available through the "Dashboard" button in the simulation interface


## Project structure

```
shibuyacrossing_ai30/
├── cmd/
│   └── simulation/
│       └── main.go              # Simulation entry point
├── internal/
│   ├── agents/                  # Agents (pedestrians, vehicles, traffic lights)
│   ├── api/                     # REST API server
│   ├── config/                  # Simulation configuration
│   ├── environment/             # Spawn and repulsion areas
│   ├── game/                    # Main game logic
│   ├── handlers/                # Agent handlers
│   ├── pathfinding/             # Optimal path computation for pedestrians
│   ├── stats/                   # Metrics computation and comparison
│   └── ui/                      # Interface components
├── metrics_data/                # Saved metrics data
├── dashboard.html               # Visualisation dashboard
├── dashboard.css                # Dashboard styles
└── README.md                    # This file

```


## 👥 Authors

Project carried out as part of the AI30 course at UTC (Université de Technologie de Compiègne).
- Pierrick Clause
- Meïssa H'mida
- Ahmed Misselmani
- Enzo Maros
