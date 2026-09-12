# crabcore

**crabcore** is a tool for managing the configuration of a user's working environment.

It allows you to store configuration in YAML and automatically distribute:

- tool configuration;
- dynamic Go plugins;
- plugin configuration.

Plugins can add their own commands to `crabcore` and optionally run them automatically after an update.

## Installation

Download binary and run update, it will install itself to ~/.local/bin/crabcore, add ~/.local/bin to PATH in `.bashrc`
On linux with systemd it also will add itself to users timers for auto update.

## Configuration
The configuration URL can be passed directly:

```bash
crabcore update --config https://example.com/crabcore.yaml
```

Configuration will be saved to ~/.crabcore.yaml. In the future, you can run an update without config url:

```bash
crabcore update
```

### Runtime configuration

For example:

```yaml
binUrl: https://example.com/crabcore

plugins:
  - plugin: ssh
    filename: ssh.so
    url: https://example.com/ssh.so
    moduleConfig:
      ssh-config: https://example.com/ssh-config.yaml
```

`binUrl` is the URL of the current `crabcore` binary.

`plugins` is the list of plugins:

| Field | Description |
|---|---|
| `plugin` | plugin name |
| `filename` | `.so` file name |
| `url` | plugin URL |
| `moduleConfig` | module configuration |

After changing the configuration, run:

```bash
crabcore update
```

---

## Plugin development

Plugins are dynamic Go libraries:

```bash
go build -buildmode=plugin -o example.so .
```

A plugin must export a `CrabcorePlugin` variable implementing `pluginapi.SymbolProvider`.
Each symbol must point to an object implementing `pluginapi.Module`.


A minimal plugin:

```go
package main

import (
	"fmt"

	"github.com/zerodotfive/crabcore/pkg/pluginapi"

	"github.com/spf13/cobra"
)

var pluginName = "example"

var CrabcorePlugin plugin
var ExampleModule exampleModule

var _ pluginapi.SymbolProvider = (*plugin)(nil)
var _ pluginapi.Module = (*exampleModule)(nil)

var symbols = []string{"ExampleModule"}

type plugin struct{}

func (p *plugin) GetRoot() (*cobra.Command, error) {
	pluginRoot := &cobra.Command{
		Use:   pluginName,
		Short: fmt.Sprintf("SSH commands"),
	}

	return pluginRoot, nil
}

func (p *plugin) GetSymbols() []string {
	return symbols
}

type exampleModule struct{}

func (m *exampleModule) GetName() string {
	return "example"
}

func (m *exampleModule) Init(config []byte) error {
	return nil
}

func (m *exampleModule) Commands() (*cobra.Command, error) {
	return &cobra.Command{
		Use:   m.GetName(),
		Short: "Example crabcore command",
		//Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := m.run()
			return fmt.Errorf("[%s %s] %s", pluginName, m.GetName(), err)
		},
	}, nil
}

func (m *exampleModule) run() (string, error) {
	return "ok", nil
}
```

After loading the plugin, the command becomes part of the CLI:

```bash
crabcore example
```

### Running after `update`

If a command should be executed automatically after an update, add the annotation:

```go
Annotations: map[string]string{
    "run-on-update": "true",
},
```

For example:

```go
return &cobra.Command{
    Use:   m.GetName(),
    Short: "Example crabcore command",
    //Hidden: true,
    RunE: func(cmd *cobra.Command, args []string) error {
        _, err := m.run()
        return fmt.Errorf("[%s %s] %s", pluginName, m.GetName(), err)
    },
}, nil
```

In this case:

```bash
crabcore update
```

will automatically execute `example` after the update.

### Module configuration

Each module can have its own configuration URL:

```yaml
plugins:
  - plugin: example
    filename: example.so
    url: https://example.com/example.so
    moduleConfig:
      example: https://example.com/example.yaml
```

The contents of the file are passed to:

```go
func (m *ExampleModule) Init(config []byte) error
```

This allows a single `.so` to contain multiple independent modules with separate configurations.

> **Note:** Go plugins have platform and build compatibility limitations. The plugin must be built with a compatible Go toolchain and environment. See the official Go documentation for details.
