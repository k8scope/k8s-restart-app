package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/k8scope/k8s-restart-app/internal/config"
	"github.com/k8scope/k8s-restart-app/internal/k8s"
	"github.com/k8scope/k8s-restart-app/internal/ledger"
	"github.com/k8scope/k8s-restart-app/internal/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	client *kubernetes.Clientset

	serviceAddress = utils.StringEnvOrDefault("SERVICE_ADDRESS", "http://localhost:8080")
)

func init() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", homedir+"/.kube/config")
	if err != nil {
		panic(err)
	}

	client, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
}

func Test_ListApplications(t *testing.T) {
	rsp, err := http.Get(serviceAddress + "/api/v1/service")
	if err != nil {
		t.Fatalf("failed to query services from api: %v", err)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", rsp.StatusCode)
		return
	}

	bts, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
		return
	}

	cfg := config.Config{}
	err = json.Unmarshal(bts, &cfg)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
		return
	}

	if len(cfg.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(cfg.Services))
		return
	}
}

func Test_Restart(t *testing.T) {
	var (
		testCases = []struct {
			Name               string
			Target             k8s.KindNamespaceName
			ExpectedStatusCode int
		}{
			{
				Name: "non existing service",
				Target: k8s.KindNamespaceName{
					Kind:      "Deployment",
					Namespace: "default",
					Name:      "nginx",
				},
				ExpectedStatusCode: http.StatusNotFound,
			},
			{
				Name: "successful restart",
				Target: k8s.KindNamespaceName{
					Kind:      "Deployment",
					Namespace: "test-1",
					Name:      "ngx-1",
				},
				ExpectedStatusCode: http.StatusNotFound,
			},
		}
	)

	for _, tc := range testCases {
		t.Run(tc.Name, func(tt *testing.T) {
			rsp, err := http.Post(fmt.Sprintf("%s/api/v1/service/%s/%s/%s", serviceAddress, tc.Target.Kind, tc.Target.Namespace, tc.Target.Name), "application/json", nil)
			if err != nil {
				t.Fatalf("Test_Restart(): failed to restart deployment: %v", err)
				return
			}

			if rsp.StatusCode != tc.ExpectedStatusCode {
				t.Fatalf("Test_Restart(): expected status code %d, got %d", tc.ExpectedStatusCode, rsp.StatusCode)
				return
			}

			// TODO: check if the service is locked during the restart
		})
	}
}

func Test_Status_HTTP(t *testing.T) {
	var (
		testCases = []struct {
			Name               string
			ExpectedStatusCode int
		}{
			{
				Name:               "expect 400, because the request does not initiate a websocket connection",
				ExpectedStatusCode: http.StatusBadRequest,
			},
		}
	)

	for _, tc := range testCases {
		t.Run(tc.Name, func(tt *testing.T) {
			rsp, err := http.Get(fmt.Sprintf("%s/api/v1/service/status", serviceAddress))
			if err != nil {
				t.Fatalf("Test_Status_HTTP(): failed to restart deployment: %v", err)
				return
			}

			if rsp.StatusCode != tc.ExpectedStatusCode {
				t.Fatalf("Test_Status_HTTP(): expected status code %d, got %d", tc.ExpectedStatusCode, rsp.StatusCode)
				return
			}
		})
	}
}

func Test_Status_WebSocket(t *testing.T) {
	var (
		testCases = []struct {
			Name               string
			ExpectedStatusCode int
			DeadlineSec        int
			ExpectMessage      bool
		}{
			{
				Name:               "expect connection switch",
				ExpectedStatusCode: http.StatusSwitchingProtocols,
				DeadlineSec:        2,
			},
		}
	)

	// parse URL
	urli, err := url.Parse(fmt.Sprintf("%s/api/v1/service/status", serviceAddress))
	if err != nil {
		t.Fatalf("Test_Status_WebSocket(): failed to parse URL: %v", err)
		return
	}

	if urli.Scheme == "https" {
		urli.Scheme = "wss"
	} else {
		urli.Scheme = "ws"
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(tt *testing.T) {
			conn, rsp, err := websocket.DefaultDialer.Dial(urli.String(), nil)
			if err != nil {
				t.Fatalf("Test_Status_WebSocket(): failed to connect to websocket: %v", err)
				return
			}

			if rsp.StatusCode != tc.ExpectedStatusCode {
				t.Fatalf("Test_Status(): expected status code %d, got %d", tc.ExpectedStatusCode, rsp.StatusCode)
				return
			}

			ctx, cf := context.WithTimeout(context.Background(), time.Second*time.Duration(tc.DeadlineSec))
			defer cf()

			for {
				select {
				case <-ctx.Done():
					t.Log("Test_Status(): deadline exceeded")
					return
				default:
					_, bts, err := conn.ReadMessage()
					if err != nil {
						t.Fatalf("Test_Status(): failed to read message: %v", err)
						return
					}

					objs := &ledger.ObjectStatus{}
					err = json.Unmarshal(bts, objs)
					if err != nil {
						t.Fatalf("Test_Status(): failed to parse message: %v", err)
						return
					}

					if (objs == nil) && tc.ExpectMessage {
						t.Fatalf("Test_Status(): expected message, got nil")
						return
					}
				}
			}
		})
	}
}
