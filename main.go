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

type App struct {
	FilePath         string
	Modules          []string
	WorkingDirectory string
}

func main() {
	var app App

	var rootCmd = &cobra.Command{
		Use:   "skaffoldrunner",
		Short: "Skaffoldrunner reads the skaffold.yaml file and launches the selected modules",
		Long: `Skaffoldrunner is a CLI tool that helps you run Skaffold more efficiently
by allowing you to select modes, profiles, and modules interactively.`,
		Run: func(cmd *cobra.Command, args []string) {
			app.runSkaffold()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&app.WorkingDirectory, "workdir", "w", "", "Working directory (default is current directory)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func (a *App) initialise() error {
	if a.WorkingDirectory == "" {
		fmt.Println("Running skaffold from current location as no --workdir specified")
	}

	skaffoldPath := "skaffold.yaml"
	if a.WorkingDirectory != "" {
		skaffoldPath = a.WorkingDirectory + skaffoldPath
	}

	modules, err := parser.ParseYamlForModules(skaffoldPath)
	if err != nil {
		return err
	}

	if len(modules) == 0 {
		return fmt.Errorf("no modules found for file at %s", a.FilePath)
	}

	a.Modules = modules
	return nil
}

func (a *App) runSkaffold() {
	if err := a.initialise(); err != nil {
		log.Fatal(err)
	}

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
	}, false)
	if err != nil {
		log.Fatal(err)
	}

	var selectedModules []string
	if len(a.Modules) > 0 {
		selectedModules, err = prompts.MultiSelectPrompt(prompts.SelectPromptParams{Label: "Select the modules you'd like to run", Items: a.Modules}, true)
		if err != nil {
			log.Fatal(err)
		}
	}

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
	if a.WorkingDirectory != "" {
		cmd.Dir = a.WorkingDirectory
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

	<-sigChan
	fmt.Println("Interrupt signal received, terminating the process...")
}
