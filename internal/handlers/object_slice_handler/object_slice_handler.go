package objectslicehandler

import (
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// microservice pour ajouter des objets à l'environnement (evite ecritures concurrent sur un meme slice par les agents)
func LancerObjectSliceHandler(ObjectSlice *[]Object) chan<- ObjectAddRequest {
	requestCh := make(chan ObjectAddRequest)

	go func() {

		for req := range requestCh {
			//ajoute l'objet au slice Objets de game
			*ObjectSlice = append(*ObjectSlice, req.Object)

			//repond à l'agent qu'il peut continuer/achever son execution
			req.RespChan <- 1
		}
	}()

	return requestCh
}
