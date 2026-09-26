package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/zerodotfive/crabcore/pkg/paths"
	"gopkg.in/yaml.v3"
)

type HostnameWriter struct {
	hostname string
	writer   io.Writer
}

func (w *HostnameWriter) Write(p []byte) (int, error) {
	lines := bytes.SplitAfter(p, []byte("\n"))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		_, _ = fmt.Fprintf(w.writer, "[%s] %s", w.hostname, line)
	}

	return len(p), nil
}

type sshParallelModule struct {
	AutocompleteHostRegex string            `yaml:"autocompleteHostRegex"`
	CommandsDefinition    map[string]string `yaml:"commandsDefinition"`
	sshHosts              []string
	threads               int
}

func (m *sshParallelModule) GetName() string {
	return "parallel"
}

func (m *sshParallelModule) IsRootAllowed() bool {
	return false
}

func (m *sshParallelModule) Init(config []byte) error {
	if len(config) == 0 {
		return errors.New("empty config")
	}

	if err := yaml.Unmarshal(config, m); err != nil {
		return err
	}

	m.CommandsDefinition["custom"] = ""

	reAutocompleteHostRegex := regexp.MustCompile(m.AutocompleteHostRegex)

	sshConfigDPath := path.Join(paths.Home(), ".ssh/crabcore.d")

	entries, err := os.ReadDir(sshConfigDPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file, err := os.Open(filepath.Join(sshConfigDPath, entry.Name()))
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if after, ok := strings.CutPrefix(line, "host "); ok {
				line = strings.TrimSpace(after)
				if reAutocompleteHostRegex.MatchString(line) {
					m.sshHosts = append(m.sshHosts, line)
				}
			}
		}

		_ = file.Close()

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (m *sshParallelModule) filterHosts(pattern string) int {
	re := regexp.MustCompile("^" + pattern)
	var result []string

	for _, item := range m.sshHosts {
		if re.MatchString(item) {
			result = append(result, item)
		}
	}

	m.sshHosts = result

	return len(result)
}

func (m *sshParallelModule) Commands() (*cobra.Command, error) {
	moduleRoot := &cobra.Command{
		Use:   m.GetName(),
		Short: fmt.Sprintf("Run commands on multiple hosts"),
		Long:  fmt.Sprintf("crabcore ssh parallel <host regex> <command> [args]"),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return m.sshHosts, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				return slices.Collect(maps.Keys(m.CommandsDefinition)), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Help()
			}

			if m.filterHosts(args[0]) == 0 {
				return fmt.Errorf("no hosts matched %s", args[0])
			}

			subCommand, checkSubCommand := m.CommandsDefinition[args[1]]
			if !checkSubCommand {
				return fmt.Errorf("ssh parallel unknown command %s", args[1])
			}

			if args[1] == "custom" {
				if len(args) < 3 {
					return fmt.Errorf("custom command requires at least one argument")
				}

				subCommand = strings.Join(args[2:], " ")
			}

			info, err := m.run(subCommand)
			if err != nil {
				return fmt.Errorf("[%s %s] %s\n\n%s\n", pluginName, m.GetName(), err, info)
			}

			return nil
		},
	}

	moduleRoot.Flags().IntVarP(&m.threads, "threads", "t", 10, "Number of threads")

	return moduleRoot, nil
}

func (m *sshParallelModule) run(command string) (string, error) {
	jobs := make(chan string)
	results := make(chan map[string]int, len(m.sshHosts))

	var wg sync.WaitGroup

	for range m.threads {
		wg.Go(func() {
			for host := range jobs {
				results <- func() map[string]int {
					cmd := exec.Command("ssh", host, command)

					cmd.Stdout = &HostnameWriter{
						hostname: host,
						writer:   os.Stdout,
					}

					cmd.Stderr = &HostnameWriter{
						hostname: host,
						writer:   os.Stderr,
					}

					err := cmd.Run()
					if err != nil {
						var exitErr *exec.ExitError
						if errors.As(err, &exitErr) {
							return map[string]int{host: exitErr.ExitCode()}
						}

						_, _ = fmt.Fprintf(os.Stderr, "[%s %s] < %s > crabcore error: %s\n", pluginName, m.GetName(), host, err)
						return map[string]int{host: -1}
					}
					return map[string]int{host: 0}
				}()
			}
		})
	}

	for _, host := range m.sshHosts {
		jobs <- host
	}

	close(jobs)
	wg.Wait()
	close(results)
	exitCodeSuccess := true
	exitCodeInfo := ""

	for result := range results {
		for h, e := range result {
			if e != 0 {
				exitCodeSuccess = false
				exitCodeInfo = exitCodeInfo + fmt.Sprintf("%v: %d\n", h, e)
			}
		}
	}

	if !exitCodeSuccess {
		return exitCodeInfo, fmt.Errorf("one or more hosts has returned non-zero exit code")
	}

	return "", nil
}
