package pgapi

import (
	"strings"
	"text/template"
)

const SERVICE_TEMPLATE = `
{{range .Services}}
func Register{{$sname}}ServerHandler(impl interface{}) (*grpc.ServiceDesc, string, string, string, map[string]func(context.Context, []byte) ([]byte, error), map[string]bool) {
	
	service := ({{$sname}}Server)(impl)

	{{range .Methods}}
	hmeta = smeta.Method("{{$sname}}.{{.Name}}", func_{{$sname}}_{{.Name}}(service))
	hmeta.HandlePath("{{.HandlePath}}")
	{{- range .HandleFilter}}
	hmeta.HandleFilter({{.Name}}.{{.Func}})
	{{- end}}
	hmeta.SocketPath("{{.SocketPath}}")
	{{- range .HandleFilter}}
	hmeta.SocketFilter({{.Name}}.{{.Func}})
	{{- end}}
	{{- end}}
}
{{end}}
/* autogen service implement
import (
	"context"
	{{- range $k, $v := .Imports}}
	"{{$k}}"
	{{- end}}
)
{{- $fpack := .FilePack}}
{{- range .Services}}
{{- $sname := .Name}}
type {{$sname}}Service struct {}
var _ {{$fpack}}.{{$sname}}Server = (*{{$sname}}Service)(nil)
{{- range .Methods}}
func (s *{{$sname}}Service) {{.Name}}(ctx context.Context, req *{{.OuterInputType}}) (rsp *{{.OuterOutputType}}, err error) {
	return
}
{{- end}}
{{- end}}
*/
`

func ExecuteService(data *Template) string {
	tpl := template.Must(template.New("service_template").Parse(SERVICE_TEMPLATE))
	sb := new(strings.Builder)
	if err := tpl.Execute(sb, data); err != nil {
		panic(err)
	}
	return sb.String()
}

type Template struct {
	Pname    string
	Services []ServiceDesc
}

type ServiceDesc struct {
	Name    string
	Methods []MethodDesc
}

type MethodDesc struct {
	Name            string
	Tag             string
	InputType       string
	OutputType      string
	OuterInputType  string
	OuterOutputType string
	Idempotent      bool
}
