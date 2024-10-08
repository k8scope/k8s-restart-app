package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/k8scope/k8s-restart-app/internal/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	client *kubernetes.Clientset

	serviceAddress = "http://localhost:8080"
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
