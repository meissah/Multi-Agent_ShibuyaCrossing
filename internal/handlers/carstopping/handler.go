package carstopping

import (
	"time"

	trafficlight "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/traffic_light"
	cars "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/grid_handler"
	"gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/messaging"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

func LancerCarStopperAtRedLight(chID <-chan ObjectID, chH chan trafficlight.State, chV chan trafficlight.State, env *[]Object) {
	var RedLightCarStopH, RedLightCarStopV []*cars.Car

	carH1 := cars.NewCar(<-chID, 330, 510, North, 0, false, nil)
	carH2 := cars.NewCar(<-chID, 395, 250, South, 0, false, nil)
	carH3 := cars.NewCar(<-chID, 360, 510, North, 1, false, nil)
	carH4 := cars.NewCar(<-chID, 425, 250, South, 1, false, nil)
	RedLightCarStopH = append(RedLightCarStopH, carH1, carH2, carH3, carH4)
	for _, car := range RedLightCarStopH {
		car.State = "inative"
		car.IsPlaceholder = true
	}

	//lane gauche
	carV1 := cars.NewCar(<-chID, 245, 325, East, 0, false, nil)
	carV2 := cars.NewCar(<-chID, 510, 430, West, 1, false, nil)
	//lane droite
	carV3 := cars.NewCar(<-chID, 245, 355, East, 1, false, nil)
	carV4 := cars.NewCar(<-chID, 510, 400, West, 0, false, nil)
	RedLightCarStopV = append(RedLightCarStopV, carV1, carV2, carV3, carV4)
	for _, car := range RedLightCarStopV {
		car.State = "inative"
		car.IsPlaceholder = true
	}

	*env = append(*env, carH1, carH2, carH3, carH4, carV1, carV2, carV3, carV4)

	go func() {
		for {
			select {
			case stateH := <-chH:
				//fmt.Println("RECEIVED FROM CHANNEL H")
				if stateH == trafficlight.Red {
					//fmt.Println("Just RECEIVED A STOP HORIZONTAL")
					for _, car := range RedLightCarStopH {
						car.State = "active"
					}
				} else {
					//fmt.Println("Just RECEIVED A GO HORIZONTAL")
					for _, car := range RedLightCarStopH {
						car.State = "inative"
					}
				}

			case stateV := <-chV:
				//fmt.Println("RECEIVED FROM CHANNEL V")
				if stateV == trafficlight.Red {
					//fmt.Println("Just RECEIVED A STOP VERTICAL")
					for _, car := range RedLightCarStopV {
						car.State = "active"
					}
				} else {
					//fmt.Println("Just RECEIVED A GO VERTICAL")
					for _, car := range RedLightCarStopV {
						car.State = "inative"
					}
				}
			}
		}
	}()
	return
}

func LancerCarStopperForPietons(chID <-chan ObjectID, pedGrid *grid_handler.PedestrianGrid, env *[]Object) {

	var PedCarStopW, PedCarStopE, PedCarStopN, PedCarStopS *cars.Car

	PedCarStopW = cars.NewCar(<-chID, 285, 425, West, 1, false, nil)
	PedCarStopE = cars.NewCar(<-chID, 470, 330, East, 0, false, nil)
	PedCarStopN = cars.NewCar(<-chID, 330, 290, North, 0, false, nil)
	PedCarStopS = cars.NewCar(<-chID, 425, 470, South, 1, false, nil)

	PedCarStopW.State = "inative"
	PedCarStopW.IsPlaceholder = true
	PedCarStopE.State = "inative"
	PedCarStopE.IsPlaceholder = true
	PedCarStopN.State = "inative"
	PedCarStopN.IsPlaceholder = true
	PedCarStopS.State = "inative"
	PedCarStopS.IsPlaceholder = true

	// PedCarStopW.State = "active"
	// PedCarStopE.State = "active"
	// PedCarStopN.State = "active"
	// PedCarStopS.State = "active"

	*env = append(*env, PedCarStopW, PedCarStopE, PedCarStopN, PedCarStopS)

	gridChan := pedGrid.RequestChan()

	PedWch, PedEch, PedNch, PedSch := make(chan []Object), make(chan []Object), make(chan []Object), make(chan []Object)

	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)

		for range ticker.C {

			gridChan <- messaging.GetInRectRequest{
				TopLeft:     *NewPoint(255, 400),
				BottomRight: *NewPoint(300, 428),
				ReplyChan:   PedWch,
			}
			PedW := <-PedWch

			if len(PedW) > 0 {
				PedCarStopW.State = "active"
			} else {
				PedCarStopW.State = "inative"
			}
			gridChan <- messaging.GetInRectRequest{
				TopLeft:     *NewPoint(452, 310),
				BottomRight: *NewPoint(506, 340),
				ReplyChan:   PedEch,
			}

			PedE := <-PedEch
			if len(PedE) > 0 {
				PedCarStopE.State = "active"
			} else {
				PedCarStopE.State = "inative"
			}
			gridChan <- messaging.GetInRectRequest{
				TopLeft:     *NewPoint(315, 250),
				BottomRight: *NewPoint(353, 303),
				ReplyChan:   PedNch,
			}
			PedN := <-PedNch
			if len(PedN) > 0 {
				PedCarStopN.State = "active"
			} else {
				PedCarStopN.State = "inative"
			}
			gridChan <- messaging.GetInRectRequest{
				TopLeft:     *NewPoint(406, 453),
				BottomRight: *NewPoint(427, 504),
				ReplyChan:   PedSch,
			}
			PedS := <-PedSch
			if len(PedS) > 0 {
				PedCarStopS.State = "active"
			} else {
				PedCarStopS.State = "inative"
			}

		}

		defer ticker.Stop()

	}()
	return
}
