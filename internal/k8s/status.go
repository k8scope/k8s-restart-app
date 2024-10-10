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

// GetPods returns a list of pods, identified by the selectors
func GetPods(ctx context.Context, client *kubernetes.Clientset, namespace string, selector map[string]string) ([]corev1.Pod, error) {
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(labels.Set(selector)))})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

type PodStatus map[corev1.PodPhase]int

// PodStatuses checks if the restart for the deployment or statefulset is complete
// It returns a map of pod status and the number of pods with that status
// as well as a boolean indicating if the restart is complete.
// If the passed pod list is empty, we treat the restart as complete
// If from the passed pod list, the first pod doesn't have an owner reference, we treat the restart as complete
// If from the passed pod list, the first pod doesn't have an owner reference of kind replicaset or statefulset, we treat the restart as complete
// In all other cases, we check the status of the pod and add it to the map
// If there are no pods, we treat the restart as complete
func PodStatuses(pods []corev1.Pod) (PodStatus, bool) {
	if len(pods) < 1 {
		// if there are no pods, restart won't change anything
		// therefore we treat it as complete
		return nil, true
	}

	if len(pods[0].OwnerReferences) < 1 {
		// in case of a pod was created without owner reference,
		// we can't determine if the restart for the deployment or statefulset is complete
		// so we return true to prevent deadlocks
		return nil, true
	}

	// get the first owner reference with kind replicaset or statefulset
	podOwnerRef := firstOwnerRefWithKindReplicaSetStatefulSet(pods[0].OwnerReferences)
	if podOwnerRef == nil {
		// if the pod doesn't have an owner reference of kind replicaset or statefulset
		// then the pod is not part of the deployment (replicaset) or statefulset
		return nil, true
	}

	podStatus := PodStatus{}
	for _, pod := range pods {
		if !compareOwnerRefs(pod.OwnerReferences, *podOwnerRef) {
			// pod is not owned by the same replicasets or statefulset
			// so we skip it
			continue
		}
		podStatus[pod.Status.Phase]++
	}
	_, ok := podStatus[corev1.PodRunning]
	if len(podStatus) > 1 || !ok {
		// if there are multiple pod statuses or no pod is in running state
		// then the restart is not complete
		return podStatus, false
	}
	return podStatus, true
}

// compareOwnerRefs compares the owner references and returns true if the owner reference is in the list
func compareOwnerRefs(a []metav1.OwnerReference, b metav1.OwnerReference) bool {
	for _, ref := range a {
		if ref == b {
			return true
		}
	}
	return false
}

// firstOwnerRefWithKindReplicaSetStatefulSet returns the first owner reference with kind replicaset or statefulset
func firstOwnerRefWithKindReplicaSetStatefulSet(a []metav1.OwnerReference) *metav1.OwnerReference {
	for _, ref := range a {
		if ref.Kind == "ReplicaSet" || ref.Kind == "StatefulSet" {
			return &ref
		}
	}
	return nil
}
