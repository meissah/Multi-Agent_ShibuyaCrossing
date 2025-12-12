package perceptcharacteristicshandler

import (
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// microservice pour l'ajout d'un agent dans le slice du jeu, evite des ecritures concurrentes entre spawn cars et spawn pietons
func LancerCharacteristicsHandler(gObjets *[]Object, pCChn chan<- ReceiveCharacRequest) (chan<- int, <-chan []PhysicalCharacteristics) {
	ReceiveRequestCh := make(chan []PhysicalCharacteristics)
	updateCharacsCh := make(chan int)

	var currentCharacs []PhysicalCharacteristics
	var currentCharac *PhysicalCharacteristics

	go func() {
		for {
			select {
			case <-updateCharacsCh:
				currentCharacs = make([]PhysicalCharacteristics, 0)
				for _, v := range *gObjets {
					receiveCh := make(chan *PhysicalCharacteristics)
					pCChn <- ReceiveCharacRequest{
						Object:        v,
						ReceiveCharac: receiveCh,
					}

					currentCharac = <-receiveCh
					currentCharacs = append(currentCharacs, *currentCharac)
				}
			case ReceiveRequestCh <- currentCharacs:
			}
		}
	}()

	return updateCharacsCh, ReceiveRequestCh
}
