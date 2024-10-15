package k8s

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestGetPods(t *testing.T) {
	type args struct {
		ctx       context.Context
		client    *kubernetes.Clientset
		namespace string
		selector  map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    []corev1.Pod
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPods(tt.args.ctx, tt.args.client, tt.args.namespace, tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPods() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodStatuses(t *testing.T) {
	type args struct {
		pods []corev1.Pod
	}
	tests := []struct {
		name  string
		args  args
		want  PodStatus
		want1 bool
	}{

		{
			name: "no pods",
			args: args{
				pods: []corev1.Pod{},
			},
			want:  nil,
			want1: true,
		},
		{
			name: "pod without owner reference",
			args: args{
				pods: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{},
							Name:            "abc",
							Namespace:       "default",
						},
					},
				},
			},
			want:  nil,
			want1: true,
		},
		{
			name: "pod with owner reference of unknown OwnerReference",
			args: args{
				pods: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       "Unknown",
									Name:       "abc",
									APIVersion: "v1",
								},
							},
							Name:      "abc",
							Namespace: "default",
						},
					},
				},
			},
			want:  nil,
			want1: true,
		},
		{
			name: "pod with owner reference of kind ReplicaSet and undefined status",
			args: args{
				pods: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       "ReplicaSet",
									Name:       "abc",
									APIVersion: "v1",
								},
							},
						},
					},
				},
			},
			want: PodStatus{
				"": 1,
			},
			want1: false,
		},
		{
			name: "pod with owner reference of kind ReplicaSet and Error status",
			args: args{
				pods: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       "ReplicaSet",
									Name:       "abc",
									APIVersion: "v1",
								},
							},
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodFailed,
						},
					},
				},
			},
			want: PodStatus{
				corev1.PodFailed: 1,
			},
			want1: false,
		},
		{
			name: "pod with owner reference of kind ReplicaSet and status Running",
			args: args{
				pods: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       "ReplicaSet",
									Name:       "abc",
									APIVersion: "v1",
								},
							},
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
				},
			},
			want: PodStatus{
				corev1.PodRunning: 1,
			},
			want1: true,
		},
		{
			name: "two pods with owner reference and status Running",
			args: args{
				pods: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       "ReplicaSet",
									Name:       "abc",
									APIVersion: "v1",
								},
							},
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       "ReplicaSet",
									Name:       "abc",
									APIVersion: "v1",
								},
							},
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
				},
			},
			want: PodStatus{
				corev1.PodRunning: 2,
			},
			want1: true,
		},
		{
			name: "two pods with different owner reference and status Running",
			args: args{
				pods: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       "ReplicaSet",
									Name:       "abc",
									APIVersion: "v1",
								},
							},
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind:       "ReplicaSet",
									Name:       "abc1",
									APIVersion: "v1",
								},
							},
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
				},
			},
			want: PodStatus{
				corev1.PodRunning: 1,
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, isRestarted := PodStatuses(tt.args.pods)

			t.Logf("PodStatuses() status = %v", status)

			diff := cmp.Diff(status, tt.want)
			if diff != "" {
				t.Errorf("PodStatuses() mismatch (-want +got):\n%s", diff)
				return
			}

			if isRestarted != tt.want1 {
				t.Errorf("PodStatuses() isRestarted = %v, want %v", isRestarted, tt.want1)
				return
			}

		})
	}
}

func Test_compareOwnerRefs(t *testing.T) {
	type args struct {
		a []metav1.OwnerReference
		b metav1.OwnerReference
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty owner references",
			args: args{
				a: []metav1.OwnerReference{},
				b: metav1.OwnerReference{},
			},
			want: false,
		},
		{
			name: "owner references with different APIVersion",
			args: args{
				a: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "ReplicaSet",
						Name:       "abc",
					},
				},
				b: metav1.OwnerReference{
					APIVersion: "v2",
					Kind:       "ReplicaSet",
					Name:       "abc",
				},
			},
			want: false,
		},
		{
			name: "owner references with different Kind",
			args: args{
				a: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "ReplicaSet",
						Name:       "abc",
					},
				},
				b: metav1.OwnerReference{
					APIVersion: "v1",
					Kind:       "ReplicaSet1",
					Name:       "abc",
				},
			},
			want: false,
		},
		{
			name: "owner references with different Name",
			args: args{
				a: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "ReplicaSet",
						Name:       "abc",
					},
				},
				b: metav1.OwnerReference{
					APIVersion: "v1",
					Kind:       "ReplicaSet",
					Name:       "abc1",
				},
			},
			want: false,
		},
		{
			name: "owner references with same APIVersion, Kind and Name",
			args: args{
				a: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "ReplicaSet",
						Name:       "abc",
					},
				},
				b: metav1.OwnerReference{
					APIVersion: "v1",
					Kind:       "ReplicaSet",
					Name:       "abc",
				},
			},
			want: true,
		},
		{
			name: "two owner references, first not matching, second matching",
			args: args{
				a: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "ReplicaSet",
						Name:       "abc",
					},
					{
						APIVersion: "v1",
						Kind:       "ReplicaSet",
						Name:       "abc1",
					},
				},
				b: metav1.OwnerReference{
					APIVersion: "v1",
					Kind:       "ReplicaSet",
					Name:       "abc",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareOwnerRefs(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("compareOwnerRefs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_firstOwnerRefWithKindReplicaSetStatefulSet(t *testing.T) {
	type args struct {
		a []metav1.OwnerReference
	}
	tests := []struct {
		name string
		args args
		want *metav1.OwnerReference
	}{
		{
			name: "empty owner references",
			args: args{
				a: []metav1.OwnerReference{},
			},
			want: nil,
		},
		{
			name: "owner references with kind ReplicaSet",
			args: args{
				a: []metav1.OwnerReference{
					{
						Kind: "ReplicaSet",
					},
				},
			},
			want: &metav1.OwnerReference{
				Kind: "ReplicaSet",
			},
		},
		{
			name: "owner references with kind StatefulSet",
			args: args{
				a: []metav1.OwnerReference{
					{
						Kind: "StatefulSet",
					},
				},
			},
			want: &metav1.OwnerReference{
				Kind: "StatefulSet",
			},
		},
		{
			name: "owner references with kind ReplicaSet and StatefulSet",
			args: args{
				a: []metav1.OwnerReference{
					{
						Kind: "ReplicaSet",
					},
					{
						Kind: "StatefulSet",
					},
				},
			},
			want: &metav1.OwnerReference{
				Kind: "ReplicaSet",
			},
		},
		{
			name: "owner references with kind StatefulSet and ReplicaSet",
			args: args{
				a: []metav1.OwnerReference{
					{
						Kind: "StatefulSet",
					},
					{
						Kind: "ReplicaSet",
					},
				},
			},
			want: &metav1.OwnerReference{
				Kind: "StatefulSet",
			},
		},
		{
			name: "owner references with unsupported kind",
			args: args{
				a: []metav1.OwnerReference{
					{
						Kind: "Unsupported",
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstOwnerRefWithKindReplicaSetStatefulSet(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("firstOwnerRefWithKindReplicaSetStatefulSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
