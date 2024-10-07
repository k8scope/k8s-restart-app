package lock

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
		m    map[string]time.Time
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
				m:    map[string]time.Time{},
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
				m: map[string]time.Time{
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
				m:    map[string]time.Time{"test": {}},
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
		m    map[string]time.Time
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
				m:    map[string]time.Time{"test": {}},
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
				m:    map[string]time.Time{},
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
				m: map[string]time.Time{
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
		m    map[string]time.Time
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
				m:    map[string]time.Time{"test": {}},
			},
			want: []string{"test"},
		},
		{
			name: "get locks with two locks",
			fields: fields{
				rwmu: sync.RWMutex{},
				m: map[string]time.Time{
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

func TestInMem_IsLocked(t *testing.T) {
	type fields struct {
		rwmu sync.RWMutex
		m    map[string]time.Time
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "is locked",
			fields: fields{
				rwmu: sync.RWMutex{},
				m:    map[string]time.Time{"test": {}},
			},
			args: args{
				name: "test",
			},
			want: true,
		},
		{
			name: "is not locked",
			fields: fields{
				rwmu: sync.RWMutex{},
				m:    map[string]time.Time{},
			},
			args: args{
				name: "test",
			},
			want: false,
		},
		{
			name: "is not locked with other lock",
			fields: fields{
				rwmu: sync.RWMutex{},
				m: map[string]time.Time{
					"other": {},
				},
			},
			args: args{
				name: "test",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &InMem{
				rwmu: tt.fields.rwmu,
				m:    tt.fields.m,
			}
			if got := l.IsLocked(tt.args.name); got != tt.want {
				t.Errorf("InMem.IsLocked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMem_ForceUnlockAfter(t *testing.T) {
	type fields struct {
		rwmu sync.RWMutex
		m    map[string]time.Time
	}
	type args struct {
		duration time.Duration
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantLock []string
		wantErr  bool
	}{
		{
			name: "force unlock for locks which are older than 1 second",
			fields: fields{
				rwmu: sync.RWMutex{},
				m: map[string]time.Time{
					"test": time.Now().Add(-time.Hour),
				},
			},
			args: args{
				duration: 1 * time.Second,
			},
			wantLock: []string{},
			wantErr:  false,
		},
		{
			name: "do not unlock",
			fields: fields{
				rwmu: sync.RWMutex{},
				m: map[string]time.Time{
					"test":  time.Now().Add(time.Hour),
					"other": time.Now().Add(time.Hour),
				},
			},
			args: args{
				duration: 1 * time.Second,
			},
			wantLock: []string{
				"test",
				"other",
			},
			wantErr: false,
		},
		{
			name: "force unlock after 5 seconds with two locks",
			fields: fields{
				rwmu: sync.RWMutex{},
				m: map[string]time.Time{
					"test":  time.Now().Add(-time.Hour),
					"other": time.Now().Add(time.Hour),
				},
			},
			args: args{
				duration: 5 * time.Second,
			},
			wantLock: []string{"other"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &InMem{
				rwmu: tt.fields.rwmu,
				m:    tt.fields.m,
			}
			l.ForceUnlockAfter(tt.args.duration)
			// we need to wait for the force unlock to happen, so we can check the locks
			time.Sleep(1 * time.Second)

			locks := l.GetLocks()
			diff := cmp.Diff(locks, tt.wantLock)
			if diff != "" {
				t.Errorf("unexpected locks (-got +want):\n%s", diff)
				return
			}

		})
	}
}
