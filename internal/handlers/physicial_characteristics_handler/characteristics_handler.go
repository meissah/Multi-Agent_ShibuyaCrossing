package physicialcharacteristicshandler

import (
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// microservice pour l'ajout d'un agent dans le slice du jeu, evite des ecritures concurrentes entre spawn cars et spawn pietons
func LancerCharacteristicsHandler() (chan<- ChangeCharacRequest, chan<- ReceiveCharacRequest) {
	AddRequestCh := make(chan ChangeCharacRequest)
	ReceiveRequestCh := make(chan ReceiveCharacRequest)

	go func() {
		for {
			select {
			case req := <-AddRequestCh:
				req.Object.ChangePhysicalCharac(req.NewCharacteristics)
			case req := <-ReceiveRequestCh:
				ptrCharac := req.Object.PhysicalCharacteristics()
				req.ReceiveCharac <- ptrCharac
			}
		}
	}()

	return AddRequestCh, ReceiveRequestCh
}
