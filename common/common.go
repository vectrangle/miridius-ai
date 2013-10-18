package common

import (
    "github.com/zond/stockholm-ai/state"
)

// integer Min function instead of using math.Min
func Min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

/*
   Delay == 0                   - node is claimed
   0 < Delay < DELAY_NO_UNITS   - some units will arrive in Delay turns
   Delay = DELAY_NO_UNITS       - no units approaching
*/
const DELAY_NO_UNITS = 1000

type UnitCounts struct {
    Units      int
    EnemyUnits int
    Delay      int
    // EnemyDelay int
    Adjacent bool
}

func NewUnitCounts() *UnitCounts {
    return &UnitCounts{0, 0, DELAY_NO_UNITS, false}
}

func (self *UnitCounts) add(u *UnitCounts) {
    self.Units += u.Units
    self.EnemyUnits += u.EnemyUnits
    self.Delay = Min(self.Delay, u.Delay)
    // self.EnemyDelay = math.Min(self.EnemyDelay, u.EnemyDelay)
    self.Adjacent = self.Adjacent || u.Adjacent
}

func CountEdgeUnits(me state.PlayerId, edge *state.Edge) (u *UnitCounts) {
    u = NewUnitCounts()
    for index, unitMap := range edge.Units {
        for player, numUnits := range unitMap {
            if numUnits > 0 {
                if player == me {
                    u.Units += numUnits
                } else {
                    u.EnemyUnits += numUnits
                }
                u.Delay = Min(u.Delay, len(edge.Units)-index)
            }
        }
    }
    return
}

func CountNodeUnits(me state.PlayerId, node *state.Node) (u *UnitCounts) {
    u = NewUnitCounts()
    for player, numUnits := range node.Units {
        if player == me {
            u.Units += numUnits
        } else {
            u.EnemyUnits += numUnits
        }
    }
    if u.Units > 0 {
        u.Delay = 0
        u.Adjacent = true
    }
    if u.EnemyUnits > 0 {
        u.Delay = 0
    }
    return
}

func CountAllUnits(me state.PlayerId, s *state.State) (result map[state.NodeId]*UnitCounts) {
    // allocate map
    result = make(map[state.NodeId]*UnitCounts, len(s.Nodes))
    for nodeId := range s.Nodes {
        result[nodeId] = NewUnitCounts()
    }
    // calculate result
    for nodeId, node := range s.Nodes {
        result[nodeId].add(CountNodeUnits(me, node))
        for _, edge := range node.Edges {
            result[edge.Dst].add(CountEdgeUnits(me, &edge))
            if result[edge.Src].Units > 0 {
                result[edge.Dst].Adjacent = true
            }
        }
    }
    return
}
