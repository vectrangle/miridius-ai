Miridius AI
==

I wrote a of couple AIs for http://stockholm-ai.appspot.com/

This is intended as project for me to learn Go, so don't expect my code to be perfect ;)

Balanced AI
--

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

Aggressive AI
--

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
