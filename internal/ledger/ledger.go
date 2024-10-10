package ledger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/k8scope/k8s-restart-app/internal/k8s"
	"github.com/k8scope/k8s-restart-app/internal/lock"
	"github.com/leonsteinhaeuser/observer/v2"
	"k8s.io/client-go/kubernetes"
)

type ObjectStatus struct {
	KindNamespaceName k8s.KindNamespaceName `json:"kind_namespace_name"`
	Status            Status                `json:"status"`
	IsLocked          bool                  `json:"is_locked"`
}

func (o *ObjectStatus) send(err error, observer *observer.Observer[ObjectStatus]) {
	if err != nil {
		observer.NotifyAll(ObjectStatus{
			KindNamespaceName: o.KindNamespaceName,
			Status: Status{
				Message: err.Error(),
			},
		})
	}
	observer.NotifyAll(*o)
}

type Status struct {
	Message     string        `json:"message"`
	PodStatus   k8s.PodStatus `json:"pod_status"`
	LastRestart string        `json:"last_restart"`
}

type Ledger struct {
	client           *kubernetes.Clientset
	watchIntervalSec int

	transactionLock sync.Mutex
	watchedObjects  map[string]struct{}
	transactionsCh  *observer.Observer[ObjectStatus]

	lock *lock.Lock

	cancelCh chan struct{}
}

func New(client *kubernetes.Clientset, lock *lock.Lock, watchIntervalSec int) *Ledger {
	return &Ledger{
		client:           client,
		watchIntervalSec: watchIntervalSec,
		watchedObjects:   make(map[string]struct{}),
		transactionLock:  sync.Mutex{},
		transactionsCh:   new(observer.Observer[ObjectStatus]),
		cancelCh:         make(chan struct{}),
		lock:             lock,
	}
}

func (l *Ledger) Close() {
	// signal all watchers to stop
	l.cancelCh <- struct{}{}
}

func (l *Ledger) watch(kindNamespaceName k8s.KindNamespaceName) {
	for {
		select {
		case <-l.cancelCh:
			slog.Warn("ledger watch cancelled", "kindNamespaceName", kindNamespaceName)
			return
		default:
			switch kindNamespaceName.Kind {
			case "Deployment":
				ctx, cf := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
				defer cf()
				objsts := ObjectStatus{
					KindNamespaceName: kindNamespaceName,
					Status: Status{
						Message:     "",
						PodStatus:   k8s.PodStatus{},
						LastRestart: "",
					},
					IsLocked: true,
				}

				deployment, err := k8s.GetDeployment(ctx, l.client, kindNamespaceName)
				if err != nil {
					slog.Error("failed to get deployment", "error", err, "kindNamespaceName", kindNamespaceName)
					objsts.send(err, l.transactionsCh)
					return
				}

				pods, err := k8s.GetPods(ctx, l.client, deployment.Namespace, deployment.Spec.Selector.MatchLabels)
				if err != nil {
					slog.Error("failed to gets by label selector", "error", err, "kindNamespaceName", kindNamespaceName)
					objsts.send(err, l.transactionsCh)
					return
				}

				status, isRestarted := k8s.PodStatuses(pods)
				if isRestarted {
					err := l.lock.Unlock(kindNamespaceName.String())
					if !errors.Is(err, lock.ErrResourceNotLocked) {
						slog.Error("failed to unlock resource", "error", err, "kindNamespaceName", kindNamespaceName)
						objsts.send(err, l.transactionsCh)
						return
					}
					objsts.IsLocked = false
				}
				objsts.Status.PodStatus = status
				objsts.Status.LastRestart = deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"]
				objsts.send(nil, l.transactionsCh)
			case "StatefulSet":
				ctx, cf := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
				defer cf()
				objsts := ObjectStatus{
					KindNamespaceName: kindNamespaceName,
					Status: Status{
						Message:     "",
						PodStatus:   k8s.PodStatus{},
						LastRestart: "",
					},
					IsLocked: true,
				}

				statefulset, err := k8s.GetStatefulset(ctx, l.client, kindNamespaceName)
				if err != nil {
					slog.Error("failed to get statefulset", "error", err, "kindNamespaceName", kindNamespaceName)
					objsts.send(err, l.transactionsCh)
					return
				}

				pods, err := k8s.GetPods(ctx, l.client, statefulset.Namespace, statefulset.Spec.Selector.MatchLabels)
				if err != nil {
					slog.Error("failed to gets by label selector", "error", err, "kindNamespaceName", kindNamespaceName)
					objsts.send(err, l.transactionsCh)
					return
				}

				status, isRestarted := k8s.PodStatuses(pods)
				if isRestarted {
					err := l.lock.Unlock(kindNamespaceName.String())
					if !errors.Is(err, lock.ErrResourceNotLocked) {
						slog.Error("failed to unlock resource", "error", err, "kindNamespaceName", kindNamespaceName)
						objsts.send(err, l.transactionsCh)
						return
					}
					objsts.IsLocked = false
				}
				objsts.Status.PodStatus = status
				objsts.Status.LastRestart = statefulset.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"]
				objsts.send(nil, l.transactionsCh)
			default:
				slog.Error("invalid kind", "kind", kindNamespaceName.Kind)
				objsts := ObjectStatus{
					KindNamespaceName: kindNamespaceName,
					Status:            Status{},
				}

				objsts.send(fmt.Errorf("invalid kind: %s", kindNamespaceName.Kind), l.transactionsCh)
				return
			}

		}
		time.Sleep(time.Duration(l.watchIntervalSec) * time.Second)
	}
}

// Register registers a new channel for observing the status of all objects.
// The channel will receive updates every watchIntervalSec seconds.
func (l *Ledger) Register() (<-chan ObjectStatus, observer.CancelFunc) {
	return l.transactionsCh.Subscribe()
}

// Watch starts watching the object with the given kindNamespaceName.
// The status of the object will be sent to all registered channels every watchIntervalSec seconds.
func (l *Ledger) Watch(kindNamespaceName k8s.KindNamespaceName) {
	l.transactionLock.Lock()
	defer l.transactionLock.Unlock()

	if _, ok := l.watchedObjects[kindNamespaceName.String()]; ok {
		slog.Warn("object already watched", "kindNamespaceName", kindNamespaceName)
		return
	}
	l.watchedObjects[kindNamespaceName.String()] = struct{}{}
	go l.watch(kindNamespaceName)
}
