package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/k8scope/k8s-restart-app/internal/config"
	"github.com/k8scope/k8s-restart-app/internal/k8s"
)

func TestMiddlewareValidation(t *testing.T) {
	type fields struct {
		config config.Config
	}
	type args struct {
		prams map[string]string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
	}{
		{
			name: "with kind, namespace and name",
			fields: fields{
				config: config.Config{
					Services: []k8s.KindNamespaceName{
						{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "test",
						},
					},
				},
			},
			args: args{
				prams: map[string]string{
					"kind":      "Deployment",
					"namespace": "default",
					"name":      "test",
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "without kind, namespace and name",
			fields: fields{
				config: config.Config{
					Services: []k8s.KindNamespaceName{
						{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "test",
						},
					},
				},
			},
			args: args{
				prams: map[string]string{},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "with wrong kind",
			fields: fields{
				config: config.Config{
					Services: []k8s.KindNamespaceName{
						{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "test",
						},
					},
				},
			},
			args: args{
				prams: map[string]string{
					"kind":      "Service",
					"namespace": "default",
					"name":      "test",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service not found in config",
			fields: fields{
				config: config.Config{
					Services: []k8s.KindNamespaceName{
						{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "test",
						},
					},
				},
			},
			args: args{
				prams: map[string]string{
					"kind":      "Deployment",
					"namespace": "abc",
					"name":      "test",
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/v1/service/{kind}/{namespace}/{name}/restart", nil)

			rctx := chi.NewRouteContext()
			for k, v := range tt.args.prams {
				rctx.URLParams.Add(k, v)
			}

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			MiddlewareValidation(tt.fields.config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("MiddlewareValidation() status mismatch = %v, want %v", w.Code, tt.wantStatus)
				return
			}
		})
	}
}
