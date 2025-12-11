package agentcounthandler

import (
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// microservice pour avoir le prochain objectID disponible, evite des ecritures concurrentes en car et pieton spawn
func LancerAgentCountHandler(objectCount *ObjectID) <-chan ObjectID {
	out := make(chan ObjectID, 1) // toujours sur son channel le prochain ID d'objet disponible

	go func() {
		for {
			//assure continuellement la disponibilité d'un agentID valable dans le channel
			out <- *objectCount
			*objectCount++
		}
	}()

	return out
}
