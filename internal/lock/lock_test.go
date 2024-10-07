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
