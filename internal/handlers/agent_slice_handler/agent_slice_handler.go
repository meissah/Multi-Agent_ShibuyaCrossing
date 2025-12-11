package agentslicehandler

import (
	"sync"

	"sync/atomic"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

type AddAgtRequest struct {
	Agent    Agent
	SyncChan chan int
}

// microservice pour l'ajout d'un agent dans le slice du jeu, evite des ecritures concurrentes entre spawn cars et spawn pietons
func LancerAgentSliceHandler(AgentSlice *atomic.Pointer[[]Agent], ObjectSlice *atomic.Pointer[[]Object], syncChans *sync.Map) (chan<- AddAgtRequest, chan<- Agent) {
	AddRequestCh := make(chan AddAgtRequest)
	DeleteAgtsCh := make(chan Agent)

	go func() {
		for {
			select {
			case req := <-AddRequestCh:
				//stocke le channel de synchronisation de l'agent dans l'environnement game
				syncChans.Store(req.Agent.AgentID(), req.SyncChan)

				var newAgtSlice []Agent
				var newObjtSlice []Object

				//Creation nouveau slice environnement de Agents
				oldAgtSlice := AgentSlice.Load()
				if oldAgtSlice == nil {
					newAgtSlice = make([]Agent, 0)
					newAgtSlice = append(newAgtSlice, req.Agent)
				} else {
					newAgtSlice = make([]Agent, len(*oldAgtSlice)+1)
					copy(newAgtSlice, *oldAgtSlice)
					newAgtSlice = append(newAgtSlice, req.Agent)
				}

				//Creation nouveau slice environnement de Objets
				oldObjtSlice := ObjectSlice.Load()
				if oldObjtSlice == nil {
					newObjtSlice := make([]Object, 0)
					newObjtSlice = append(newObjtSlice, req.Agent)
				} else {
					newObjtSlice := make([]Object, len(*oldObjtSlice)+1)
					copy(newObjtSlice, *oldObjtSlice)
					newObjtSlice = append(newObjtSlice, req.Agent)
				}

				//ajoute l'agent au slice agents de l'environnement
				AgentSlice.Store(&newAgtSlice)
				//ajoute l'agent au slice agents de l'environnement
				ObjectSlice.Store(&newObjtSlice)
			case Agt := <-DeleteAgtsCh:
				oldAgtSlice := AgentSlice.Load()
				for i, v := range *oldAgtSlice {
					if v.AgentID() == Agt.AgentID() {
						newAgtSlice := make([]Agent, len(*oldAgtSlice)-1)
						copy(newAgtSlice[:i], (*oldAgtSlice)[:i])
						copy(newAgtSlice[i:], (*oldAgtSlice)[i+1:])
						//ajoute l'agent au slice agents de l'environnement
						AgentSlice.Store(&newAgtSlice)
						v.EndChan() <- 1
						break
					}
				}
				oldObjtSlice := ObjectSlice.Load()
				for i, v := range *oldObjtSlice {
					if v.ObjectID() == ObjectID(Agt.AgentID()) {
						newObjtSlice := make([]Object, len(*oldObjtSlice)-1)
						copy(newObjtSlice[:i], (*oldObjtSlice)[:i])
						copy(newObjtSlice[i:], (*oldObjtSlice)[i+1:])
						//ajoute l'agent au slice agents de l'environnement
						ObjectSlice.Store(&newObjtSlice)
						break
					}
				}
			}
		}
	}()

	return AddRequestCh, DeleteAgtsCh
}
