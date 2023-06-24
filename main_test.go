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

func Test_chownRequests(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "root")
	if err != nil {
		t.Fatal(err)
	}
	mountDir := path.Join(rootDir, "data")
	err = os.MkdirAll(mountDir, 0755)
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

	requests := map[string]ChownRequest{mountDir: {Path: "/data", User: currentUID, Group: currentGID, Name: "data"}}
	chownRequests(rootDir, requests)
	// Change own requires privilege, so it's a bit hard to assert.
	// We set it the current uid & gid to make it easier to run for now.
}

func Test_doChownRequest(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "root")
	if err != nil {
		t.Fatal(err)
	}
	mountDir := path.Join(rootDir, "data")
	err = os.MkdirAll(mountDir, 0755)
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
			ChownRequest{Path: "/data", User: currentUID, Group: currentGID, Policy: PolicyRecursive},
			assert.NoError,
		},
		{
			"root-only",
			ChownRequest{Path: "/data", User: currentUID, Group: currentGID, Policy: PolicyRootOnly},
			assert.NoError,
		},
		{
			"mode-only",
			ChownRequest{Path: "/data", User: -1, Group: -1, Mode: 0755},
			assert.NoError,
		},
		{
			"not-exist-path",
			ChownRequest{Path: "/path/to/non-exist", User: currentUID, Group: currentGID},
			assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, doChownRequest(rootDir, tt.args), fmt.Sprintf("doChownRequest(%v)", tt.args))
		})
	}
}

func Test_doChownRequestForMode(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "root")
	if err != nil {
		t.Fatal(err)
	}
	mountDir := path.Join(rootDir, "data")
	err = os.MkdirAll(mountDir, 0777)
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

	request := ChownRequest{Path: "/data", User: currentUID, Group: currentGID, Policy: PolicyRootOnly, Mode: 0700}
	err = doChownRequest(rootDir, request)
	if err != nil {
		t.Fatal(err)
	}

	f, err = os.Lstat(mountDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, os.FileMode(0700), f.Mode().Perm())
}
