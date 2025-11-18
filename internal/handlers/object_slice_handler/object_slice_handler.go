package objectslicehandler

import (
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// seul goroutine pouvant ajouter des objects au slice agents de l'environement (microservice)
func LancerObjectSliceHandler(ObjectSlice *[]Object) chan<- ObjectAddRequest {
	requestCh := make(chan ObjectAddRequest)

	go func() {

		for req := range requestCh {
			//fmt.Println("Adding new Object")
			*ObjectSlice = append(*ObjectSlice, req.Object)
			req.RespChan <- 1
			//fmt.Println("Added new Object ", X, " ", Y)
		}
	}()

	return requestCh
}
