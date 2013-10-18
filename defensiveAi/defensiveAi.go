// defensiveAi by Miridius
package defensiveAi

import (
    common "github.com/miridius/ai/common"
    stockholmCommon "github.com/zond/stockholm-ai/common"
    state "github.com/zond/stockholm-ai/state"
)

/*
Defensive AI
- Sends 1 unit to adjacent unclaimed nodes only
- Maintains node.size/2 units at each node
- Attacks with any extras

v1 Algorithm:
1. For each node that has >1 unit
    a. For each edge:
        i. ensure that we still have guys available
        ii. if edge.Dst is unclaimed and has no units (friendly or enemy) on their way will beat me there, send 1 guy
    b. If units > node.size/2, send (units - node.size/2) units towards nearest enemy or unclaimed node
*/
type DefensiveAi1 struct{}

/*
Orders will analyze all nodes in s and return orders for each one
*/
func (self DefensiveAi1) Orders(logger stockholmCommon.Logger, me state.PlayerId, s *state.State) (result state.Orders) {

    logger.Printf("DefensiveAi1 calculating orders for player: %v", me)

    // gather data
    unitCounts := common.CountAllUnits(me, s)

    // 1. For each node that has >1 unit
    for nodeId, node := range s.Nodes {
        if units := node.Units[me]; units > 1 {
            // a. For each edge:
            for _, edge := range node.Edges {
                // i. ensure that we still have guys available
                if units <= 1 {
                    break
                }
                // ii. if edge.Dst is unclaimed and has no units (friendly or enemy) on their way that will beat me there, send 1 guy
                if unitCounts[edge.Dst].Delay > len(edge.Units) {
                    result = append(result, state.Order{
                        Src:   edge.Src,
                        Dst:   edge.Dst,
                        Units: 1,
                    })
                    units--
                }
            }
            // b. If units > node.size/2, send up to (units - node.size/2) available units towards nearest/least defended enemy node
            // if we
            available := unitCounts[nodeId].Units - unitCounts[nodeId].EnemyUnits
            if sendUnits := common.Min(available, units-node.Size/2); sendUnits > 0 {
                var cheapest float64 = -1
                cheapestEdge := nodeId
                for dst, dstUnits := range unitCounts {
                    if dst == nodeId {
                        continue
                    }
                    if dstUnits.Adjacent && dstUnits.EnemyUnits > 0 {
                        path := s.Path(node.Id, dst, nil)
                        // how many men do I lose to capture this node
                        thisCost := float64(dstUnits.EnemyUnits-dstUnits.Units) * (1 + float64(len(path))*0.2)
                        if cheapest == -1 || thisCost < cheapest {
                            cheapest = thisCost
                            cheapestEdge = path[0]
                        }
                    }
                }
                if cheapest != -1 {
                    //if we have enough units to capture the node, send that many
                    if int(cheapest) < common.Min(units-1, available) && int(cheapest) > sendUnits {
                        sendUnits = int(cheapest)
                    }
                    result = append(result, state.Order{
                        Src:   node.Id,
                        Dst:   cheapestEdge,
                        Units: sendUnits,
                    })
                }
            }
        }
    }
    // all done, return the orders list
    return
}
