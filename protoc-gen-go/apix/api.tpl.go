package apix

import (
	. "github.com/obase/api/x"
	"strings"
	"text/template"
)

const SERVICE_TEMPLATE = `
{{- $fname := .FileName}}
func init_{{$fname}}(server *apix.Server) {
	{{- range .ServerOption}}
	server.ServerOption({{.Name}}.{{.Func}})
	{{- end}}
	{{- range .MiddleFilter}}
	server.MiddleFilter({{.Name}}.{{.Func}})
	{{- end}}
}
{{range .Services}}
{{- $sname := .Name}}
{{- range .Methods}}
func func_{{$sname}}_{{.Name}}(service {{$sname}}Server) apix.MethodFunc {
	return func(ctx context.Context, data []byte) (interface{}, error) {
		var req *{{.InputType}}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &req); err != nil {
				return nil, apix.ParsingRequestError(err, "{{.Tag}}")
			}
		}
		return service.{{.Name}}(ctx, req)
	}
}
{{end}}
func Register{{$sname}}Service(server *apix.Server, service {{$sname}}Server) {
	var smeta *apix.Service
	var hmeta *apix.Method

	server.Init(init_{{$fname}})
	smeta = server.Service(&_{{$sname}}_serviceDesc, service)
	smeta.GroupPath("{{.GroupPath}}")
	{{- range .GroupFilter}}
	smeta.GroupFilter({{.Name}}.{{.Func}})
	{{- end}}
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

	tpl := template.Must(template.New("generateContentTemplate").Parse(SERVICE_TEMPLATE))

	sb := new(strings.Builder)
	if err := tpl.Execute(sb, data); err != nil {
		panic(err)
	}
	return sb.String()
}

type Template struct {
	FileName     string
	FilePack     string
	ServerOption []*funcDesc
	MiddleFilter []*funcDesc
	Files        []*fileDesc
	Services     []*serviceDesc
	Imports      map[string]interface{} // 生成service实现代码需要的imports
}

type funcDesc struct {
	Name string
	Func string
}

type fileDesc struct {
	MessageName string
	FieldName   string // 定义为ISN的字段名字
}

type serviceDesc struct {
	Name        string
	GroupPath   string
	GroupFilter []*funcDesc
	Methods     []*methodDesc
}

type methodDesc struct {
	Name            string
	Tag             string
	OuterInputType  string
	InputType       string
	OuterOutputType string
	OutputType      string
	HandlePath      string
	HandleBody      Body
	HandleFilter    []*funcDesc
	SocketPath      string
	SocketFilter    []*funcDesc
}
