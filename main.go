package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// RepoConfig holds the configuration for a repo
type RepoConfig struct {
	Prompt string `yaml:"prompt"`
}

// Config holds the top-level configuration for the program
type Config struct {
	Version string                `yaml:"version"`
	Repos   map[string]RepoConfig `yaml:"repos"`
}

func main() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		return
	}

	debug := false
	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	bytes, err := ioutil.ReadFile(filepath.Join(home, ".qrgpt.yaml"))
	if err != nil {
		fmt.Println(err)
		return
	}
	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		panic(err)
	}
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "usage: qrgpt [prompt_var_1] [prompt_var_2] ...")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "CONFIG")
		fmt.Fprintln(os.Stderr, "------")
		fmt.Fprintln(os.Stderr, string(bytes))
		os.Exit(1)
	}

	// Get the origin URL of the repo
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	originURL := strings.TrimSpace(string(output))

	var repoName string
	if strings.Contains(originURL, "@") {
		parts := strings.Split(originURL, "@")
		repoName = strings.Replace(parts[1], ":", "/", 1)
	} else {
		repoName = strings.TrimPrefix(originURL, "https://")
	}

	var repoConfig *RepoConfig
	for name, repo := range config.Repos {
		if name == repoName {
			repoConfig = &repo
			break
		}
	}
	if repoConfig == nil {
		fmt.Println("origin does not match any repos in configuration")
		return
	}

	execFunc := template.FuncMap(map[string]any{
		"exec": func(args ...string) string {
			cmd := exec.Command(args[0], args[1:]...)
			if debug {
				spew.Dump(cmd)
			}
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Fprintln(os.Stderr, "====> ERROR:", err)
				fmt.Fprintln(os.Stderr, string(output))
				os.Exit(1)
			}
			return string(output)
		},
	})
	tmpl, err := template.New("prompt").Funcs(execFunc).Funcs(sprig.FuncMap()).Parse(repoConfig.Prompt)
	if err != nil {
		panic(err)
	}
	var promptBuilder strings.Builder
	err = tmpl.Execute(&promptBuilder, struct{ Args []string }{os.Args})
	if err != nil {
		panic(err)
	}
	fmt.Println(promptBuilder.String())
}
