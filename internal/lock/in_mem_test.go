package lock

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewInMem(t *testing.T) {
	tests := []struct {
		name string
		want *InMem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInMem(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInMem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMem_Lock(t *testing.T) {
	type fields struct {
		rwmu sync.RWMutex
		m    map[string]struct{}
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "lock",
			fields: fields{
				rwmu: sync.RWMutex{},
				m:    map[string]struct{}{},
			},
			args: args{
				name: "test",
			},
			wantErr: false,
		},
		{
			name: "lock with other lock",
			fields: fields{
				rwmu: sync.RWMutex{},
				m: map[string]struct{}{
					"other": {},
				},
			},
			args: args{
				name: "test",
			},
			wantErr: false,
		},
		{
			name: "already locked",
			fields: fields{
				rwmu: sync.RWMutex{},
				m:    map[string]struct{}{"test": {}},
			},
			args: args{
				name: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &InMem{
				rwmu: tt.fields.rwmu,
				m:    tt.fields.m,
			}
			if err := l.Lock(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("InMem.Lock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInMem_Unlock(t *testing.T) {
	type fields struct {
		rwmu sync.RWMutex
		m    map[string]struct{}
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "unlock",
			fields: fields{
				rwmu: sync.RWMutex{},
				m:    map[string]struct{}{"test": {}},
			},
			args: args{
				name: "test",
			},
			wantErr: false,
		},
		{
			name: "not locked",
			fields: fields{
				rwmu: sync.RWMutex{},
				m:    map[string]struct{}{},
			},
			args: args{
				name: "test",
			},
			wantErr: true,
		},
		{
			name: "not locked with other lock",
			fields: fields{
				rwmu: sync.RWMutex{},
				m: map[string]struct{}{
					"other": {},
				},
			},
			args: args{
				name: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &InMem{
				rwmu: tt.fields.rwmu,
				m:    tt.fields.m,
			}
			if err := l.Unlock(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("InMem.Unlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInMem_GetLocks(t *testing.T) {
	type fields struct {
		rwmu sync.RWMutex
		m    map[string]struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "get locks",
			fields: fields{
				rwmu: sync.RWMutex{},
				m:    map[string]struct{}{"test": {}},
			},
			want: []string{"test"},
		},
		{
			name: "get locks with two locks",
			fields: fields{
				rwmu: sync.RWMutex{},
				m: map[string]struct{}{
					"test":  {},
					"other": {},
				},
			},
			want: []string{"test", "other"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &InMem{
				rwmu: tt.fields.rwmu,
				m:    tt.fields.m,
			}
			if got := l.GetLocks(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InMem.GetLocks() = %v, want %v", got, tt.want)
			}
		})
	}
}
