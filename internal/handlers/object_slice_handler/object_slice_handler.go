package objectslicehandler

import (
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// microservice pour ajouter des objets à l'environnement (evite ecritures concurrent sur un meme slice par les agents)
// resp upadte du slice object
func LancerObjectSliceHandler(ObjectSlice *[]Object) chan<- ObjectAddRequest {
	requestCh := make(chan ObjectAddRequest, 10000)

	go func() {

		for req := range requestCh {
			//ajoute l'objet au slice Objets de game
			*ObjectSlice = append(*ObjectSlice, req.Object)

			//repond à l'agent qu'il peut continuer/achever son execution ( append terminé  ) -> permet dans start.go de s'assurer que l'agent n'avance pas avant que son objet soit bien ajouté dans le slice des objets de l'environnement
			req.RespChan <- 1
		}
	}()

	return requestCh
}
