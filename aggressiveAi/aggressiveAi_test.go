package aggressiveAi

import (
    "github.com/zond/stockholm-ai/state"
    "log"
    "os"
    "testing"
)

func TestOrders(t *testing.T) {
    // define logger
    logger := log.New(os.Stdout, "", 0)

    // set up players
    players := make([]state.PlayerId, 4)
    players[0] = "a"
    players[1] = "b"
    players[2] = "c"
    players[3] = "d"

    //set up game
    ai := AggressiveAi1{}
    orderMap := make(map[state.PlayerId]state.Orders, len(players))
    s := state.RandomState(logger, players)

    //play game
    var onlyPlayerLeft *state.PlayerId
    for onlyPlayerLeft == nil {
        for _, player := range players {
            orderMap[player] = ai.Orders(logger, player, s)
        }
        onlyPlayerLeft = s.Next(logger, orderMap)
        //logger.Printf("orders: %v", orderMap)
        logger.Printf("state: %v", s)
    }

    //print winner
    logger.Printf("onlyPlayerLeft: %v", *onlyPlayerLeft)
}
