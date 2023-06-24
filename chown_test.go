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
			"com.launchplatform.oci-hooks.mount-chown.data.path":  "/path/to/root",
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "2000:2000",
		}}, map[string]ChownRequest{
			"/path/to/root": {Name: "data", Path: "/path/to/root", User: 2000, Group: 2000},
		},
		},
		{
			"recursive-policy", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":   "/path/to/root",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":  "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data.policy": PolicyRecursive,
		}}, map[string]ChownRequest{
			"/path/to/root": {
				Name:   "data",
				Path:   "/path/to/root",
				User:   2000,
				Group:  2000,
				Policy: PolicyRecursive,
			},
		},
		},
		{
			"root-only-policy", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":   "/path/to/root",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":  "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data.policy": PolicyRootOnly,
		}}, map[string]ChownRequest{
			"/path/to/root": {
				Name:   "data",
				Path:   "/path/to/root",
				User:   2000,
				Group:  2000,
				Policy: PolicyRootOnly,
			},
		},
		},
		{
			"multiple", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data0.path":  "/path/to/root0",
			"com.launchplatform.oci-hooks.mount-chown.data0.owner": "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data1.path":  "/path/to/root1",
			"com.launchplatform.oci-hooks.mount-chown.data1.owner": "3000:4000",
		}}, map[string]ChownRequest{
			"/path/to/root0": {Name: "data0", Path: "/path/to/root0", User: 2000, Group: 2000},
			"/path/to/root1": {Name: "data1", Path: "/path/to/root1", User: 3000, Group: 4000},
		},
		},
		{
			"invalid-key", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":    "/path/to/root",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":   "2000",
			"com.launchplatform.oci-hooks.mount-chown.data.invalid": "others",
		}}, map[string]ChownRequest{
			"/path/to/root": {Name: "data", Path: "/path/to/root", User: 2000, Group: 0},
		},
		},
		{
			"relative-path", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":  "/path/../../../../etc/passwd",
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "2000:2000",
		}}, map[string]ChownRequest{},
		},
		{
			"leading-relative-path", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":  "./path/to/root",
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "2000:2000",
		}}, map[string]ChownRequest{},
		},
		{
			"evil-relative-path", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":  "../../../etc/passwd",
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "2000:2000",
		}}, map[string]ChownRequest{},
		},
		{
			"invalid-policy", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":   "/path/to/root",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":  "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data.policy": "invalid",
		}}, map[string]ChownRequest{},
		},
		{
			"empty-path", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":  "",
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "2000:2000",
		}}, map[string]ChownRequest{},
		},
		{
			"missing-path", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "2000:2000",
		}}, map[string]ChownRequest{},
		},
		{
			"empty-owner", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":  "/path/to/root",
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "",
		}}, map[string]ChownRequest{},
		},
		{
			"missing-owner", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path": "/path/to/root",
		}}, map[string]ChownRequest{},
		},
		{
			"invalid-owner", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.path":  "/path/to/root",
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "foobar",
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
