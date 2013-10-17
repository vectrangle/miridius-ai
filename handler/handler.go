package handler

import (
	"net/http"
	"fmt"
	"github.com/zond/stockholm-ai/ai"
	"github.com/zond/stockholm-ai/hub/common"
	"github.com/miridius/ai/aggressiveAi"
	"github.com/miridius/ai/balancedAi"
)

func init() {
	http.HandleFunc("/balanced/v1", ai.HTTPHandlerFunc(common.GAELoggerFactory, balancedAi.BalancedAi1{}))
	http.HandleFunc("/aggressive/v1.1", ai.HTTPHandlerFunc(common.GAELoggerFactory, aggressiveAi.AggressiveAi1{}))
	http.HandleFunc("/", hello)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!\n\n")
	fmt.Fprintf(w, "Currently serving:\n/balanced/v1\n/aggressive/v1.1")
}
