package agentslicehandler

import (
	"sync"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

type Request struct {
	Agent    Agent
	SyncChan chan int
}

// microservice pour l'ajout d'un agent dans le slice du jeu, evite des ecritures concurrentes entre spawn cars et spawn pietons
func LancerAgentSliceHandler(AgentSlice *[]Agent, syncChans *sync.Map) chan<- Request {
	requestCh := make(chan Request)

	go func() {

		for req := range requestCh {
			//stocke le channel de synchronisation de l'agent dans l'environnement game
			syncChans.Store(req.Agent.AgtID(), req.SyncChan)

			//ajoute l'agent au slice agents de l'envrionnement
			*AgentSlice = append(*AgentSlice, req.Agent)

		}
	}()

	return requestCh
}
