package carstopping

import (
	trafficlight "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/agents/traffic_light"
	agentslicehandler "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/handlers/agent_slice_handler"
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

func LancerCarStopper(chH chan trafficlight.State, chV chan trafficlight.State, env *[]Object, chAddA chan<- agentslicehandler.AddAgtRequest, chDelA chan<- Agent) {
	// go func() {
	// 	select {
	// 	case stateH := <-chH:
	// 		if stateH == trafficlight.Red {
	// 			fmt.Println("Just RECEIVED A STOP HORIZONTAL")
	// 		} else {
	// 			fmt.Println("Just RECEIVED A GO HORIZONTAL")
	// 		}

	// 	case stateV := <-chV:
	// 		if stateV == trafficlight.Red {
	// 			fmt.Println("Just RECEIVED A STOP VERTICAL")
	// 		} else {
	// 			fmt.Println("Just RECEIVED A GO VERTICAL")
	// 		}
	// 	}
	// }()
	return
}
