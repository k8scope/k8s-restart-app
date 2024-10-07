package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/k8scope/k8s-restart-app/internal/lock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

var (
	ErrInvalidKindNamespaceNameFormat = fmt.Errorf("invalid format")
	ErrInvalidKind                    = fmt.Errorf("invalid kind")
)

type KindNamespaceName struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// KindNamespaceNameFromString parses a string into a KindNamespaceName
// The string should be in the format of "Kind/Namespace/Name"
//
// Example:
//
//	KindNamespaceNameFromString("Deployment/my-namespace/my-deployment")
//
// This will return a KindNamespaceName with Kind: Deployment, Namespace: my-namespace, Name: my-deployment
func KindNamespaceNameFromString(s string) (*KindNamespaceName, error) {
	segment := strings.Split(s, "/")
	if len(segment) != 3 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidKindNamespaceNameFormat, s)
	}
	if segment[0] != "Deployment" && segment[0] != "StatefulSet" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidKind, segment[0])
	}
	return &KindNamespaceName{
		Kind:      segment[0],
		Namespace: segment[1],
		Name:      segment[2],
	}, nil
}

func (s KindNamespaceName) String() string {
	return s.Kind + "/" + s.Namespace + "/" + s.Name
}

func RestartService(ctx context.Context, clientset *kubernetes.Clientset, lock *lock.Lock, service KindNamespaceName) error {
	switch service.Kind {
	case "Deployment":
		err := lock.Lock(service.String())
		if err != nil {
			// we don't want to unlock the lock here, because we want to keep the lock until the service is restarted
			return err
		}
		return restartDeployment(ctx, clientset, service)
	case "StatefulSet":
		err := lock.Lock(service.String())
		if err != nil {
			// we don't want to unlock the lock here, because we want to keep the lock until the service is restarted
			return err
		}
		return restartStatefulSet(ctx, clientset, service)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidKind, service.Kind)
	}
}

func restartDeployment(ctx context.Context, clientset *kubernetes.Clientset, service KindNamespaceName) error {
	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
	_, err := clientset.AppsV1().Deployments(service.Namespace).Patch(ctx, service.Name, types.MergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch deployment: %w", err)
	}
	return nil
}

func restartStatefulSet(ctx context.Context, clientset *kubernetes.Clientset, service KindNamespaceName) error {
	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
	_, err := clientset.AppsV1().StatefulSets(service.Namespace).Patch(ctx, service.Name, types.MergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch statefulset: %w", err)
	}
	return nil
}
