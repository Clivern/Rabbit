// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"

	"github.com/stretchr/testify/assert"
)

func TestUnitEventActor(t *testing.T) {
	t.Run("names the access key behind the request", func(t *testing.T) {
		r := httptest.NewRequest("PUT", "/", nil)
		r.Header.Set("User-Agent", "python-httpx/0.28.1")
		r = r.WithContext(module.WithPrincipal(r.Context(), &module.Principal{
			AccessKey: &db.AccessKey{Id: "key-1", Name: "demo"},
		}))

		actor := (&API{}).EventActor(r)
		assert.Equal(t, db.EventActorTypeAccessKey, actor.ActorType)
		assert.Equal(t, db.Id("key-1"), actor.ActorId)
		assert.Equal(t, "demo", actor.ActorName)
		assert.Equal(t, "python-httpx/0.28.1", actor.UserAgent)
		assert.NotEmpty(t, actor.IPAddress)
	})

	t.Run("names the user behind the request", func(t *testing.T) {
		r := httptest.NewRequest("PUT", "/", nil)
		r = r.WithContext(module.WithPrincipal(r.Context(), &module.Principal{
			User: &db.User{Id: "user-1", Name: "Ahmed"},
		}))

		actor := (&API{}).EventActor(r)
		assert.Equal(t, db.EventActorTypeUser, actor.ActorType)
		assert.Equal(t, db.Id("user-1"), actor.ActorId)
		assert.Equal(t, "Ahmed", actor.ActorName)
	})
}

func TestUnitEventStream(t *testing.T) {
	w := httptest.NewRecorder()

	emit := (&API{}).EventStream(w, time.Minute)
	emit("delta", map[string]string{"content": "Hi"})
	emit("done", map[string]string{"status": "idle"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(t, "event: delta\ndata: {\"content\":\"Hi\"}\n\nevent: done\ndata: {\"status\":\"idle\"}\n\n", w.Body.String())
	assert.True(t, w.Flushed)
}
