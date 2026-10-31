// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/middleware"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/util"
)

// ErrorMapping maps a module error to an HTTP status and a translated message.
type ErrorMapping struct {
	Err     error
	Status  int
	Message string
}

// EventStream switches the response to server-sent events and returns a function that writes and flushes one event.
func (a *API) EventStream(w http.ResponseWriter, timeout time.Duration) func(event string, data any) {
	controller := http.NewResponseController(w)
	_ = controller.SetWriteDeadline(time.Now().Add(timeout))

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	return func(event string, data any) {
		body, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, body)
		_ = controller.Flush()
	}
}

// WriteJSON writes data as a JSON response using util.WriteJSON.
func (a *API) WriteJSON(w http.ResponseWriter, statusCode int, data any) error {
	return util.WriteJSON(w, statusCode, data)
}

// DecodeJSON reads and decodes JSON from the request body.
func (a *API) DecodeJSON(r *http.Request, v any) error {
	return util.DecodeJSON(r, v)
}

// DecodeAndValidate decodes JSON and validates the struct in one step.
func (a *API) DecodeAndValidate(r *http.Request, v any) error {
	return util.DecodeAndValidate(r, v)
}

// WriteValidationError writes a decode or validation error as a 400 JSON response.
func (a *API) WriteValidationError(w http.ResponseWriter, err error) {
	util.WriteValidationError(w, err)
}

// ParsePagination parses limit and offset from query parameters.
func (a *API) ParsePagination(r *http.Request) (limit, offset int) {
	return util.ParsePagination(r)
}

// GetUser returns the authenticated user from the request context.
func (a *API) GetUser(r *http.Request) (*db.User, bool) {
	return middleware.GetUserFromContext(r.Context())
}

// WriteError writes the mapped response for err and reports false when err has no mapping.
func (a *API) WriteError(w http.ResponseWriter, err error, mappings []ErrorMapping) bool {
	for _, m := range mappings {
		if errors.Is(err, m.Err) {
			a.WriteJSON(w, m.Status, map[string]any{
				"errorMessage": m.Message,
			})

			return true
		}
	}

	return false
}

// EventActor is the user or access key behind the request, as the actor of the events it triggers.
func (a *API) EventActor(r *http.Request) module.EventRequest {
	principal, _ := module.PrincipalFrom(r.Context())

	actor := module.EventRequest{
		IPAddress: middleware.GetClientIP(r),
		UserAgent: r.UserAgent(),
	}

	switch {
	case principal.IsUser():
		actor.ActorType = db.EventActorTypeUser
		actor.ActorId = principal.User.Id
		actor.ActorName = principal.User.Name
	case principal.IsAccessKey():
		actor.ActorType = db.EventActorTypeAccessKey
		actor.ActorId = principal.AccessKey.Id
		actor.ActorName = principal.AccessKey.Name
	}

	return actor
}
