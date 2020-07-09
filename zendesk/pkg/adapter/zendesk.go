/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package adapter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nukosuke/go-zendesk/zendesk"
	"go.uber.org/zap"
)

// ZendeskAPIHandler listen for Zendesk API Events
type ZendeskAPIHandler interface {
	Start(stopCh <-chan struct{}) error
}

// constats for the CE data
const (
	ceType    = "com.zendesk.ticket.new"
	ceSource  = "com.zendesk.source"
	ceSubject = "New Zendesk Ticket"
)

type zendeskAPIHandler struct {
	port     int
	username string
	password string

	ceClient cloudevents.Client
	srv      *http.Server

	logger *zap.SugaredLogger
}

// NewZendeskAPIHandler creates the default implementation of the Zendesk API Events handler
func NewZendeskAPIHandler(ceClient cloudevents.Client, port int, username, password string, logger *zap.SugaredLogger) ZendeskAPIHandler {
	return &zendeskAPIHandler{
		port:     port,
		username: username,
		password: password,

		ceClient: ceClient,
		logger:   logger,
	}
}

// Start the server for receiving Zendesk events. Will block until the stop channel closes.
func (h *zendeskAPIHandler) Start(stopCh <-chan struct{}) error {
	h.logger.Info("Starting Zendesk event handler...")

	m := http.NewServeMux()
	m.HandleFunc("/", h.handleAll)

	h.srv = &http.Server{
		Addr:    ":" + strconv.Itoa(h.port),
		Handler: m,
	}

	done := make(chan bool, 1)
	go h.gracefulShutdown(stopCh, done)

	h.logger.Infof("Zendesk Source is ready to handle requests at %s", h.srv.Addr)
	if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on %s: %v", h.srv.Addr, err)
	}

	<-done
	h.logger.Infof("Server stopped")
	return nil
}

func (h *zendeskAPIHandler) validateAuthHeader(r *http.Request) error {

	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return errors.New("No Auth Parameters")
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return errors.New("Could not decode the auth header")
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return errors.New("misformated credentials at auth header")
	}

	if pair[0] == h.username && pair[1] == h.password {
		return nil
	}
	return nil
}

// handleAll receives all Zendesk events at a single resource, it
// is up to this function to parse event wrapper and dispatch.
func (h *zendeskAPIHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.handleError(errors.New("request without body not supported"), w)
		return
	}

	if err := h.validateAuthHeader(r); err != nil {
		h.handleError(err, w)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.handleError(err, w)
		return
	}

	event := &zendesk.Ticket{}
	err = json.Unmarshal(body, event)
	if err != nil {
		h.handleError(fmt.Errorf("could not unmarshall JSON request: %s", err.Error()), w)
		return
	}

	cEvent, err := h.cloudEventFromTicket(event)
	if err != nil {
		h.logger.Info("Error Creating CloudEvent")
		h.handleError(err, w)
	}

	if result := h.ceClient.Send(context.Background(), *cEvent); !cloudevents.IsACK(result) {
		h.logger.Info("Error Sending CloudEvent")
		h.handleError(result, w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write([]byte("500 - Something bad happened!"))
	if err != nil {
		h.logger.Info("Error Writing HTTP response")
		h.handleError(err, w)
	}
}

func (h *zendeskAPIHandler) gracefulShutdown(stopCh <-chan struct{}, done chan<- bool) {
	<-stopCh
	h.logger.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h.srv.SetKeepAlivesEnabled(false)
	if err := h.srv.Shutdown(ctx); err != nil {
		h.logger.Fatalf("Could not gracefully shutdown the server: %v", err)
	}
	close(done)
}

func (h *zendeskAPIHandler) handleError(err error, w http.ResponseWriter) {
	h.logger.Error("An error ocurred", zap.Error(err))
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (h *zendeskAPIHandler) cloudEventFromTicket(ticket *zendesk.Ticket) (*cloudevents.Event, error) {
	h.logger.Info("Proccesing Zendesk event")
	data, err := json.Marshal(ticket)
	if err != nil {
		return nil, err
	}
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	ceID := strconv.Itoa(int(ticket.ID))

	event.SetID(ceID)
	event.SetType(ceType)
	event.SetSource(ceSource)
	//event.SetExtension("api_app_id", "wrapper.APIAppID")
	//event.SetTime(time.Unix(int64(120), 0))
	event.SetSubject(ceSubject)
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("failed to set event data: %w", err)
	}

	return &event, nil
}
