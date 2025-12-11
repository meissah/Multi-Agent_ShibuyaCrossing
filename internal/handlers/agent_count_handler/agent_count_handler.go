package agentcounthandler

import (
	. "gitlab.utc.fr/amisselm/shibuyacrossing_ai30/internal/utils"
)

// microservice pour avoir le prochain agentID disponible, evite des ecritures concurrentes en car et pieton spawn
func LancerAgentCountHandler(agentCount *AgentID) <-chan AgentID {
	out := make(chan AgentID, 1) // toujours sur son channel le prochain ID d'agent disponible

	go func() {
		for {
			//assure continuellement la disponibilité d'un agentID valable dans le channel
			out <- *agentCount
			*agentCount++
		}
	}()

	return out
}
