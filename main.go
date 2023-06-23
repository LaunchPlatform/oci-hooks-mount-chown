package main

import (
	"encoding/json"
	"fmt"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
	"strings"
)

const (
	defaultLogLevel = "info"
)

var (
	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	logLevel  = defaultLogLevel
)

func loadSpec(stateInput io.Reader) spec.Spec {
	var state spec.State
	err := json.NewDecoder(stateInput).Decode(&state)
	if err != nil {
		log.Fatalf("Failed to parse stdin with error %s", err)
	}
	configPath := path.Join(state.Bundle, "config.json")
	jsonFile, err := os.Open(configPath)
	defer jsonFile.Close()
	if err != nil {
		log.Fatalf("Failed to open OCI spec file %s with error %s", configPath, err)
	}
	var containerSpec spec.Spec
	err = json.NewDecoder(jsonFile).Decode(&containerSpec)
	if err != nil {
		log.Fatalf("Failed to parse OCI spec JSON file %s with error %s", configPath, err)
	}
	return containerSpec
}

func chownMountPoints(containerSpec spec.Spec, mountPointRequests map[string]ChownRequest) {
	for _, mount := range containerSpec.Mounts {
		_, ok := mountPointRequests[mount.Destination]
		if !ok {
			log.Tracef("Cannot find mount point %s to chown, skip", mount.Destination)
			continue
		}
		// TODO:
	}
}

func run() {
	containerSpec := loadSpec(os.Stdin)
	destRequests := parseChownRequests(containerSpec.Annotations)
	requestsJson, err := json.Marshal(destRequests)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Parsed requests: %s", string(requestsJson))
	chownMountPoints(containerSpec, destRequests)
	log.Infof("Done")
}

func setupLogLevel() {
	var found = false
	for _, level := range LogLevels {
		if level == strings.ToLower(logLevel) {
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "Log Level %q is not supported, choose from: %s\n", logLevel, strings.Join(LogLevels, ", "))
		os.Exit(1)
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	log.SetLevel(level)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:     "archive_overlay [options]",
		Short:   "Invoked as a poststop OCI-hooks to archive upperdir of specific overlay mount",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			setupLogLevel()
			log.Infof("Run archive_overlay %s", Version)
			run()
		},
	}
	pFlags := rootCmd.PersistentFlags()
	logLevelFlagName := "log-level"
	pFlags.StringVar(
		&logLevel,
		logLevelFlagName,
		logLevel,
		fmt.Sprintf("Log messages above specified level (%s)", strings.Join(LogLevels, ", ")),
	)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
