package main

// Note: Although most API calls specify their intended HTTP method, they
// currently accept every HTTP method.

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/websocket"
	irma "github.com/privacybydesign/irmago"
	"github.com/privacybydesign/irmago/server"
)

type DTMF = string
type Secret = string

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Accept any origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

var irmaExternalURLRegexp *regexp.Regexp = regexp.MustCompile(`^http(s?)://(.*)/irma/session`)

type SessionResponse struct {
	SessionPtr  *irma.Qr `json:"sessionPtr,omitempty"`
	Phonenumber string   `json:"phonenumber,omitempty"`
	Dtmf        string   `json:"dtmf,omitempty"`
}

func (cfg Configuration) phonenumber(dtmf string) string {
	return cfg.ServicePhoneNumber + "," + dtmf
}

func (cfg Configuration) irmaRequest(purpose string, dtmf string) (irma.RequestorRequest, error) {
	condiscon, ok := cfg.PurposeToAttributes[purpose]
	if !ok {
		return nil, fmt.Errorf("unknown call purpose: %#v", purpose)
	}

	disclosure := irma.NewDisclosureRequest()
	disclosure.Disclose = condiscon
	disclosure.ClientReturnURL = "tel:" + cfg.phonenumber(dtmf)

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
func (cfg Configuration) handleSession(w http.ResponseWriter, r *http.Request) {
	// This function is responsible for ensuring the irma session secret is
	// stored in the database before it returns the QR code to the user.
	purpose := r.FormValue("purpose")
	dtmf, err := cfg.db.NewSession(purpose)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	request, err := cfg.irmaRequest(purpose, dtmf)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transport := irma.NewHTTPTransport(cfg.IrmaServerURL)

	if cfg.IrmaHeaderValue != "" {
		var headerKey string
		headerKey = cfg.IrmaHeaderKey
		if headerKey == "" {
			headerKey = "Authorization"
		}

		transport.SetHeader(headerKey, cfg.IrmaHeaderValue)
	}

	var pkg server.SessionPackage
	err = transport.Post("session", &pkg, request)
	if err != nil {
		log.Printf("failed to request irma session: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = cfg.db.storeSecret(dtmf, pkg.Token)
	if err != nil {
		log.Printf("failed to store irma secret: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var session SessionResponse
	session.SessionPtr = pkg.SessionPtr
	session.Phonenumber = cfg.phonenumber(dtmf)
	session.Dtmf = dtmf

	if cfg.IrmaExternalURL != "" {
		// Rewrite IRMA server url to match irma-external-url arg
		baseURL := fmt.Sprintf("%v/irma/session", cfg.IrmaExternalURL)
		session.SessionPtr.URL = irmaExternalURLRegexp.ReplaceAllString(pkg.SessionPtr.URL, baseURL)
	}

	// Update the request server URL to include the session token.
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		log.Printf("failed to marshal QR code: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	go cfg.pollIrmaSessionDaemon(pkg.Token)
	w.Write(sessionJSON)
}

// Should be called when the session status becomes DONE.
func (cfg Configuration) cacheDisclosedAttributes(sessionToken string) {
	transport := irma.NewHTTPTransport(
		fmt.Sprintf("%s/session/%s/", cfg.IrmaServerURL, sessionToken))

	result := &server.SessionResult{}
	err := transport.Get("result", result)
	if err != nil {
		log.Printf("failed to get irma session result: %v", err)
		return
	}

	status := string(result.Status)
	if status != "DONE" {
		log.Printf("unexpected irma session status: %#v", status)
		return
	}

	disclosedData := result.Disclosed
	disclosedJSON, err := json.Marshal(disclosedData)
	if err != nil {
		log.Printf("failed to marshal disclosed attributes: %v", err)
		return
	}

	disclosed := string(disclosedJSON)
	err = cfg.db.storeDisclosed(sessionToken, disclosed)
	if err != nil {
		log.Printf("failed to store disclosed attributes: %v", err)
		return
	}
}

// Upgrade connection to websocket, start polling IRMA session,
// Send IRMA session updates over websocket
func (cfg Configuration) handleSessionStatus(w http.ResponseWriter, r *http.Request) {
	dtmf := r.FormValue("dtmf")
	if dtmf == "" {
		http.Error(w, "No dtmf passed", http.StatusBadRequest)
		return
	}

	sessionToken, err := cfg.db.secretFromDTMF(dtmf)
	if err == ErrNoRows {
		http.Error(w, "session not found", http.StatusNotFound)
	} else if err != nil {
		log.Printf("failed to retrieve secret from dtmf: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	statusUpdates := make(chan Message)
	cfg.broadcaster.Subscribe(sessionToken, statusUpdates)
	defer cfg.broadcaster.Unsubscribe(sessionToken, statusUpdates)

	status, err := cfg.db.getStatus(sessionToken)
	if err != nil {
		http.Error(w, "unknown token", http.StatusNotFound)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("failed to upgrade session status connection:", err)
		return
	}
	defer ws.Close()

	err = ws.WriteMessage(websocket.TextMessage, []byte(status))
	if err != nil {
		log.Printf("failed to write session status on fresh websocket: %v", err)
		return
	}

	for status := range statusUpdates {
		if status.Key != "status" {
			continue
		}

		msg := []byte(status.Value)
		err = ws.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			// Write error after write has succeeded before,
			// the client probably disconnected.
			return
		} else if IrmaStatusIsFinal(status.Value) {
			return
		}
	}
}

// A citizen has called the service number. Amazon connect picked up and
// triggered a lambda. The lambda made a POST request to the backend, handled
// here. This POST request should contain only the DTMF code the caller sent.
// We respond with a fresh secret that will be placed as metadata in the call by
// the lambda. The secret can later be used by the agent frontend to receive the
// disclosed Irma attributes.
// This handler should only be exposed on an internal port, reachable from the
// related Amazon Lambda but not from the internet.
func (cfg Configuration) handleCall(w http.ResponseWriter, r *http.Request) {
	dtmf := r.FormValue("dtmf")
	secret, err := cfg.db.secretFromDTMF(dtmf)
	if err == ErrNoRows {
		http.Error(w, "session not found", http.StatusNotFound)
	} else if err != nil {
		log.Printf("failed to retrieve secret from dtmf: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	} else {
		io.WriteString(w, secret)
		cfg.db.setStatus(secret, "CALLED")
		log.Printf("someone called %v", secret)
	}
}

type DiscloseResponse struct {
	Purpose   string          `json:"purpose"`
	Disclosed json.RawMessage `json:"disclosed"`
}

// An agent frontend has accepted a call and sends us a GET request with the
// associated secret. We respond with the disclosed attributes. If the disclosed
// attributes are not yet available, we synchronously poll the IRMA server to
// get them.
func (cfg Configuration) handleDisclose(w http.ResponseWriter, r *http.Request) {
	secret := r.FormValue("secret")
	if secret == "" {
		http.Error(w, "disclosure needs secret", http.StatusBadRequest)
		return
	}

	purpose, disclosed, err := cfg.db.getDisclosed(secret)
	if err == ErrNoRows {
		// invalid or expired secret
		http.Error(w, "session not found", http.StatusNotFound)
		return
	} else if err != nil {
		// some database error
		log.Printf("failed to get disclosed attributes: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	} else if disclosed == "" {
		// disclosed not set yet, irma session probably failed
		log.Printf("disclosed attributes not yet received")
		http.Error(w, "irma session failed", http.StatusUnauthorized)
		return
	}

	// return valid disclosed attributes
	response := DiscloseResponse{
		Purpose:   purpose,
		Disclosed: json.RawMessage([]byte(disclosed)),
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("failed to marshal disclose response: %#v", response)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(responseJSON)
}

func (cfg Configuration) handleSessionUpdate(w http.ResponseWriter, r *http.Request) {
	secret := r.FormValue("secret")
	status := r.FormValue("status")
	cfg.db.setStatus(secret, status)
}

func (cfg Configuration) handleMetrics(w http.ResponseWriter, r *http.Request) {
	response, err := cfg.getConnectCurrentMetrics()

	if err != nil {
		log.Printf("failed to fetch metrics: %#v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(response)

	if err != nil {
		log.Printf("failed to marshal disclose response: %#v", response)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(responseJSON)
}

// Status panel waitlist status feed.
// Upgrade connection to websocket, register a channel with the ConnectPoll,
// pass updates to websocket.
func (cfg Configuration) handleAgentFeed(w http.ResponseWriter, r *http.Request) {
	waitListStatus := make(Listener)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("failed to upgrade session status connection:", err)
		return
	}
	defer ws.Close()

	cfg.broadcaster.Subscribe("kcc", waitListStatus)
	defer cfg.broadcaster.Unsubscribe("kcc", waitListStatus)

	for update := range waitListStatus {
		if update.Key != "amazon-connect" {
			continue
		}

		err = ws.WriteMessage(websocket.TextMessage, []byte(update.Value))
		if err != nil {
			break
		}
	}

}
