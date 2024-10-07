package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/k8scope/k8s-restart-app/internal/lock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type KindNamespaceName struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
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
		return fmt.Errorf("invalid service kind: %s", service.Kind)
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
