package k8s

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func GetDeployment(ctx context.Context, client *kubernetes.Clientset, service KindNamespaceName) (*appsv1.Deployment, error) {
	ctx2, cf := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cf()
	deployment, err := client.AppsV1().Deployments(service.Namespace).Get(ctx2, service.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func GetStatefulset(ctx context.Context, client *kubernetes.Clientset, service KindNamespaceName) (*appsv1.StatefulSet, error) {
	ctx2, cf := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cf()
	statefulset, err := client.AppsV1().StatefulSets(service.Namespace).Get(ctx2, service.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return statefulset, nil
}

func GetPodStatus(ctx context.Context, client *kubernetes.Clientset, namespace string, selector map[string]string) (map[string]corev1.PodStatus, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(labels.Set(selector)))})
	if err != nil {
		return nil, err
	}
	podStatus := map[string]corev1.PodStatus{}
	for _, pod := range pods.Items {
		podStatus[pod.Name] = pod.Status
	}
	return podStatus, nil
}

func PodsAllHealthy(pods map[string]corev1.PodStatus) bool {
	for _, pod := range pods {
		if pod.Phase != corev1.PodRunning {
			return false
		}
	}
	return true
}

type PodStatus map[corev1.PodPhase]int

func GetPodStatusFormat(pod map[string]corev1.PodStatus) PodStatus {
	status := PodStatus{}
	for _, s := range pod {
		status[s.Phase]++
	}
	return status
}
