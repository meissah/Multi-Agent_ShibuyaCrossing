package agentcounthandler

import (
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// seul goroutine ayant un acces direct à AgentCount de l'environement (microservice)
func LancerAgentCountHandler(agentCount *AgentID) <-chan AgentID {
	out := make(chan AgentID, 1)

	go func() {
		for {
			//fmt.Println("Sending new AgentID")
			out <- *agentCount
			*agentCount++
		}
	}()

	return out
}
