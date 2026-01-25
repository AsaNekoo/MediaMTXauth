package auth

import (
	"MediaMTXAuth/internal/services"
	"MediaMTXAuth/internal/storage/memory"
	"errors"
	"testing"
)

func TestAuth_Validate(t *testing.T) {
	type args struct {
		namespace string
		userName  string
		streamKey string
		action    string
	}

	storage := &memory.Storage{}
	nsService := services.NewNamespaceService(storage)
	userService := services.NewUserService(storage)
	auth := New(userService, nsService)
	ns, err := nsService.Create("test_namespace")

	if err != nil {
		t.Fatal(err)
	}

	u, err := userService.Create("test", "testtest", false, ns.Name)

	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "publish valid",
			args: args{
				namespace: ns.Name,
				userName:  u.Name,
				streamKey: u.StreamKey,
				action:    "publish",
			},
			wantErr: nil,
		},
		{
			name: "read valid",
			args: args{
				namespace: ns.Name,
				userName:  u.Name,
				streamKey: u.StreamKey,
				action:    "read",
			},
			wantErr: nil,
		},
		// TODO: cover other cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.Validate(tt.args.namespace, tt.args.userName, tt.args.streamKey, tt.args.action)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
