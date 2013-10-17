// aggressiveAi by Miridius
package aggressiveAi

import (
	common "github.com/zond/stockholm-ai/common"
	state "github.com/zond/stockholm-ai/state"
	"sort"
)

/*
Aggressive AI
Leaves 1 unit to hold each node and sends all the rest either out to claim new nodes or into battle

v1 Algorithm:
0. Calculate available soldiers = every soldier on the map (including edges) excluding one guy left to defend on each friendly node
1. For each unclaimed or enemy node:
	a. Find closest available soldier, add path of that soldier to a queue
2. Sort that queue by shortest path first
3. For each path in the queue
	a. Pop the path
	b. If that soldier is still available, create an order for it (if on a node), and mark it as unavailable
	c. If the soldier is not available any more, find next closest available soldier and add that path to queue, then re-sort the queue
5. Send any remaining available soldiers towards closest enemy node (or friendly nodes with enemy units)

v1.1:
 - fixed 2 different memory leaks

Ideas for improvements:
 - count enemy units on edges as belonging to the node they are going to land on, so that we defend nodes with incoming soldiers instead of only leaving 1 guy there
 - only leave enough units to defend a node as are needed to kill all enemy soldiers, send the rest into battle elsewhere

*/
type AggressiveAi1 struct{}

// describes a packet of soldiers who are not already occupied with a task
type availableSoldiers struct {
	num int 	// number of soldiers
	delay int 	// how long until they arrive at a node (0 means already there)
}

// describes a single path in the path queue
type Path struct {
	src, dst state.NodeId
	delay, length int
}

// define a sortable queue of Paths
type PathQueue []Path
func (s PathQueue) Len() int		{ return len(s) }
func (s PathQueue) Swap(i, j int)	{ s[i], s[j] = s[j], s[i] }

// create a type and method to sort a PathQueue by their length
type ByDist struct { PathQueue }
func (s ByDist) Less(i, j int) bool	{ return s.PathQueue[i].length < s.PathQueue[j].length }

