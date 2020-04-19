package gengopbapi

import (
	"strings"
	"text/template"
)

const SERVICE_TEMPLATE = `
{{- $pname := .Package}}
{{range .Services}}
{{- $sname := .Name}}
{{- $sname_go := .GoName}}
// service: {{$sname}}
func Register{{$sname_go}}ServerHandler(impl interface{}) (*grpc.ServiceDesc, string, string, map[string]func(context.Context, []byte) (interface{}, error)) {
	service := impl.({{$sname_go}}Server)
	adapters := make(map[string]func(context.Context, []byte) (interface{}, error))
	{{range .Methods}}
	// method: {{.Name}}
	adapters["{{.Name}}"] = func(ctx context.Context, data []byte) (ret interface{}, err error) {
		var req *{{.InputType}}
		if len(data) > 0 {
			if err = json.Unmarshal(data, &req); err != nil {
				return
			}
		}
		ret, err = service.{{.Name}}(ctx, req)
		return
	}
	{{- end}}
	return &_{{$sname_go}}_serviceDesc, "{{$pname}}", "{{.Name}}", adapters
}
{{- end}}
/*---------------autogen service implement------------------
package service

import (
	"context"
	// TBD: path for package "{{.GoPackageName}}"
)
{{- $palias := .GoPackageName}}
{{- range .Services}}
{{- $sname := .GoName}}
type {{$sname}}Service struct {}
var _ {{$pname}}.{{$sname}}Server = (*{{$sname}}Service)(nil)
{{- range .Methods}}
func (s *{{$sname}}Service) {{.Name}}(ctx context.Context, req *{{$palias}}.{{.OuterInputType}}) (rsp *{{$palias}}.{{.OuterOutputType}}, err error) {
	return
}
{{- end}}
{{- end}}
---------------autogen service implement------------------*/
`

func ExecuteService(data *PbapiObject) string {
	tpl := template.Must(template.New("service_template").Parse(SERVICE_TEMPLATE))
	sb := new(strings.Builder)
	if err := tpl.Execute(sb, data); err != nil {
		panic(err)
	}
	return sb.String()
}

type PbapiObject struct {
	Package       string // pb包名, 例如"google.api"
	GoPackageName string // go包名, 例如"google_api"
	GoImportPath  string
	Services      []*ServiceDesc
	Imports       map[string]string
}

type ServiceDesc struct {
	Name    string // pb服务名, 例如"Insert_Student"
	GoName  string // go服务名, 例如"InsertStudent"
	Methods []*MethodDesc
}

type MethodDesc struct {
	Name            string // pb方法名, 例如"insert_student"
	GoName          string // go方法名, 例如"InsertStudent"
	InputType       string // go输入类型
	OutputType      string // go输出类型
	OuterInputType  string
	OuterOutputType string
	Idempotent      bool
}
