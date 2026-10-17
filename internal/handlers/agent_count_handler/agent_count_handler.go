package agentcounthandler

import (
	. "github.com/meissah/Multi-Agent_ShibuyaCrossing/internal/utils"
)

// microservice pour avoir le prochain objectID disponible, evite des ecritures concurrentes en car et pieton spawn
func LancerAgentCountHandler(objectCount *ObjectID, objectSlice *[]Object) <-chan ObjectID {
	out := make(chan ObjectID, 1) // toujours sur son channel le prochain ID d'objet disponible

	*objectCount = ObjectID(len(*objectSlice))

	go func() {
		for {
			//assure continuellement la disponibilité d'un agentID valable dans le channel
			out <- *objectCount
			*objectCount++
		}
	}()

	return out
}
