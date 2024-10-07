package k8s

import (
	"reflect"
	"testing"
)

func TestKindNamespaceNameFromString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *KindNamespaceName
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				s: "Deployment/my-namespace/my-deployment",
			},
			want: &KindNamespaceName{
				Kind:      "Deployment",
				Namespace: "my-namespace",
				Name:      "my-deployment",
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			args: args{
				s: "Deployment/my-namespace",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid kind",
			args: args{
				s: "Service/my-namespace/my-service",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KindNamespaceNameFromString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("KindNamespaceNameFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KindNamespaceNameFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKindNamespaceName_String(t *testing.T) {
	type fields struct {
		Kind      string
		Name      string
		Namespace string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "valid",
			fields: fields{
				Kind:      "Deployment",
				Namespace: "my-namespace",
				Name:      "my-deployment",
			},
			want: "Deployment/my-namespace/my-deployment",
		},
		{
			name: "empty",
			fields: fields{
				Kind:      "",
				Namespace: "",
				Name:      "",
			},
			want: "//",
		},
		{
			name: "with service kind",
			fields: fields{
				Kind:      "Service",
				Namespace: "my-namespace",
				Name:      "my-service",
			},
			want: "Service/my-namespace/my-service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := KindNamespaceName{
				Kind:      tt.fields.Kind,
				Name:      tt.fields.Name,
				Namespace: tt.fields.Namespace,
			}
			if got := s.String(); got != tt.want {
				t.Errorf("KindNamespaceName.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
