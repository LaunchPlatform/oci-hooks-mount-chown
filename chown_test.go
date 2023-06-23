package main

import (
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
			"/path/to/mount-point": {Name: "data", MountPoint: "/path/to/mount-point", Owner: "2000:2000"},
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
				Owner:      "2000:2000",
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
				Owner:      "2000:2000",
				Policy:     PolicyRootOnly,
			},
		},
		},
		{
			"multiple", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data0.mount-point": "/path/to/mount-point0",
			"com.launchplatform.oci-hooks.mount-chown.data0.owner":       "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data1.mount-point": "/path/to/mount-point1",
			"com.launchplatform.oci-hooks.mount-chown.data1.owner":       "3000:3000",
		}}, map[string]ChownRequest{
			"/path/to/mount-point0": {Name: "data0", MountPoint: "/path/to/mount-point0", Owner: "2000:2000"},
			"/path/to/mount-point1": {Name: "data1", MountPoint: "/path/to/mount-point1", Owner: "3000:3000"},
		},
		},
		{
			"invalid-key", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data.invalid":     "others",
		}}, map[string]ChownRequest{"/path/to/mount-point": {Name: "data", MountPoint: "/path/to/mount-point", Owner: "2000:2000"}},
		},
		{
			"invalid-policy", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.mount-chown.data.owner":       "2000:2000",
			"com.launchplatform.oci-hooks.mount-chown.data.policy":      "invalid",
		}}, map[string]ChownRequest{},
		},
		{
			"empty-owner", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.owner": "/path/to/archive\"2000:2000\",-to",
		}}, map[string]ChownRequest{},
		},
		{
			"empty-mount-point", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.mount-chown.data.mount-point": "/path/to/mount-point",
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
