package lock

import (
	"reflect"
	"testing"
	"time"
)

func TestNewLock(t *testing.T) {
	type args struct {
		locker              Locker
		forceUnlockAfterSec int
	}
	tests := []struct {
		name string
		args args
		want *Lock
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLock(tt.args.locker, tt.args.forceUnlockAfterSec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLock_Lock(t *testing.T) {
	type fields struct {
		locker Locker
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
				locker: &InMem{
					m: map[string]time.Time{},
				},
			},
			args: args{
				name: "test/test/test",
			},
			wantErr: false,
		},
		{
			name: "lock with other lock",
			fields: fields{
				locker: &InMem{
					m: map[string]time.Time{
						"other": time.Now(),
					},
				},
			},
			args: args{
				name: "test/test/test",
			},
			wantErr: false,
		},
		{
			name: "already locked",
			fields: fields{
				locker: &InMem{
					m: map[string]time.Time{
						"test/test/test": time.Now(),
					},
				},
			},
			args: args{
				name: "test/test/test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lock{
				locker: tt.fields.locker,
			}
			if err := l.Lock(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Lock.Lock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLock_IsLocked(t *testing.T) {
	type fields struct {
		locker Locker
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
			name: "not locked",
			fields: fields{
				locker: &InMem{
					m: map[string]time.Time{},
				},
			},
			args: args{
				name: "test/test/test",
			},
			want: false,
		},
		{
			name: "locked",
			fields: fields{
				locker: &InMem{
					m: map[string]time.Time{
						"test/test/test": time.Now(),
					},
				},
			},
			args: args{
				name: "test/test/test",
			},
			want: true,
		},
		{
			name: "locked with other lock",
			fields: fields{
				locker: &InMem{
					m: map[string]time.Time{
						"other":          time.Now(),
						"test/test/test": time.Now(),
					},
				},
			},
			args: args{
				name: "test/test/test",
			},
			want: true,
		},
		{
			name: "not locked with two locks",
			fields: fields{
				locker: &InMem{
					m: map[string]time.Time{
						"other":          time.Now(),
						"test/test/test": time.Now(),
					},
				},
			},
			args: args{
				name: "test2/test2/test2",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lock{
				locker: tt.fields.locker,
			}
			if got := l.IsLocked(tt.args.name); got != tt.want {
				t.Errorf("Lock.IsLocked() = %v, want %v", got, tt.want)
			}
		})
	}
}
