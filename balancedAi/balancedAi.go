// balancedAi by Miridius
package balancedAi

import (
    common "github.com/zond/stockholm-ai/common"
    state "github.com/zond/stockholm-ai/state"
)

/*
Balanced AI
Basically just aims to multiply as much as possible
v1 algorithm:
1. For each node i in s:
    a. i.Attraction = how much growth I would gain by sending 1 soldier there
2. For each node j in s where I have units:
    a. For each node k, attraction to k = k.Attraction / (distance from j to k)
    b. attraction of each edge connected to j is the sum of the attractions of nodes whos path start with that edge
    c. attraction of not moving = j.Attraction
    d. leave 1 unit to hold the node, and divide remaining units amongst all edges proportionally based on attraction ratios

Known issues:
1. Soldiers currently on edges are not considered in calculations, which causes the AI to send out units more often than really necessary.
2. Playing multiple balanced AIs against each other can result in deadlock
*/
type BalancedAi1 struct{}

/*
Orders will analyze all nodes in s and return orders for each one
*/
func (self BalancedAi1) Orders(logger common.Logger, me state.PlayerId, s *state.State) (result state.Orders) {

    logger.Printf("BalancedAi1 calculating orders for player: %v", me)

    var attraction, totalAttraction float64
    var edge state.NodeId
    // Calculate base attraction for all nodes
    attractions := make(map[state.NodeId]float64, len(s.Nodes)+1)
    for _, node := range s.Nodes {
        if node.Units[me] < 1 {
            attraction = 1
        } else {
            attraction = 0
        }
        attraction = attraction + (0.2 * float64(node.Units[me]) / float64(node.Size))

        attractions[node.Id] = attraction
    }

    // For each node in s
    for _, node := range s.Nodes {
        // If I have units there (after leaving 1 behind to defend)
        if units := node.Units[me] - 1; units > 0 {
            // Check my attraction to all other nodes and keep an attraction sum for each starting edge.
            edgeAttractions := make(map[state.NodeId]float64, len(node.Edges)+1)
            totalAttraction = 0
            for _, destNode := range s.Nodes {
                path := s.Path(node.Id, destNode.Id, nil)
                if len(path) > 0 {
                    edge = path[0]
                    attraction = attractions[destNode.Id] / float64(len(path))
                } else {
                    edge = node.Id
                    attraction = attractions[destNode.Id]
                }
                edgeAttractions[edge] = edgeAttractions[edge] + attraction
                totalAttraction = totalAttraction + attraction
            }
            // go through all edges and send units accordingly
            for edgeId, att := range edgeAttractions {
                // in case of rounding errors or some other hiccup, make sure current edge's attraction <= total
                if att > totalAttraction {
                    totalAttraction = att
                }
                sendUnits := int(float64(units) * att / totalAttraction)
                units = units - sendUnits
                totalAttraction = totalAttraction - att
                result = append(result, state.Order{
                    Src:   node.Id,
                    Dst:   edgeId,
                    Units: sendUnits,
                })
            }
        }
    }
    return
}
