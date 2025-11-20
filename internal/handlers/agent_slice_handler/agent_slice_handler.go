package agentslicehandler

import (
	"sync"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

type AddAgtRequest struct {
	Agent    Agent
	SyncChan chan int
}

// microservice pour l'ajout d'un agent dans le slice du jeu, evite des ecritures concurrentes entre spawn cars et spawn pietons
func LancerAgentSliceHandler(AgentSlice *[]Agent, syncChans *sync.Map) (chan<- AddAgtRequest, chan []Agent, chan Agent) {
	AddRequestCh := make(chan AddAgtRequest)
	ReceiveAgtsCh := make(chan []Agent)
	DeleteAgtsCh := make(chan Agent)

	go func() {
		for {
			select {
			case req := <-AddRequestCh:
				//stocke le channel de synchronisation de l'agent dans l'environnement game
				syncChans.Store(req.Agent.AgtID(), req.SyncChan)

				//ajoute l'agent au slice agents de l'environnement
				*AgentSlice = append(*AgentSlice, req.Agent)
			case ReceiveAgtsCh <- func() []Agent {
				original := *AgentSlice
				copySlice := make([]Agent, len(original))
				copy(copySlice, original)
				return copySlice
			}():
			case Agt := <-DeleteAgtsCh:
				for i, v := range *AgentSlice {
					if v.AgtID() == Agt.AgtID() {
						*AgentSlice = append((*AgentSlice)[:i], (*AgentSlice)[i+1:]...)
						break
					}
				}
				Agt.EndChan() <- 1
			}
		}
	}()

	return AddRequestCh, ReceiveAgtsCh, DeleteAgtsCh
}
