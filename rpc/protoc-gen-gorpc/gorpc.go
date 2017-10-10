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

func (p *protorpcPlugin) Name() string { return "gorpc" }

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
// rpc service can't handle other proto message!!!
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
func (p *protorpcPlugin) genServiceInterface(
	file *generator.FileDescriptor,
	svc *descriptor.ServiceDescriptorProto,
) {
	const serviceInterfaceTmpl = `
type {{.ServiceName}} interface {
	{{.CallMethodList}}
}
`
	const callMethodTmpl = `
{{.MethodName}}(in *{{.ArgsType}}, out *{{.ReplyType}}) error`

	// gen call method list
	var callMethodList string
	for _, m := range svc.Method {
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(callMethodTmpl))
		t.Execute(out, &struct{ ServiceName, MethodName, ArgsType, ReplyType string }{
			ServiceName: generator.CamelCase(svc.GetName()),
			MethodName:  generator.CamelCase(m.GetName()),
			ArgsType:    p.TypeName(p.ObjectNamed(m.GetInputType())),
			ReplyType:   p.TypeName(p.ObjectNamed(m.GetOutputType())),
		})
		callMethodList += out.String()

		p.RecordTypeUse(m.GetInputType())
		p.RecordTypeUse(m.GetOutputType())
	}

	// gen all interface code
	{
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(serviceInterfaceTmpl))
		t.Execute(out, &struct{ ServiceName, CallMethodList string }{
			ServiceName:    generator.CamelCase(svc.GetName()),
			CallMethodList: callMethodList,
		})
		p.P(out.String())
	}
}

func (p *protorpcPlugin) genServiceServer(
	file *generator.FileDescriptor,
	svc *descriptor.ServiceDescriptorProto,
) {
	const serviceHelperFunTmpl = `
// Accept{{.ServiceName}}Client accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func Accept{{.ServiceName}}Client(lis net.Listener, x {{.ServiceName}}) {
	srv := rpc.NewServer()
	if err := srv.RegisterName("{{.ServiceRegisterName}}", x); err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("lis.Accept(): %v\n", err)
		}
		go srv.ServeCodec(protorpc.NewServerCodec(conn))
	}
}
// Register{{.ServiceName}} publish the given {{.ServiceName}} implementation on the server.
func Register{{.ServiceName}}(srv *rpc.Server, x {{.ServiceName}}) error {
	if err := srv.RegisterName("{{.ServiceRegisterName}}", x); err != nil {
		return err
	}
	return nil
}
// New{{.ServiceName}}Server returns a new {{.ServiceName}} Server.
func New{{.ServiceName}}Server(x {{.ServiceName}}) *rpc.Server {
	srv := rpc.NewServer()
	if err := srv.RegisterName("{{.ServiceRegisterName}}", x); err != nil {
		log.Fatal(err)
	}
	return srv
}
// ListenAndServe{{.ServiceName}} listen announces on the local network address laddr
// and serves the given {{.ServiceName}} implementation.
func ListenAndServe{{.ServiceName}}(network, addr string, x {{.ServiceName}}) error {
	lis, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	defer lis.Close()
	srv := rpc.NewServer()
	if err := srv.RegisterName("{{.ServiceRegisterName}}", x); err != nil {
		return err
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("lis.Accept(): %v\n", err)
		}
		go srv.ServeCodec(protorpc.NewServerCodec(conn))
	}
}
`
	{
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(serviceHelperFunTmpl))
		t.Execute(out, &struct{ PackageName, ServiceName, ServiceRegisterName string }{
			PackageName: file.GetPackage(),
			ServiceName: generator.CamelCase(svc.GetName()),
			ServiceRegisterName: p.makeServiceRegisterName(
				file, file.GetPackage(), generator.CamelCase(svc.GetName()),
			),
		})
		p.P(out.String())
	}
}

func (p *protorpcPlugin) genServiceClient(
	file *generator.FileDescriptor,
	svc *descriptor.ServiceDescriptorProto,
) {
	const clientHelperFuncTmpl = `
type {{.ServiceName}}Client struct {
	*rpc.Client
}
// New{{.ServiceName}}Client returns a {{.ServiceName}} stub to handle
// requests to the set of {{.ServiceName}} at the other end of the connection.
func New{{.ServiceName}}Client(conn io.ReadWriteCloser) (*{{.ServiceName}}Client) {
	c := rpc.NewClientWithCodec(protorpc.NewClientCodec(conn))
	return &{{.ServiceName}}Client{c}
}
{{.MethodList}}
// Dial{{.ServiceName}} connects to an {{.ServiceName}} at the specified network address.
func Dial{{.ServiceName}}(network, addr string) (*{{.ServiceName}}Client, error) {
	c, err := protorpc.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return &{{.ServiceName}}Client{c}, nil
}
// Dial{{.ServiceName}}Timeout connects to an {{.ServiceName}} at the specified network address.
func Dial{{.ServiceName}}Timeout(network, addr string, timeout time.Duration) (*{{.ServiceName}}Client, error) {
	c, err := protorpc.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	return &{{.ServiceName}}Client{c}, nil
}
`
	const clientMethodTmpl = `
func (c *{{.ServiceName}}Client) {{.MethodName}}(in *{{.ArgsType}}) (out *{{.ReplyType}}, err error) {
	if in == nil {
		in = new({{.ArgsType}})
	}
	out = new({{.ReplyType}})
	if err = c.Call("{{.ServiceRegisterName}}.{{.MethodName}}", in, out); err != nil {
		return nil, err
	}
	return out, nil
}`

	// gen client method list
	var methodList string
	for _, m := range svc.Method {
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(clientMethodTmpl))
		t.Execute(out, &struct{ ServiceName, ServiceRegisterName, MethodName, ArgsType, ReplyType string }{
			ServiceName: generator.CamelCase(svc.GetName()),
			ServiceRegisterName: p.makeServiceRegisterName(
				file, file.GetPackage(), generator.CamelCase(svc.GetName()),
			),
			MethodName: generator.CamelCase(m.GetName()),
			ArgsType:   p.TypeName(p.ObjectNamed(m.GetInputType())),
			ReplyType:  p.TypeName(p.ObjectNamed(m.GetOutputType())),
		})
		methodList += out.String()
	}

	// gen all client code
	{
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(clientHelperFuncTmpl))
		t.Execute(out, &struct{ PackageName, ServiceName, MethodList string }{
			PackageName: file.GetPackage(),
			ServiceName: generator.CamelCase(svc.GetName()),
			MethodList:  methodList,
		})
		p.P(out.String())
	}
}

func (p *protorpcPlugin) makeServiceRegisterName(
	file *generator.FileDescriptor,
	packageName, serviceName string,
) string {
	if p.getGenericServicesOptionsUsePkgName(file) {
		return packageName + "." + serviceName
	}
	return serviceName
}