/*
Orders will analyze all nodes in s and return orders for each one
*/
func (self AggressiveAi1) Orders(logger common.Logger, me state.PlayerId, s *state.State) (result state.Orders) {
	
	logger.Printf("AggressiveAi1 calculating orders for player: %v", me)
	
	//list of all nodes I don't own and haven't sent guys to yet
	unclaimed := make([]state.NodeId, 0, len(s.Nodes))
	
	//list of all nodes with enemy guys on them
	enemy := make([]state.NodeId, 0, len(s.Nodes))
	
	//map of nodes where guys are (or are about to be) available to a map of delay to number of soldiers with that delay
	allAvailable := make(map[state.NodeId]map[int]int, len(s.Nodes))
	
	//total available guys across all nodes
	totalAvailable := 0
	
	//map of source and destination node IDs (concatenated together) to the length of the path between those nodes
	shortestPaths := make(map[state.NodeId]int)
	
	// fiterate over all nodes in order to populate the unclaimed list, enemy list, available soldiers map, and shortest paths map
	for _, node := range s.Nodes {
		enemyUnits := 0
		units := 0
		// count enemy and friendly units
		for playerId, numUnits := range node.Units {
			if playerId == me {
				units += numUnits
			} else {
				enemyUnits += numUnits
			}
		}
		//logger.Printf("totalAvailable: %v", totalAvailable)
		//logger.Printf("node: %v	has units: %v  enemyUnits: %v", node.Id, units, enemyUnits)
		// check for enemy node
		if enemyUnits > 0 {
			enemy = append(enemy, node.Id)
			for src := range allAvailable {
				shortestPaths[src + node.Id] = len(s.Path(src, node.Id, nil))
			}
		// check for unclaimed node
		} else if units <= 0 {
			unclaimed = append(unclaimed, node.Id)
			// find shortest paths from all available guys to this new unclaimed node
			for src := range allAvailable {
				shortestPaths[src + node.Id] = len(s.Path(src, node.Id, nil))
			}
		}
		// check for available units on node itself
		if units > 1 {
			// always leave 1 home to keep ownership of the node
			units--
			// is this node new to us?
			if len(allAvailable[node.Id]) == 0 {
				// create map
				allAvailable[node.Id] = map[int]int{0: units}
				// find shortest path from this new node to all unclaimed and enemy nodes
				for _, dst := range unclaimed {
					shortestPaths[node.Id + dst] = len(s.Path(node.Id, dst, nil))
				}
				for _, dst := range enemy {
					shortestPaths[node.Id + dst] = len(s.Path(node.Id, dst, nil))
				}
			} else {
				// add key to map
				allAvailable[node.Id][0] += units
			}
			totalAvailable += units
			//logger.Printf("totalAvailable (from node): %v", totalAvailable)
		}
		// check for available units on each edge leaving from this node
		for _, edge := range node.Edges {
			for index, unitMap := range edge.Units {
				if units = unitMap[me]; units > 0 {
					delay := len(edge.Units) - index
					// is this node new to us?
					if len(allAvailable[edge.Dst]) == 0 {
						// create map
						allAvailable[edge.Dst] = map[int]int{delay: units}
						// find shortest path from this new node to all unclaimed nodes
						for _, dst := range unclaimed {
							shortestPaths[edge.Dst + dst] = len(s.Path(edge.Dst, dst, nil))
						}
					} else {
						// add key to map
						allAvailable[edge.Dst][delay] += units
					}
					totalAvailable += units
					//logger.Printf("totalAvailable (from edge %v): %v", edge.Dst, totalAvailable)
				}
			}
		}
	}

	// print some data to the logger for debugging/info purposes
	logger.Printf("Unclaimed: %v", len(unclaimed))
	logger.Printf("Enemy: %v", len(enemy))
	logger.Printf("totalAvailable: %v", totalAvailable)
	//logger.Printf("allAvailable: %v", len(allAvailable))
	count := 0
	for src, availables := range allAvailable {
		for _, units := range availables {
			count += units
		}
		logger.Printf("src: %v  availables: %v  count: %v", src, availables, count)
	}
	logger.Printf("nodes: %v", len(s.Nodes))
	
	// for each unclaimed node, add shortest path to pathQueue
	pathQueue := make([]Path, 0, len(unclaimed))
	for _, node := range unclaimed {
		best := Path{
			dst:	node,
			length: -1,
		}
		logger.Printf("finding best available for: %v", node)
		for src, availables := range allAvailable {
			dist := shortestPaths[src + node]
			for delay, units := range availables {
				if units > 0 && (best.length < 0 || dist + delay < best.length) {
					best.src = src
					best.delay = delay
					best.length = dist + delay
				}
			}
		}
//		logger.Printf("best option is from: %v  delay: %v  length: %v", best.src, best.delay, best.length)
		// insert into queue
		pathQueue = append(pathQueue, best)
	}
	
	// sort the queue
	sort.Sort(ByDist{pathQueue})
	
	// as long as there are still units available, for each path in the queue, try to resolve it
	for totalAvailable > 0 && len(pathQueue) > 0 {
		next := pathQueue[0]
		pathQueue = pathQueue[1:]
		if allAvailable[next.src][next.delay] > 0 {
			if allAvailable[next.src][next.delay] == 1 {
				delete(allAvailable[next.src], next.delay)
			} else {
				allAvailable[next.src][next.delay] -= 1
			}
			totalAvailable--
			if next.delay == 0 {
				result = append(result, state.Order{
					Src:   next.src,
					Dst:   s.Path(next.src, next.dst, nil)[0],
					Units: 1,
				})
//				logger.Printf("sending 1 soldier from: %v  towards: %v", next.src, next.dst)
			}
		} else {
			logger.Printf("soldier no longer available, finding next best available for: %v  (totalAvail: %v  allAvail: %v)", next.dst, totalAvailable, len(allAvailable))
			best := Path{
				dst:	next.dst,
				length:	-1,
			}
			for src, availables := range allAvailable {
				dist := shortestPaths[src + next.dst]
				//logger.Printf("src: %v  availables: %v  dist: %v", src, availables, dist)
				for delay, units := range availables {
					//logger.Printf("delay: %v  units: %v", delay, units)
					if units > 0 && (best.length < 0 || dist + delay < best.length) {
						best.src = src
						best.delay = delay
						best.length = dist + delay
					}
				}
			}
			//logger.Printf("next best option is from: %v  delay: %v  length: %v", best.src, best.delay, best.length)
			// insert into queue
			pathQueue = append(pathQueue, best)
			// re-sort the queue
			sort.Sort(ByDist{pathQueue})
		}
	}
	
	// send remaining available units to closest node that has any enemy units on it
	if totalAvailable > 0 {
		for src, availables := range allAvailable {
			if units := availables[0]; units > 0 {
				//find closest enemy node
//				logger.Printf("finding closest enemy for %v available guys on %v", units, src)
				best := -1
				bestEdge := src
				for _, dst := range enemy {
					if dst == src {
						best = 0
						bestEdge = src
					}
					dist := shortestPaths[src + dst]
					if best < 0 || dist < best {
						best = dist
						bestEdge = s.Path(src, dst, nil)[0]
					}
				}
				if bestEdge == src {
//					logger.Printf("leaving them at home on %v", bestEdge)
				} else {
//					logger.Printf("sending them along %v", bestEdge)
					result = append(result, state.Order{
						Src:   src,
						Dst:   bestEdge,
						Units: units,
					})
				}
			}
		}
	}
	
	// all done, return the orders list
	return
}
