package carstopping

import (
	trafficlight "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/traffic_light"
	cars "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/vehicule"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

func LancerCarStopper(chID <-chan ObjectID, chH chan trafficlight.State, chV chan trafficlight.State, env *[]Object) {
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
