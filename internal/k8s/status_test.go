package k8s

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func TestGetDeployment(t *testing.T) {
	type args struct {
		ctx     context.Context
		client  *kubernetes.Clientset
		service KindNamespaceName
	}
	tests := []struct {
		name    string
		args    args
		want    *appsv1.Deployment
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDeployment(tt.args.ctx, tt.args.client, tt.args.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDeployment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStatefulset(t *testing.T) {
	type args struct {
		ctx     context.Context
		client  *kubernetes.Clientset
		service KindNamespaceName
	}
	tests := []struct {
		name    string
		args    args
		want    *appsv1.StatefulSet
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStatefulset(tt.args.ctx, tt.args.client, tt.args.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatefulset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStatefulset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPodStatus(t *testing.T) {
	type args struct {
		ctx       context.Context
		client    *kubernetes.Clientset
		namespace string
		selector  map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]corev1.PodStatus
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPodStatus(tt.args.ctx, tt.args.client, tt.args.namespace, tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPodStatusFormat(t *testing.T) {
	type args struct {
		pod map[string]corev1.PodStatus
	}
	tests := []struct {
		name string
		args args
		want PodStatus
	}{
		{
			name: "one pod",
			args: args{
				pod: map[string]corev1.PodStatus{
					"pod1": {
						Phase: corev1.PodRunning,
					},
				},
			},
			want: PodStatus{
				corev1.PodRunning: 1,
			},
		},
		{
			name: "two pods with different status",
			args: args{
				pod: map[string]corev1.PodStatus{
					"pod1": {
						Phase: corev1.PodRunning,
					},
					"pod2": {
						Phase: corev1.PodPending,
					},
				},
			},
			want: PodStatus{
				corev1.PodRunning: 1,
				corev1.PodPending: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPodStatusFormat(tt.args.pod)
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Errorf("GetPodStatusFormat() mismatch (-got +want):\n%s", diff)
				return
			}
		})
	}
}
