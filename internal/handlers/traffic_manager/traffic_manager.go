package traffic_manager

import (
	"time"

	trafficlight "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/traffic_light"
)

type TrafficManager struct {
	// Groupes de feux VOITURES
	CarsNS []*trafficlight.TrafficLight // Nord-Sud
	CarsEW []*trafficlight.TrafficLight // Est-Ouest

	// Groupes de feux PIETONS
	PedsNS []*trafficlight.TrafficLight // Nord-Sud
	PedsEW []*trafficlight.TrafficLight // Est-Ouest

	// Configuration
	IsShibuyaMode bool          // true = mode shibuya, false = mode french
	StepDuration  time.Duration // Durée d'un feu vert
	RedDuration   time.Duration // Durée de sécurité (tout rouge entre deux changements)
}

// Start lance la boucle de contrôle infinie
func (tm *TrafficManager) Start() {
	go func() {
		for {
			if tm.IsShibuyaMode {
				tm.runShibuyaCycle()
			} else {
				tm.runStandardCycle()
			}
		}
	}()
}

// setGroupState : utilitaire pour changer la couleur de toute une liste de feux d'un coup
func (tm *TrafficManager) setGroupState(lights []*trafficlight.TrafficLight, state trafficlight.State) {
	for _, tl := range lights {
		tl.SetState(state)
	}
}

// French
// Cycle : Voitures NS + Piétons NS  ->  Tout Rouge  ->  Voitures EO + Piétons EO
func (tm *TrafficManager) runStandardCycle() {
	// 1. Phase Nord-Sud VERT
	tm.setGroupState(tm.CarsNS, trafficlight.Green)
	tm.setGroupState(tm.PedsEW, trafficlight.Green)
	tm.setGroupState(tm.CarsEW, trafficlight.Red)
	tm.setGroupState(tm.PedsNS, trafficlight.Red)
	time.Sleep(tm.StepDuration)

	// 2. Phase de Sécurité
	tm.setAllRed()
	time.Sleep(tm.RedDuration)

	// 3. Phase Est-Ouest VERT
	tm.setGroupState(tm.CarsNS, trafficlight.Red)
	tm.setGroupState(tm.PedsEW, trafficlight.Red)
	tm.setGroupState(tm.CarsEW, trafficlight.Green)
	tm.setGroupState(tm.PedsNS, trafficlight.Green)
	time.Sleep(tm.StepDuration)

	// 4. Phase de Sécurité
	tm.setAllRed()
	time.Sleep(tm.RedDuration)
}

// shibuya
func (tm *TrafficManager) runShibuyaCycle() {
	// 1. Voitures Nord-Sud seulement
	tm.setGroupState(tm.CarsNS, trafficlight.Green)
	tm.setGroupState(tm.CarsEW, trafficlight.Red)
	tm.setAllPeds(trafficlight.Red) // Piétons attendent
	time.Sleep(tm.StepDuration)

	// Phase de Sécurité
	tm.setAllRed()
	time.Sleep(tm.RedDuration)

	// 2. Voitures Est-Ouest seulement
	tm.setGroupState(tm.CarsNS, trafficlight.Red)
	tm.setGroupState(tm.CarsEW, trafficlight.Green)
	tm.setAllPeds(trafficlight.Red) // Piétons attendent
	time.Sleep(tm.StepDuration)

	// Phase de Sécurité
	tm.setAllRed()
	time.Sleep(tm.RedDuration)

	// 3. Tous les piétons Vert, toutes les voitures Rouge)	tm.setGroupState(tm.CarsNS, trafficlight.Red)
	tm.setGroupState(tm.CarsEW, trafficlight.Red)
	tm.setAllPeds(trafficlight.Green) // les piétons traversent
	time.Sleep(tm.StepDuration)

	// Sécurité avant de reprendre
	tm.setAllRed()
	time.Sleep(tm.RedDuration)
}

func (tm *TrafficManager) setAllRed() {
	tm.setGroupState(tm.CarsNS, trafficlight.Red)
	tm.setGroupState(tm.CarsEW, trafficlight.Red)
	tm.setAllPeds(trafficlight.Red)
}

// setAllPeds : utilitaire pour changer la couleur de tous les feux piétons d'un coup
func (tm *TrafficManager) setAllPeds(s trafficlight.State) {
	tm.setGroupState(tm.PedsNS, s)
	tm.setGroupState(tm.PedsEW, s)
}
