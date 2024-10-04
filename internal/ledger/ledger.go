package ledger

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/k8scope/k8s-restart-app/internal/k8s"
	"github.com/leonsteinhaeuser/observer/v2"
	"k8s.io/client-go/kubernetes"
)

type Status struct {
	Message   string        `json:"message"`
	PodStatus k8s.PodStatus `json:"pod_status"`
}

func (s *Status) send(err error, observer *observer.Observer[Status]) {
	if err != nil {
		observer.NotifyAll(Status{
			Message: err.Error(),
		})
	}
	observer.NotifyAll(*s)
}

type Ledger struct {
	client           *kubernetes.Clientset
	watchIntervalSec int

	transactionLock sync.Mutex
	transactions    map[string]*observer.Observer[Status]

	cancelCh chan struct{}
}

func New(client *kubernetes.Clientset, watchIntervalSec int) *Ledger {
	return &Ledger{
		client:           client,
		watchIntervalSec: watchIntervalSec,
		transactions:     map[string]*observer.Observer[Status]{},
		cancelCh:         make(chan struct{}),
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
				status := Status{}

				deployment, err := k8s.GetDeployment(ctx, l.client, kindNamespaceName)
				if err != nil {
					slog.Error("failed to get deployment", "error", err, "kindNamespaceName", kindNamespaceName)
					status.send(err, l.transactions[kindNamespaceName.String()])
					return
				}

				pods, err := k8s.GetPodStatus(ctx, l.client, deployment.Namespace, deployment.Spec.Selector.MatchLabels)
				if err != nil {
					slog.Error("failed to get pod status", "error", err, "kindNamespaceName", kindNamespaceName)
					status.send(err, l.transactions[kindNamespaceName.String()])
					return
				}

				podStatus := k8s.GetPodStatusFormat(pods)
				status.PodStatus = podStatus
				status.send(nil, l.transactions[kindNamespaceName.String()])
			case "StatefulSet":
				ctx, cf := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
				defer cf()
				status := Status{}

				statefulset, err := k8s.GetStatefulset(ctx, l.client, kindNamespaceName)
				if err != nil {
					slog.Error("failed to get statefulset", "error", err, "kindNamespaceName", kindNamespaceName)
					status.send(err, l.transactions[kindNamespaceName.String()])
					return
				}

				pods, err := k8s.GetPodStatus(ctx, l.client, statefulset.Namespace, statefulset.Spec.Selector.MatchLabels)
				if err != nil {
					slog.Error("failed to get pod status", "error", err, "kindNamespaceName", kindNamespaceName)
					status.send(err, l.transactions[kindNamespaceName.String()])
					return
				}

				podStatus := k8s.GetPodStatusFormat(pods)
				status.PodStatus = podStatus
				status.send(nil, l.transactions[kindNamespaceName.String()])
			default:
				slog.Error("invalid kind", "kind", kindNamespaceName.Kind)
				(&Status{}).send(fmt.Errorf("invalid kind: %s", kindNamespaceName.Kind), l.transactions[kindNamespaceName.String()])
				return
			}

		}
		time.Sleep(time.Duration(l.watchIntervalSec) * time.Second)
	}
}

// Register registers a new channel for observing the status of a given object.
// The channel will receive updates every watchIntervalSec seconds.
func (l *Ledger) Register(kindNamespaceName k8s.KindNamespaceName) (<-chan Status, observer.CancelFunc) {
	if _, ok := l.transactions[kindNamespaceName.String()]; !ok {
		l.transactionLock.Lock()
		defer l.transactionLock.Unlock()

		// create new observer
		l.transactions[kindNamespaceName.String()] = new(observer.Observer[Status])
		// start watching for changes
		go l.watch(kindNamespaceName)
	}
	return l.transactions[kindNamespaceName.String()].Subscribe()
}
