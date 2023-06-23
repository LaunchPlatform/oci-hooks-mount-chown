package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_parseChownRequests(t *testing.T) {
	type args struct {
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]ChownRequest
	}{
		{"empty", args{annotations: map[string]string{"foo": "bar"}}, map[string]ChownRequest{}},
		{
			"one", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "2000:2000",
		}}, map[string]ChownRequest{
			"/path/to/mount-point": {Name: "data", MountPoint: "/path/to/mount-point", User: 2000, Group: 2000},
		},
		},
		{
			"recursive-policy", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data.policy":      PolicyRecursive,
		}}, map[string]ChownRequest{
			"/path/to/mount-point": {
				Name:       "data",
				MountPoint: "/path/to/mount-point",
				User:       2000,
				Group:      2000,
				Policy:     PolicyRecursive,
			},
		},
		},
		{
			"root-only-policy", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data.policy":      PolicyRootOnly,
		}}, map[string]ChownRequest{
			"/path/to/mount-point": {
				Name:       "data",
				MountPoint: "/path/to/mount-point",
				User:       2000,
				Group:      2000,
				Policy:     PolicyRootOnly,
			},
		},
		},
		{
			"multiple", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data0.mount-point": "/path/to/mount-point0",
			"com.launchplatform.oci-hooks.mount-chown.data0.owner":       "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data1.mount-point": "/path/to/mount-point1",
			"com.launchplatform.oci-hooks.mount-chown.data1.owner":       "3000:4000",
		}}, map[string]ChownRequest{
			"/path/to/mount-point0": {Name: "data0", MountPoint: "/path/to/mount-point0", User: 2000, Group: 2000},
			"/path/to/mount-point1": {Name: "data1", MountPoint: "/path/to/mount-point1", User: 3000, Group: 4000},
		},
		},
		{
			"invalid-key", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "2000",
			"com.launchplatform.oci-hooks.mount-chown.data.invalid":     "others",
		}}, map[string]ChownRequest{
			"/path/to/mount-point": {Name: "data", MountPoint: "/path/to/mount-point", User: 2000, Group: 0},
		},
		},
		{
			"invalid-policy", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data.policy":      "invalid",
		}}, map[string]ChownRequest{},
		},
		{
			"empty-mount-point", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "2000:2000",
		}}, map[string]ChownRequest{},
		},
		{
			"missing-mount-point", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "2000:2000",
		}}, map[string]ChownRequest{},
		},
		{
			"empty-owner", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "",
		}}, map[string]ChownRequest{},
		},
		{
			"missing-owner", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
		}}, map[string]ChownRequest{},
		},
		{
			"invalid-owner", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "foobar",
		}}, map[string]ChownRequest{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseChownRequests(tt.args.annotations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseChownRequests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseOwner(t *testing.T) {
	type args struct {
		owner string
	}
	tests := []struct {
		name    string
		args    args
		uid     int
		gid     int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"only-user", args{"2000"}, 2000, 0, assert.NoError,
		},
		{
			"both", args{"2000:3000"}, 2000, 3000, assert.NoError,
		},
		{
			"empty", args{""}, 0, 0, assert.Error,
		},
		{
			"more-than-two-parts", args{"1:2:3"}, 0, 0, assert.Error,
		},
		{
			"non-int-user", args{"user"}, 0, 0, assert.Error,
		},
		{
			"non-int-both", args{"user:group"}, 0, 0, assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseOwner(tt.args.owner)
			if !tt.wantErr(t, err, fmt.Sprintf("parseOwner(%v)", tt.args.owner)) {
				return
			}
			assert.Equalf(t, tt.uid, got, "parseOwner(%v)", tt.args.owner)
			assert.Equalf(t, tt.gid, got1, "parseOwner(%v)", tt.args.owner)
		})
	}
}
