package main

import (
	"embed"
	"fmt"
	"path/filepath"

	"os"

	"github.com/revanite-io/pvtr-gcp-cloud-storage/data"
	"github.com/revanite-io/pvtr-gcp-cloud-storage/evaluation_plans"

	"github.com/privateerproj/privateer-sdk/command"
	"github.com/privateerproj/privateer-sdk/pluginkit"
)

var (
	// Version is to be replaced at build time by the associated tag
	Version = "0.0.0"
	// VersionPostfix is a marker for the version such as "dev", "beta", "rc", etc.
	VersionPostfix = "dev"
	// GitCommitHash is the commit at build time
	GitCommitHash = ""
	// BuiltAt is the actual build datetime
	BuiltAt = ""

	PluginName   = "pvtr-gcp-cloud-storage"
	RequiredVars = []string{"bucketname"}

	//go:embed data/catalogs
	files   embed.FS
	dataDir = filepath.Join("data", "catalogs")
)

func main() {
	if VersionPostfix != "" {
		Version = fmt.Sprintf("%s-%s", Version, VersionPostfix)
	}

	orchestrator := pluginkit.EvaluationOrchestrator{
		PluginName:    PluginName,
		PluginVersion: Version,
		PluginUri:     "github.com/revanite-io/pvtr-gcp-cloud-storage",
	}
	orchestrator.AddLoader(data.Loader)

	err := orchestrator.AddReferenceCatalogs(dataDir, files)
	if err != nil {
		fmt.Printf("Error loading catalog: %v\n", err)
		os.Exit(1)
	}

	orchestrator.AddRequiredVars(RequiredVars)
	err = orchestrator.AddEvaluationSuite("CCC.ObjStor", nil, evaluation_plans.CCC_ObjStor)
	if err != nil {
		fmt.Printf("Error adding evaluation suite: %v\n", err)
		os.Exit(1)
	}

	runCmd := command.NewPluginCommands(
		PluginName,
		Version,
		VersionPostfix,
		GitCommitHash,
		&orchestrator,
	)

	err = runCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}