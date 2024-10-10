package k8s

import (
	"context"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
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
