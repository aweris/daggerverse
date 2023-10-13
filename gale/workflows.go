package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type Workflows struct {
	Config *WorkflowsConfig
}

// List lists all workflows and their jobs in the repository.
func (w *Workflows) List(ctx context.Context) error {
	dir := w.Config.Source.Directory(w.Config.WorkflowsDir)

	entries, err := dir.Entries(ctx)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)
			path := filepath.Join(w.Config.WorkflowsDir, entry)

			// dagger do not support maps yet, so we're defining anonymous struct to unmarshal the yaml file to avoid
			// hit this limitation.

			var workflow struct {
				Name string                 `yaml:"name"`
				Jobs map[string]interface{} `yaml:"jobs"`
			}

			if err := file.unmarshalContentsToYAML(ctx, &workflow); err != nil {
				return err
			}

			fmt.Printf("Workflow: ")
			if workflow.Name != "" {
				fmt.Printf("%s (path: %s)\n", workflow.Name, path)
			} else {
				fmt.Printf("%s\n", path)
			}

			fmt.Println("Jobs:")

			for job := range workflow.Jobs {
				fmt.Printf(" - %s\n", job)
			}

			fmt.Println("") // extra empty line
		}
	}

	return nil
}
