// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gengogrpc contains the gRPC code generator.
package gengopbapi

import (
	"google.golang.org/protobuf/compiler/protogen"
)

// GenerateFileContent generates the gRPC service definitions, excluding the package statement.
func GenerateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}

	// 导入依赖package
	g.Import(protogen.GoImportPath("context"))
	g.Import(protogen.GoImportPath("encoding/json"))

	obj := new(PbapiObject)
	obj.Package = string(file.Desc.Name())
	obj.GoPackage = string(file.GoPackageName)
	obj.Imports = make(map[string]string)

	for _, service := range file.Services {
		sdesc := new(ServiceDesc)
		sdesc.Name = string(service.Desc.Name())
		sdesc.GoName = service.GoName
		for _, method := range service.Methods {
			mdesc := new(MethodDesc)
			mdesc.Name = string(method.Desc.Name())
			mdesc.GoName = method.GoName

			mdesc.InputType = method.Input.GoIdent.GoName
			mdesc.OutputType = method.Output.GoIdent.GoName

			// 用于生成demo代码部分
			mdesc.OuterInputType = g.QualifiedGoIdent(method.Input.GoIdent)
			mdesc.OuterOutputType = g.QualifiedGoIdent(method.Output.GoIdent)
			obj.Imports[string(method.Input.GoIdent.GoImportPath)] = ""
			obj.Imports[string(method.Output.GoIdent.GoImportPath)] = ""

			sdesc.Methods = append(sdesc.Methods, mdesc)
		}

		obj.Services = append(obj.Services, sdesc)
	}
	g.P(ExecuteService(obj))
}
