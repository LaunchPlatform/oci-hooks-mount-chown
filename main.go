package main

import (
	"encoding/json"
	"fmt"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	defaultLogLevel = "info"
)

var (
	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	logLevel  = defaultLogLevel
)

func loadState(stateInput io.Reader) spec.State {
	var state spec.State
	err := json.NewDecoder(stateInput).Decode(&state)
	if err != nil {
		log.Fatalf("Failed to parse stdin with error %s", err)
	}
	return state
}

func chownFile(name string, path string, file os.FileInfo, uid int, gid int) {
	currentUID := int(file.Sys().(*syscall.Stat_t).Uid)
	currentGID := int(file.Sys().(*syscall.Stat_t).Gid)
	if uid == currentUID && gid == currentGID {
		log.Infof("The same UID and GID of %s for %s found, skip", path, name)
		return
	}
	err := os.Lchown(path, uid, gid)
	if err != nil {
		log.Errorf("Failed to chown path %s for %s with error %s", path, name, err)
	}
}

func doChownRequest(request ChownRequest) error {
	if request.Policy == "" {
		request.Policy = PolicyRecursive
	}
	if request.Policy == PolicyRecursive {
		err := filepath.Walk(request.Path, func(filePath string, file os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			chownFile(request.Name, request.Path, file, request.User, request.Group)
			return nil
		})
		if err != nil {
			log.Errorf("Failed to chown %s recursively for %s with error %s", request.Path, request.Name, err)
			return err
		}
	} else if request.Policy == PolicyRootOnly {
		file, err := os.Lstat(request.Path)
		if err != nil {
			log.Errorf("Failed to get stat of %s for %s with error %s", request.Path, request.Name, err)
			return err
		}
		chownFile(request.Name, request.Path, file, request.User, request.Group)
	} else {
		log.Fatalf("Unknown policy %s", request.Policy)
	}
	return nil
}

func chownRequests(requests map[string]ChownRequest) {
	for _, request := range requests {
		err := doChownRequest(request)
		if err != nil {
			continue
		}
	}
}

func run() {
	containerState := loadState(os.Stdin)
	destRequests := parseChownRequests(containerState.Annotations)
	requestsJson, err := json.Marshal(destRequests)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Parsed requests: %s", string(requestsJson))
	chownRequests(destRequests)
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
		Use:     "mount_chown [options]",
		Short:   "Invoked as a createContainer OCI-hooks to chown specific mount points",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			setupLogLevel()
			log.Infof("Run mount_chown %s", Version)
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
