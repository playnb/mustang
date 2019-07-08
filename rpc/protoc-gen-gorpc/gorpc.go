package main

import (
	"bytes"
	"text/template"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

func init() {
	generator.RegisterPlugin(new(protorpcPlugin))
}

type protorpcPlugin struct {
	*generator.Generator
}

func (p *protorpcPlugin) Name() string {
	return "gorpc"
}

func (p *protorpcPlugin) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *protorpcPlugin) GenerateImports(file *generator.FileDescriptor) {
	if !p.getGenericServicesOptions(file) {
		return
	}
	if len(file.Service) > 0 {
		//TODO: 生成IMPORT部分
		//p.P(`import "net/rpc"`)
		p.P(`import "errors"`)
	}
}

//....
func (p *protorpcPlugin) getGenericServicesOptions(file *generator.FileDescriptor) bool {
	return true
}

//....
func (p *protorpcPlugin) getGenericServicesOptionsUsePkgName(file *generator.FileDescriptor) bool {
	return false
}

// Generate generates the Service interface.
// rpc proxy can't handle other proto message!!!
func (p *protorpcPlugin) Generate(file *generator.FileDescriptor) {
	if !p.getGenericServicesOptions(file) {
		return
	}
	for _, svc := range file.Service {
		p.genEmptyClient(file, svc)
		//TODO: 生成代码的部分
		//参考 https://github.com/chai2010/protorpc/blob/master/protoc-gen-go/main.go
	}
}

/*
type EchoServiceClient interface {
	Echo(ctx context.Context, in *RequestEcho, opts ...grpc.CallOption) (*ReplyEcho, error)
}
*/

func (p *protorpcPlugin) genEmptyClient(file *generator.FileDescriptor, svc *descriptor.ServiceDescriptorProto) {
	//log.Println("=========================================")

	{
		//Struct定义
		const emptyClientStruct = `
		var _insEmpty{{.ServiceName}}Client Empty{{.ServiceName}}Client
		func GetEmpty{{.ServiceName}}Client() {{.ServiceName}}Client{
			return &_insEmpty{{.ServiceName}}Client
		}
		type Empty{{.ServiceName}}Client struct {
		}
		`

		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(emptyClientStruct))
		t.Execute(out, &struct{ ServiceName string }{
			ServiceName: generator.CamelCase(svc.GetName()),
		})
		p.P(out.String())
		//log.Println(out.String())
	}

	{
		//Method定义
		const emptyClientMethod = `
		func (client *Empty{{.ServiceName}}Client) {{.MethodName}}(ctx context.Context, in *{{.ArgsType}}, opts ...grpc.CallOption) (*{{.ReplyType}}, error){
			out := new({{.ReplyType}})
			return out, errors.New("Use Empty Client")		
		}
		`

		methodString := ""
		for _, m := range svc.Method {
			out := bytes.NewBuffer([]byte{})
			t := template.Must(template.New("").Parse(emptyClientMethod))
			t.Execute(out, &struct{ ServiceName, MethodName, ArgsType, ReplyType string }{
				ServiceName: generator.CamelCase(svc.GetName()),
				MethodName:  generator.CamelCase(m.GetName()),
				ArgsType:    p.TypeName(p.ObjectNamed(m.GetInputType())),
				ReplyType:   p.TypeName(p.ObjectNamed(m.GetOutputType())),
			})

			methodString += out.String()
		}
		p.P(methodString)
		//log.Println(methodString)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////