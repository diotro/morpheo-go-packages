package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/MorpheoOrg/go-morpheo/common"
	uuid "github.com/satori/go.uuid"
)

// Orchestrator HTTP API routes
const (
	OrchestratorStatusUpdateRoute = "/update_status"
	OrchestratorLearnResultRoute  = "/learndone"
	OrchestratorPredResultRoute   = "/preddone"
)

// Orchestrator describes Morpheo's orchestrator API
type Orchestrator interface {
	UpdateUpletStatus(upletType, status string, upletID uuid.UUID) error
	PostLearnResult(learnupletID uuid.UUID, data io.Reader) error
}

// OrchestratorAPI is a wrapper around our orchestrator API
type OrchestratorAPI struct {
	Orchestrator

	Hostname string
	Port     int
}

// UpdateUpletStatus changes the status field of a learnuplet/preduplet
func (o *OrchestratorAPI) UpdateUpletStatus(upletType string, status string, upletID uuid.UUID) error {
	if _, ok := common.ValidUplets[upletType]; !ok {
		return fmt.Errorf("[orchestrator-api] Uplet type \"%s\" is invalid. Allowed values are %s", upletType, common.ValidUplets)
	}
	if _, ok := common.ValidStatuses[status]; !ok {
		return fmt.Errorf("[orchestrator-api] Status \"%s\" is invalid. Allowed values are %s", status, common.ValidStatuses)
	}
	url := fmt.Sprintf("http://%s:%d%s/%s/%s", o.Hostname, o.Port, OrchestratorStatusUpdateRoute, upletType, upletID)

	payload, err := json.Marshal(map[string]string{"status": status})
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error JSON-marshaling status update payload: %s", url, err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error building status update POST request against %s: %s", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error performing status update POST request against %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[orchestrator-api] Unexpected status code (%s): status update POST request against %s", resp.Status, url)
	}
	return nil
}

func (o *OrchestratorAPI) postData(route string, upletID uuid.UUID, data io.Reader) error {
	url := fmt.Sprintf("http://%s:%d%s/%s", o.Hostname, o.Port, route, upletID)

	req, err := http.NewRequest(http.MethodPost, url, data)
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error building result POST request against %s: %s", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[orchestrator-api] Error performing result POST request against %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[orchestrator-api] Unexpected status code (%s): result POST request against %s", resp.Status, url)
	}
	return nil
}

// PostLearnResult forwards a JSON-formatted learn result to the orchestrator HTTP API
func (o *OrchestratorAPI) PostLearnResult(learnupletID uuid.UUID, data io.Reader) error {
	return o.postData(OrchestratorLearnResultRoute, learnupletID, data)
}

// OrchestratorAPIMock mocks the Orchestrator API, always returning ok to update queries except for
// given "unexisting" pred/learn uplet with a given UUID
type OrchestratorAPIMock struct {
	Orchestrator

	UnexistingUplet string
}

// NewOrchestratorAPIMock returns with a mock of the Orchestrator API
func NewOrchestratorAPIMock() (s *OrchestratorAPIMock) {
	return &OrchestratorAPIMock{
		UnexistingUplet: "ea408171-0205-475e-8962-a02855767260",
	}
}

// UpdateUpletStatus returns nil except if OrchestratorAPIMock.UnexistingUpletID is passed
func (o *OrchestratorAPIMock) UpdateUpletStatus(upletType, status string, upletID uuid.UUID) error {
	if upletID.String() != o.UnexistingUplet {
		log.Printf("[orchestrator-mock] Received update status for %s-uplet %s. Status: %s", upletType, upletID, status)
		return nil
	}
	return fmt.Errorf("[orchestrator-mock][status-update] Unexisting uplet %s", upletID)
}

// PostLearnResult returns nil except if OrchestratorAPIMock.UnexistingUpletID is passed
func (o *OrchestratorAPIMock) PostLearnResult(learnupletID uuid.UUID, dataReader io.Reader) error {
	if learnupletID.String() != o.UnexistingUplet {
		buf := bytes.Buffer{}
		_, err := buf.ReadFrom(dataReader)
		data := buf.String()
		log.Printf("[orchestrator-mock] Received learn result for learn-uplet %s: \n %s", learnupletID, data)
		if err != nil {
			log.Printf("[orchestrator-mock] Error forwarding performance to stdout: %s", err)
		}
		return nil
	}
	return fmt.Errorf("[orchestrator-mock][status-update] Unexisting uplet %s", learnupletID)
}
