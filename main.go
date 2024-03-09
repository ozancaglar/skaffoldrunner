package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ozancaglar/skaffoldrunner/parser"
	"github.com/ozancaglar/skaffoldrunner/prompts"
	"github.com/spf13/cobra"
)

const (
	SKAFFOLD     = "skaffold"
	PROFILE_FLAG = "-p"
	MODULE_FLAG  = "-m"
)

// App represents your application and holds data needed across commands
type App struct {
	FilePath         string
	Modules          []string
	WorkingDirectory string
}

func (a *App) initialise() {
	var rootCmd = &cobra.Command{
		Use:   "skaffoldrunner",
		Short: "Skaffoldrunner reads your provided skaffold.yaml file and launches the selected modules for you",
		Run: func(cmd *cobra.Command, args []string) {
			a.FilePath, _ = cmd.Flags().GetString("file")
			a.WorkingDirectory, _ = cmd.Flags().GetString("workdir")

			if a.FilePath != "" {
				modules, err := parser.ParseYamlForModules(a.FilePath)
				if err != nil {
					log.Fatal(err)
				}

				if len(modules) == 0 {
					log.Fatalf("no modules found for file at %s", a.FilePath)
				}

				a.Modules = modules
			} else {
				fmt.Println("assuming you want to run all modules, as no --file flag was provided")
			}

			if a.WorkingDirectory != "" {
				fmt.Printf("Working directory set to: %s\n", a.WorkingDirectory)
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&a.FilePath, "file", "f", "", "Path to the YAML file")
	rootCmd.PersistentFlags().StringVarP(&a.WorkingDirectory, "workdir", "w", "", "Working directory")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := App{}
	app.initialise()

	modeResult, err := prompts.SelectPrompt(prompts.SelectPromptParams{
		Label: "Which mode would you like to run in?",
		Items: []string{"dev", "run"},
	})
	if err != nil {
		log.Fatal(err)
	}

	profileResult, err := prompts.SelectPrompt(prompts.SelectPromptParams{
		Label: "Which profile would you like to run against?",
		Items: []string{"local"},
	})
	if err != nil {
		log.Fatal(err)
	}

	selectedFlags, err := prompts.MultiSelectPrompt(prompts.SelectPromptParams{
		Label: "Select the optional flags you'd like to use",
		Items: []string{"--port-forward", "--tail"},
	})
	if err != nil {
		log.Fatal(err)
	}

	var selectedModules []string
	if len(app.Modules) > 0 {
		selectedModules, err = prompts.MultiSelectPrompt(prompts.SelectPromptParams{Label: "Select the modules you'd like to run", Items: app.Modules})
		if err != nil {
			log.Fatal(err)
		}
	}

	// Prepare arguments slice
	var args []string
	args = append(args, modeResult)

	args = append(args, PROFILE_FLAG, profileResult)
	args = append(append(args, MODULE_FLAG), strings.Join(selectedModules, ","))
	args = append(args, selectedFlags...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, SKAFFOLD, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if app.WorkingDirectory != "" {
		cmd.Dir = app.WorkingDirectory
	} else {
		fmt.Println("running skaffold from current location as no --workdir specified")
	}

	// Listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run the command in a goroutine so that it can be stopped with context
	go func() {
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for an interrupt signal
	<-sigChan
	fmt.Println("Interrupt signal received, terminating the process...")
}
