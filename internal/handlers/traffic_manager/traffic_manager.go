package traffic_manager

import (
	"sync/atomic"
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

	PedsDiag []*trafficlight.TrafficLight // groupe pour la diagonale

	// 1 = Shibuya, 0 = French
	ShibuyaFlag  int32
	StepDuration time.Duration
	RedDuration  time.Duration

	started int32
}

// Start lance la boucle de contrôle infinie
func (tm *TrafficManager) Start() {
	// Si déjà démarré, on ne fait rien
	if !atomic.CompareAndSwapInt32(&tm.started, 0, 1) {
		return
	}

	go func() {
		for {
			if atomic.LoadInt32(&tm.ShibuyaFlag) == 1 {
				tm.runShibuyaCycle()
			} else {
				tm.runStandardCycle()
			}
		}
	}()
}

// SetMode change le mode de manière sécurisée
func (tm *TrafficManager) SetMode(isShibuya bool) {
	if isShibuya {
		atomic.StoreInt32(&tm.ShibuyaFlag, 1)
	} else {
		atomic.StoreInt32(&tm.ShibuyaFlag, 0)
	}
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
	tm.setGroupState(tm.PedsDiag, trafficlight.Red) // Diagonale toujours interdite
	time.Sleep(tm.StepDuration)
	// 2. Phase de Sécurité
	tm.setAllRed()
	time.Sleep(tm.RedDuration)

	// 3. Phase Est-Ouest VERT
	tm.setGroupState(tm.CarsNS, trafficlight.Red)
	tm.setGroupState(tm.PedsEW, trafficlight.Red)
	tm.setGroupState(tm.CarsEW, trafficlight.Green)
	tm.setGroupState(tm.PedsNS, trafficlight.Green)
	tm.setGroupState(tm.PedsDiag, trafficlight.Red)
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

	// 3. Tous les piétons Vert, toutes les voitures Rouge)
	tm.setGroupState(tm.CarsNS, trafficlight.Red)
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
	tm.setGroupState(tm.PedsNS, trafficlight.Red)
	tm.setGroupState(tm.PedsEW, trafficlight.Red)
	tm.setGroupState(tm.PedsDiag, trafficlight.Red)
}

// setAllPeds : utilitaire pour changer la couleur de tous les feux piétons d'un coup
func (tm *TrafficManager) setAllPeds(s trafficlight.State) {
	tm.setGroupState(tm.PedsNS, s)
	tm.setGroupState(tm.PedsEW, s)
	tm.setGroupState(tm.PedsDiag, s)
}
