package templates

import (
	"bytes"
	"net/url"
	"text/template"

	"github.com/rancher/wharfie/pkg/registries"

	"github.com/k3s-io/k3s/pkg/daemons/config"
	"github.com/k3s-io/k3s/pkg/version"
)

type ContainerdRuntimeConfig struct {
	RuntimeType string
	BinaryName  string
}

type ContainerdConfig struct {
	NodeConfig            *config.Node
	DisableCgroup         bool
	SystemdCgroup         bool
	IsRunningInUserNS     bool
	EnableUnprivileged    bool
	NoDefaultEndpoint     bool
	NonrootDevices        bool
	PrivateRegistryConfig *registries.Registry
	ExtraRuntimes         map[string]ContainerdRuntimeConfig
	Program               string
}

type RegistryEndpoint struct {
	OverridePath bool
	URL          *url.URL
	Rewrites     map[string]string
	Config       registries.RegistryConfig
}

type HostConfig struct {
	Default   *RegistryEndpoint
	Program   string
	Endpoints []RegistryEndpoint
}

var HostsTomlHeader = "# File generated by " + version.Program + ". DO NOT EDIT.\n"

const HostsTomlTemplate = `
{{- /* */ -}}
# File generated by {{ .Program }}. DO NOT EDIT.
{{ with $e := .Default }}
{{- if $e.URL }}
server = "{{ $e.URL }}"
capabilities = ["pull", "resolve", "push"]
{{ end }}
{{- if $e.Config.TLS }}
{{- if $e.Config.TLS.CAFile }}
ca = [{{ printf "%q" $e.Config.TLS.CAFile }}]
{{- end }}
{{- if or $e.Config.TLS.CertFile $e.Config.TLS.KeyFile }}
client = [[{{ printf "%q" $e.Config.TLS.CertFile }}, {{ printf "%q" $e.Config.TLS.KeyFile }}]]
{{- end }}
{{- if $e.Config.TLS.InsecureSkipVerify }}
skip_verify = true
{{- end }}
{{ end }}
{{ end }}
[host]
{{ range $e := .Endpoints -}}
[host."{{ $e.URL }}"]
  capabilities = ["pull", "resolve"]
  {{- if $e.OverridePath }}
  override_path = true
  {{- end }}
{{- if $e.Config.TLS }}
  {{- if $e.Config.TLS.CAFile }}
  ca = [{{ printf "%q" $e.Config.TLS.CAFile }}]
  {{- end }}
  {{- if or $e.Config.TLS.CertFile $e.Config.TLS.KeyFile }}
  client = [[{{ printf "%q" $e.Config.TLS.CertFile }}, {{ printf "%q" $e.Config.TLS.KeyFile }}]]
  {{- end }}
  {{- if $e.Config.TLS.InsecureSkipVerify }}
  skip_verify = true
  {{- end }}
{{ end }}
{{- if $e.Rewrites }}
  [host."{{ $e.URL }}".rewrite]
  {{- range $pattern, $replace := $e.Rewrites }}
    "{{ $pattern }}" = "{{ $replace }}"
  {{- end }}
{{ end }}
{{ end -}}
`

func ParseTemplateFromConfig(templateBuffer string, config interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Funcs(templateFuncs).Parse(templateBuffer))
	template.Must(t.New("base").Parse(ContainerdConfigTemplate))
	if err := t.Execute(out, config); err != nil {
		return "", err
	}
	return out.String(), nil
}

func ParseHostsTemplateFromConfig(templateBuffer string, config interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Funcs(templateFuncs).Parse(templateBuffer))
	if err := t.Execute(out, config); err != nil {
		return "", err
	}
	return out.String(), nil
}
