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
	"k8s.io/client-go/kubernetes"
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
		kindNamesaceName := getKindNamespaceNameFromRequest(r)
		err := k8s.RestartService(r.Context(), client, kindNamesaceName)
		if err != nil {
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
