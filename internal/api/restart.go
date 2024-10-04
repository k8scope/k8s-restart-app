package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/k8scope/k8s-restart-app/internal/config"
	"github.com/k8scope/k8s-restart-app/internal/k8s"
	"github.com/k8scope/k8s-restart-app/internal/ledger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/client-go/kubernetes"
)

var (
	metricGaugeConnectedWatchers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restart_app_connected_status_watchers",
		Help: "The number of connected status watchers",
	}, []string{"kind", "namespace", "name"})
	metricCountRestarts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "restart_app_restarts_total",
		Help: "The total number of restarts",
	}, []string{"kind", "namespace", "name"})
	metricCountRestartsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "restart_app_restarts_failed_total",
		Help: "The total number of failed restarts",
	}, []string{"kind", "namespace", "name"})
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

func Restart(client *kubernetes.Clientset) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		kindNamespaceName := getKindNamespaceNameFromRequest(r)
		metricCountRestarts.WithLabelValues(kindNamespaceName.Kind, kindNamespaceName.Namespace, kindNamespaceName.Name).Inc()
		err := k8s.RestartService(r.Context(), client, kindNamespaceName)
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
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		kindNamespaceName := getKindNamespaceNameFromRequest(r)

		metricGaugeConnectedWatchers.WithLabelValues(kindNamespaceName.Kind, kindNamespaceName.Namespace, kindNamespaceName.Name).Inc()
		defer metricGaugeConnectedWatchers.WithLabelValues(kindNamespaceName.Kind, kindNamespaceName.Namespace, kindNamespaceName.Name).Dec()

		// start listening for updates
		statusCh, unregister := ledger.Register(kindNamespaceName)
		// when the client disconnects, we stop listening for updates and unregister the client
		defer unregister()

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("failed to upgrade connection", "error", err)
			return
		}
		defer conn.Close()

		for {
			select {
			case <-ctx.Done():
				// client disconnected
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
