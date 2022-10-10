package main

import (
	"fmt"
	"net/http"

	"github.com/kevineaton/kesplora-api/api"
)

// main effectively sets up the API listener and then calls into the routes
func main() {
	// TODO: better logging
	fmt.Printf("\nStarting...\n")
	conf := api.SetupConfig()
	r := api.SetupAPI()

	err := http.ListenAndServe(fmt.Sprintf(":%s", conf.APIPort), r)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
}
