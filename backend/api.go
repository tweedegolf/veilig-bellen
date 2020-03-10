package main

import "encoding/json"
import "fmt"
import "log"
import "net/http"

import _ "golang.org/x/net/websocket"
import "github.com/privacybydesign/irmago"
import "github.com/privacybydesign/irmago/server"

type Configuration struct {
	PostgresAddress     string
	ListenAddress       string
	IrmaServerURL       string
	ServicePhoneNumber  string
	PurposeToAttributes map[string]irma.AttributeConDisCon
}

// TODO: These fields should be in a database, not in memory.
type Session struct {
	// The DTMF code that relates the incoming call to the Irma session.
	DTMF string
	// The secret that allows retrieving this session's disclosed Irma
	// attributes from the backend.
	DisclosureSecret    string
	DisclosedAttributes []irma.DisclosedAttribute
}

func (cfg Configuration) generateDTMF() string {
	return "0000000000" // TODO
}

func (cfg Configuration) irmaRequest(purpose string, dtmf string) (irma.RequestorRequest, error) {
	condiscon, ok := cfg.PurposeToAttributes[purpose]
	if !ok {
		return nil, fmt.Errorf("Unknown call purpose: %#v", purpose)
	}

	disclosure := irma.NewDisclosureRequest()
	disclosure.Disclose = condiscon
	disclosure.ClientReturnURL = "tel:" + cfg.ServicePhoneNumber + dtmf

	request := &irma.ServiceProviderRequest{
		Request: disclosure,
	}

	return request, nil
}

// A citizen pressed the call with Irma button on a page on the Gemeente
// Nijmegen website in order to start a new calling Irma session. The citizen
// frontend makes a POST request to the backend with only one piece of
// information: The subject the citizen wants to ask about. The backend looks up
// what Irma attributes the agent will need and asks the Irma server for a new
// Irma session. The backend generates a DTMF code and responds with a JSON
// object with a valid Irma session response with a tel return url containing
// the DTMF code.
func (cfg Configuration) handleNewSession(w http.ResponseWriter, r *http.Request) {
	purpose := r.FormValue("purpose")
	dtmf := cfg.generateDTMF()
	request, err := cfg.irmaRequest(purpose, dtmf)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transport := irma.NewHTTPTransport(cfg.IrmaServerURL)
	var pkg server.SessionPackage
	err = transport.Post("session", &pkg, request)
	if err != nil {
		log.Printf("failed to request irma session: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	qr := pkg.SessionPtr
	// Update the request server URL to include the session token.
	transport.Server += fmt.Sprintf("session/%s/", pkg.Token)
	qrJSON, err := json.Marshal(qr)
	if err != nil {
		log.Printf("failed to marshal QR code to JSON: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: Spawn goroutine to poll irma server.
	// TODO: Persist session in cluster.

	w.Write(qrJSON)
}

// A citizen has called the service number. Amazon connect picked up and
// triggered a lambda. The lambda made a POST request to the backend, handled
// here. This POST request should contain only the DTMF code the caller sent.
// We respond with a fresh secret that will be placed as metadata in the call by
// the lambda. The secret can later be used by the agent frontend to receive the
// revealed Irma attributes.
// TODO: This needs authentication.
func (cfg Configuration) handleCall(w http.ResponseWriter, r *http.Request) {

}
