package bettermux

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBetterMux_HandleRouteWithMiddleware(t *testing.T) {
	root := New()

	root.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	withMw := root.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	})

	withMw.Route("/api/", func(api BetterMux) {
		api.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("api hello"))
		})

		api.HandleFunc("/mixed/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("mixed subtree"))
		})
	})

	withMw.Route("/admin", func(admin BetterMux) {
		admin.HandleFunc("/dashboard", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("admin dashboard"))
		})
	})

	tests := []struct {
		name       string
		path       string
		wantBody   string
		wantHeader string
		wantStatus int
	}{
		{"root handler", "/ping", "pong", "", http.StatusOK},
		{"route with middleware", "/api/hello", "api hello", "applied", http.StatusOK},
		{"route subtree with middleware", "/api/mixed/", "mixed subtree", "applied", http.StatusOK},
		{"route without trailing slash path", "/admin/dashboard", "admin dashboard", "applied", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			res := httptest.NewRecorder()

			root.ServeHTTP(res, req)

			if res.Code != tt.wantStatus {
				t.Fatalf("%s: expected status %d, got %d", tt.name, tt.wantStatus, res.Code)
			}

			if body := res.Body.String(); body != tt.wantBody {
				t.Fatalf("%s: expected body %q, got %q", tt.name, tt.wantBody, body)
			}

			if got := res.Header().Get("X-Middleware"); got != tt.wantHeader {
				t.Fatalf("%s: expected X-Middleware header %q, got %q", tt.name, tt.wantHeader, got)
			}
		})
	}
}

func TestBetterMux_MethodHelpersAndRouteTemplateCompatibility(t *testing.T) {
	root := New()

	withMw := root.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	})

	withMw.Route("/api/", func(api BetterMux) {
		api.Get("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("get user"))
		})
		api.Post("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("create user"))
		})
		api.Put("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("update user"))
		})
		api.Patch("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("patch user"))
		})
		api.Delete("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("delete user"))
		})
		api.Options("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	})

	tests := []struct {
		name       string
		method     string
		path       string
		wantBody   string
		wantHeader string
		wantStatus int
	}{
		{"GET users template", http.MethodGet, "/api/users/42", "get user", "applied", http.StatusOK},
		{"POST users template", http.MethodPost, "/api/users/42", "create user", "applied", http.StatusCreated},
		{"PUT users template", http.MethodPut, "/api/users/99", "update user", "applied", http.StatusOK},
		{"PATCH users template", http.MethodPatch, "/api/users/7", "patch user", "applied", http.StatusOK},
		{"DELETE users template", http.MethodDelete, "/api/users/12", "delete user", "applied", http.StatusOK},
		{"OPTIONS users template", http.MethodOptions, "/api/users/123", "", "applied", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			res := httptest.NewRecorder()

			root.ServeHTTP(res, req)

			if res.Code != tt.wantStatus {
				t.Fatalf("%s: expected status %d, got %d", tt.name, tt.wantStatus, res.Code)
			}

			if body := res.Body.String(); body != tt.wantBody {
				t.Fatalf("%s: expected body %q, got %q", tt.name, tt.wantBody, body)
			}

			if got := res.Header().Get("X-Middleware"); got != tt.wantHeader {
				t.Fatalf("%s: expected X-Middleware header %q, got %q", tt.name, tt.wantHeader, got)
			}
		})
	}
}
