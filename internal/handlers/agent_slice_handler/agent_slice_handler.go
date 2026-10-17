package agentslicehandler

import (
	"sync"

	. "github.com/meissah/Multi-Agent_ShibuyaCrossing/internal/utils"
)

type AddAgtRequest struct {
	Agent    Agent
	SyncChan chan int
}

// microservice pour l'ajout d'un agent dans le slice du jeu, evite des ecritures concurrentes entre spawn cars et spawn pietons
func LancerAgentSliceHandler(AgentSlice *[]Agent, ObjectSlice *[]Object, syncChans *sync.Map) (chan<- AddAgtRequest, chan<- Agent) {
	AddRequestCh := make(chan AddAgtRequest)
	DeleteAgtsCh := make(chan Agent)

	go func() {
		for {
			select {
			case req := <-AddRequestCh:
				//stocke le channel de synchronisation de l'agent dans l'environnement game
				syncChans.Store(req.Agent.AgentID(), req.SyncChan)

				//Creation nouveau slice environnement de Agents
				oldAgtSlice := *AgentSlice
				newAgtSlice := make([]Agent, 0, len(oldAgtSlice)+1)
				newAgtSlice = append(newAgtSlice, oldAgtSlice...)
				newAgtSlice = append(newAgtSlice, req.Agent)

				//ajoute l'agent au slice agents de l'environnement
				*AgentSlice = newAgtSlice

				//Creation nouveau slice environnement de Objets
				oldObjtSlice := *ObjectSlice
				newObjtSlice := make([]Object, 0, len(oldObjtSlice)+1)
				newObjtSlice = append(newObjtSlice, oldObjtSlice...)
				newObjtSlice = append(newObjtSlice, req.Agent)

				//ajoute l'agent au slice agents de l'environnement
				*ObjectSlice = newObjtSlice

			case Agt := <-DeleteAgtsCh:
				// remove matching Object from ObjectSlice
				oldObjtSlice := *ObjectSlice
				for i, v := range oldObjtSlice {
					if v.ObjectID() == ObjectID(Agt.AgentID()) {
						// allocate a new slice with capacity len-1 and length 0 (new backing array)
						newObjtSlice := make([]Object, 0, len(oldObjtSlice)-1)
						// append elements before i, then elements after i
						newObjtSlice = append(newObjtSlice, oldObjtSlice[:i]...)
						newObjtSlice = append(newObjtSlice, oldObjtSlice[i+1:]...)
						// replace
						*ObjectSlice = newObjtSlice
						break
					}
				}

				// remove matching Agent from AgentSlice
				oldAgtSlice := *AgentSlice
				for i, v := range oldAgtSlice {
					if v.AgentID() == Agt.AgentID() {
						// new backing array with capacity len-1
						newAgtSlice := make([]Agent, 0, len(oldAgtSlice)-1)
						newAgtSlice = append(newAgtSlice, oldAgtSlice[:i]...)
						newAgtSlice = append(newAgtSlice, oldAgtSlice[i+1:]...)
						// replace
						*AgentSlice = newAgtSlice

						// notify the agent (keep using v from the old slice)
						v.EndChan() <- 1
						break
					}
				}
				syncChans.Delete(Agt.AgentID())

			}
		}
	}()

	return AddRequestCh, DeleteAgtsCh
}
