package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/k8scope/k8s-restart-app/internal/config"
	"github.com/k8scope/k8s-restart-app/internal/k8s"
	"github.com/k8scope/k8s-restart-app/internal/ledger"
	"github.com/k8scope/k8s-restart-app/internal/lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/client-go/kubernetes"
)

var (
	metricGaugeConnectedWatchers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "restart_app_connected_status_watchers",
		Help: "The number of connected status watchers",
	})
	metricCountRestarts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "restart_app_restarts_total",
		Help: "The total number of restarts",
	}, []string{"kind", "namespace", "name"})
	metricCountRestartsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "restart_app_restarts_failed_total",
		Help: "The total number of failed restarts",
	}, []string{"kind", "namespace", "name"})

	// upgrader is used to upgrade the HTTP connection to a WebSocket connection.
	// This is used to send status updates to the client.
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func getKindNamespaceNameFromRequest(r *http.Request) k8s.KindNamespaceName {
	kind := chi.URLParam(r, "kind")
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	return k8s.KindNamespaceName{
		Kind:      kind,
		Namespace: namespace,
		Name:      name,
	}
}

func MiddlewareValidation(config config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			kindNamespaceName := getKindNamespaceNameFromRequest(r)
			isFound := false

			if kindNamespaceName.Kind == "" || kindNamespaceName.Namespace == "" || kindNamespaceName.Name == "" {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}

			if kindNamespaceName.Kind != "Deployment" && kindNamespaceName.Kind != "StatefulSet" {
				http.Error(w, "invalid kind", http.StatusBadRequest)
				return
			}

			for _, service := range config.Services {
				if service.Kind == kindNamespaceName.Kind && service.Namespace == kindNamespaceName.Namespace && service.Name == kindNamespaceName.Name {
					isFound = true
					break
				}
			}
			if !isFound {
				http.Error(w, "service not found", http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Restart(client *kubernetes.Clientset, lck *lock.Lock) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		kindNamespaceName := getKindNamespaceNameFromRequest(r)
		metricCountRestarts.WithLabelValues(kindNamespaceName.Kind, kindNamespaceName.Namespace, kindNamespaceName.Name).Inc()
		err := k8s.RestartService(r.Context(), client, lck, kindNamespaceName)
		if errors.Is(err, lock.ErrResourceLocked) {
			http.Error(w, err.Error(), http.StatusLocked)
			return
		}
		if err != nil {
			metricCountRestartsFailed.WithLabelValues(kindNamespaceName.Kind, kindNamespaceName.Namespace, kindNamespaceName.Name).Inc()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func ListApplications(services config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(services)
		if err != nil {
			slog.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

func Status(ledger *ledger.Ledger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		metricGaugeConnectedWatchers.Inc()
		defer metricGaugeConnectedWatchers.Dec()

		// start listening for updates
		statusCh, unregister := ledger.Register()
		// when the client disconnects, we stop listening for updates and unregister the client
		defer unregister() //nolint:errcheck

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("failed to upgrade connection", "error", err)
			return
		}
		defer conn.Close() //nolint:errcheck

		for {
			select {
			case <-ctx.Done():
				// client disconnected
				slog.Info("client disconnected, stopping sending updates to client")
				return
			case status := <-statusCh:
				bts, err := json.Marshal(status)
				if err != nil {
					slog.Error("failed to marshal status", "error", err)
					return
				}

				err = conn.WriteMessage(websocket.TextMessage, bts)
				if err != nil {
					slog.Error("failed to write message", "error", err)
					return
				}
			}
		}
	}
}
