package agentslicehandler

import (
	"sync"

	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

type Request struct {
	Agent    Agent
	SyncChan chan int
}

// seul goroutine pouvant ajouter des agents au slice agents de l'environement (microservice)
func LancerAgentSliceHandler(AgentSlice *[]Agent, syncChans *sync.Map) chan<- Request {
	requestCh := make(chan Request)

	go func() {

		for req := range requestCh {
			//fmt.Println("Adding new Agent")
			syncChans.Store(req.Agent.AgtID(), req.SyncChan)
			*AgentSlice = append(*AgentSlice, req.Agent)
			//fmt.Println("Added new Agent")

			// go func(agt Agent) {
			// 	step := 0
			// 	for {
			// 		step++
			// 		c, _ := syncChans.Load(agt.AgtID())
			// 		c.(chan int) <- step
			// 		<-c.(chan int)
			// 	}
			// }(req.Agent)

		}
	}()

	return requestCh
}
