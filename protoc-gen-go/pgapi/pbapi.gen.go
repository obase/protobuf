package pgapi

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/obase/api/x"
	"strings"
)

func init() {
	generator.RegisterPlugin(new(apix))
}

type apix struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "grpc".
func (g *apix) Name() string {
	return "apix"
}

// Init initializes the plugin.
func (g *apix) Init(gen *generator.Generator) {
	g.gen = gen
}

func (g *apix) typePath(name string) string {
	obj := g.gen.ObjectNamed(name)
	return string(obj.GoImportPath())
}

func (g *apix) outerTypeName(name string, fpack string) string {
	obj := g.gen.ObjectNamed(name)
	pack := g.gen.DefaultPackageName(obj)
	if pack == "" {
		pack = fpack + "."
	}
	return pack + generator.CamelCaseSlice(obj.TypeName())
}

func (g *apix) typeName(name string) string {
	g.gen.RecordTypeUse(name)
	obj := g.gen.ObjectNamed(name)
	return g.gen.TypeName(obj)
}

// GenerateImports generates the import declaration for this file.
func (g *apix) GenerateImports(file *generator.FileDescriptor) {

}

// Generate generates code for the services in the given file.
func (g *apix) Generate(file *generator.FileDescriptor) {

	g.addImport("context", "")
	g.addImport("encoding/json", "")


	val := new(Template)
	val.FileName = identifier(*file.Name)
	val.FilePack = string(g.gen.GoPackageName(generator.GoImportPath(*file.Package)))
	val.Imports = make(map[string]interface{})

	// File Option处理
	tmp, err := proto.GetExtension(file.Options, x.E_ServerOption)
	if err == nil {
		if pfs, ok := tmp.([]*x.PackFunc); ok {
			for _, pf := range pfs {
				val.ServerOption = append(val.ServerOption, g.addImport(pf.Pack, pf.Func))
			}
		}
	}

	tmp, err = proto.GetExtension(file.Options, x.E_MiddleFilter)
	if err == nil {
		if pfs, ok := tmp.([]*x.PackFunc); ok {
			for _, pf := range pfs {
				val.MiddleFilter = append(val.MiddleFilter, g.addImport(pf.Pack, pf.Func))
			}
		}
	}

	// Message Option处理
	for _, message := range file.MessageType {
		var fileok bool
		for _, field := range message.Field {
			tmp, err = proto.GetExtension(field.Options, x.E_Field)
			if err == nil {
				if fld, ok := tmp.(*x.Field); ok {
					if !fileok && fld.File {
						fdesc := new(fileDesc)
						fdesc.MessageName = generator.CamelCase(*message.Name)
						fdesc.FieldName = generator.CamelCase(*field.Name)
						val.Files = append(val.Files, fdesc)
						fileok = true
					}
				}
			}
		}
	}

	// Service Option处理
	for _, service := range file.Service {
		sdesc := new(serviceDesc)
		sdesc.Name = generator.CamelCase(*service.Name)

		tmp, err = proto.GetExtension(service.Options, x.E_Group)
		if err == nil {
			if grp, ok := tmp.(*x.Group); ok {
				sdesc.GroupPath = grp.Path
			}
		}

		tmp, err = proto.GetExtension(service.Options, x.E_GroupFilter)
		if err == nil {
			if pfs, ok := tmp.([]*x.PackFunc); ok {
				for _, pf := range pfs {
					sdesc.GroupFilter = append(sdesc.GroupFilter, g.addImport(pf.Pack, pf.Func))
				}
			}
		}

		for _, method := range service.Method {
			mdesc := new(methodDesc)
			mdesc.Name = generator.CamelCase(*method.Name)
			mdesc.Tag = sdesc.Name + "." + mdesc.Name

			mdesc.InputType = g.typeName(*method.InputType)
			mdesc.OutputType = g.typeName(*method.OutputType)

			mdesc.OuterInputType = g.outerTypeName(*method.InputType, val.FilePack)
			mdesc.OuterOutputType = g.outerTypeName(*method.OutputType, val.FilePack)

			val.Imports[g.typePath(*method.InputType)] = nil
			val.Imports[g.typePath(*method.OutputType)] = nil

			tmp, err = proto.GetExtension(method.Options, x.E_Handle)
			if err == nil {
				if hdl, ok := tmp.(*x.Handle); ok {
					mdesc.HandlePath = hdl.Path
					mdesc.HandleBody = hdl.Body
				}
			}

			tmp, err = proto.GetExtension(method.Options, x.E_HandleFilter)
			if err == nil {
				if pfs, ok := tmp.([]*x.PackFunc); ok {
					for _, pf := range pfs {
						mdesc.HandleFilter = append(mdesc.HandleFilter, g.addImport(pf.Pack, pf.Func))
					}
				}
			}

			tmp, err = proto.GetExtension(method.Options, x.E_Socket)
			if err == nil {
				if skt, ok := tmp.(*x.Socket); ok {
					mdesc.SocketPath = skt.Path
				}
			}

			tmp, err = proto.GetExtension(method.Options, x.E_SocketFilter)
			if err == nil {
				if pfs, ok := tmp.([]*x.PackFunc); ok {
					for _, pf := range pfs {
						mdesc.SocketFilter = append(mdesc.SocketFilter, g.addImport(pf.Pack, pf.Func))
					}
				}
			}

			sdesc.Methods = append(sdesc.Methods, mdesc)
		}

		val.Services = append(val.Services, sdesc)
	}

	g.gen.P(ExecuteService(val))
}

func (g *apix) addImport(pack string, fn string) *funcDesc {
	gname := g.gen.AddImport(generator.GoImportPath(pack))
	return &funcDesc{
		Name: string(gname),
		Func: fn,
	}
}

func identifier(pack string) string {
	sb := new(strings.Builder)
	for _, ch := range pack {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			sb.WriteRune(ch)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}
