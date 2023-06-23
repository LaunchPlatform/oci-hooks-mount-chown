package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"reflect"
	"syscall"
	"testing"
)

func Test_loadSpec(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bundle")
	if err != nil {
		t.Fatal(err)
	}
	specValue := spec.Spec{
		Version: spec.Version,
		Mounts: []spec.Mount{
			{
				Destination: "/data",
				Source:      "/path/to/source",
				Options:     []string{"nodev"},
			},
		},
	}
	configData, err := json.Marshal(specValue)
	if err != nil {
		t.Fatal(err)
	}
	configPath := path.Join(tempDir, "config.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatal(err)
	}
	stateData, err := json.Marshal(spec.State{Bundle: tempDir})
	if err != nil {
		t.Fatal(err)
	}
	resultSpec := loadSpec(bytes.NewReader(stateData))
	assert.True(t, reflect.DeepEqual(resultSpec, specValue))
}

func Test_chownMountPoints(t *testing.T) {
	mountDir, err := os.MkdirTemp("", "mount")
	if err != nil {
		t.Fatal(err)
	}
	nestedFileData := []byte("MOCK_CONTENT")
	nestedFileDir := path.Join(mountDir, "nested", "dir")
	nestedFilePath := path.Join(nestedFileDir, "file.txt")
	err = os.MkdirAll(nestedFileDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(nestedFilePath, nestedFileData, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Get mount dir info
	f, err := os.Lstat(mountDir)
	if err != nil {
		t.Fatal(err)
	}

	// Get current ownership
	currentUID := int(f.Sys().(*syscall.Stat_t).Uid)
	currentGID := int(f.Sys().(*syscall.Stat_t).Gid)

	containerSpec := spec.Spec{
		Version: spec.Version,
		Mounts: []spec.Mount{
			{
				Destination: "/dev",
				Source:      "tmpfs",
				Type:        "tmpfs",
				Options: []string{
					"nosuid",
					"strictatime",
					"mode=755",
					"size=65536k",
				},
			},
			{
				Destination: mountDir,
				Source:      "/path/to/source",
				Type:        "overlay",
				Options: []string{
					"lowerdir=/path/to/lower",
					"upperdir=/path/to/upper",
					"workdir=/path/to/work",
					"private",
				},
			},
		},
	}
	archives := map[string]ChownRequest{mountDir: {MountPoint: "/data", User: currentUID, Group: currentGID, Name: "data"}}
	chownMountPoints(containerSpec, archives)
	// Change own requires privilege, so it's a bit hard to assert.
	// We set it the current uid & gid to make it easier to run for now.
}

func Test_chownMountPoint(t *testing.T) {
	mountDir, err := os.MkdirTemp("", "mount")
	if err != nil {
		t.Fatal(err)
	}
	nestedFileData := []byte("MOCK_CONTENT")
	nestedFileDir := path.Join(mountDir, "nested", "dir")
	nestedFilePath := path.Join(nestedFileDir, "file.txt")
	err = os.MkdirAll(nestedFileDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(nestedFilePath, nestedFileData, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Get mount dir info
	f, err := os.Lstat(mountDir)
	if err != nil {
		t.Fatal(err)
	}

	// Get current ownership
	currentUID := int(f.Sys().(*syscall.Stat_t).Uid)
	currentGID := int(f.Sys().(*syscall.Stat_t).Gid)

	tests := []struct {
		name    string
		args    ChownRequest
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"recursive",
			ChownRequest{MountPoint: mountDir, User: currentUID, Group: currentGID, Policy: PolicyRecursive},
			assert.NoError,
		},
		{
			"root-only",
			ChownRequest{MountPoint: mountDir, User: currentUID, Group: currentGID, Policy: PolicyRootOnly},
			assert.NoError,
		},
		{
			"not-exist-path",
			ChownRequest{MountPoint: "/path/to/non-exist", User: currentUID, Group: currentGID},
			assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, chownMountPoint(tt.args), fmt.Sprintf("chownMountPoint(%v)", tt.args))
		})
	}
}
