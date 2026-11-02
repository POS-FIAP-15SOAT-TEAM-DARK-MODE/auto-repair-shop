// Package e2e runs the BDD end-to-end scenarios of the local ecosystem with godog.
//
// The suite talks only to the gateway (E2E_BASE_URL, default http://localhost:8080). It needs the ecosystem running
// (docker compose up, see the README); it is not a unit test and does not run in `go test ./...` of the gateway.
package e2e

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "auto-repair-shop-e2e",
		ScenarioInitializer: initializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Strict:   true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("e2e scenarios failed")
	}
}
