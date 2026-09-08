package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/tylergannon/tractor/graph"
)

type executableIdentity struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Version string `json:"version"`
}

var (
	executableOnce        sync.Once
	executableCurrent     executableIdentity
	executableIdentityErr error
)

func (r *Runner) setManifestProvenance(manifest *runManifest) error {
	executable, err := currentExecutableIdentity()
	if err != nil {
		return err
	}
	graphHash, err := graphSHA256(r.graph)
	if err != nil {
		return err
	}
	manifest.Argv = append([]string(nil), os.Args...)
	manifest.Executable = executable
	manifest.PipelineSource = r.config.PipelineSource
	manifest.GraphSHA256 = graphHash
	return nil
}

func normalizePipelineSource(source string) (string, error) {
	if source == "" {
		return "api", nil
	}
	switch source {
	case "api", "inline":
		return source, nil
	}
	if name, ok := strings.CutPrefix(source, "builtin:"); ok && strings.TrimSpace(name) != "" {
		return source, nil
	}
	if path, ok := strings.CutPrefix(source, "file:"); ok && filepath.IsAbs(path) {
		return source, nil
	}
	return "", fmt.Errorf("invalid pipeline source %q", source)
}

func currentExecutableIdentity() (executableIdentity, error) {
	executableOnce.Do(func() {
		path, err := os.Executable()
		if err != nil {
			executableIdentityErr = fmt.Errorf("resolve executable: %w", err)
			return
		}
		file, err := os.Open(path)
		if err != nil {
			executableIdentityErr = fmt.Errorf("open executable: %w", err)
			return
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			executableIdentityErr = fmt.Errorf("hash executable: %w", copyErr)
			return
		}
		if closeErr != nil {
			executableIdentityErr = fmt.Errorf("close executable after hashing: %w", closeErr)
			return
		}
		version := "unknown"
		if build, ok := debug.ReadBuildInfo(); ok && build.Main.Version != "" {
			version = build.Main.Version
		}
		executableCurrent = executableIdentity{
			Path: path, SHA256: hex.EncodeToString(hash.Sum(nil)), Version: version,
		}
	})
	return executableCurrent, executableIdentityErr
}

func graphSHA256(pipeline graph.Graph) (string, error) {
	encoded, err := json.Marshal(pipeline)
	if err != nil {
		return "", fmt.Errorf("encode resolved graph for provenance: %w", err)
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}
