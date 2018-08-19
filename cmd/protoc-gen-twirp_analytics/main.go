package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/twitchtv/protogen"
	"github.com/twitchtv/protogen/stringutils"
	"github.com/twitchtv/protogen/typemap"
)

const version = "v0.0.1"

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	g := newGenerator()
	protogen.RunProtocPlugin(g)
}

func newGenerator() *generator {
	return &generator{output: new(bytes.Buffer)}
}

type generator struct {
	reg    *typemap.Registry
	output *bytes.Buffer
}

func (g *generator) Generate(in *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	genFiles, err := protogen.FilesToGenerate(in)
	if err != nil {
		return nil, err
	}
	g.reg = typemap.New(in.ProtoFile)

	resp := new(plugin.CodeGeneratorResponse)
	for _, f := range genFiles {
		respFile := g.generateFile(f)
		if respFile != nil {
			resp.File = append(resp.File, respFile)
		}
	}
	return resp, nil
}

func (g *generator) generateFile(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	g.P("// Code generated by protoc-gen-twirp_analytics ", version, ", DO NOT EDIT.")
	g.P("// source: ", file.GetName())
	g.P("")

	pkgname, _ := goPackageName(file)

	g.P("package ", pkgname)
	g.P()

	g.P(`import "context"`)
	g.P(`import "github.com/Xe/ln"`)
	g.P(`import "gopkg.in/segmentio/analytics-go.v3"`)
	g.P()

	for _, service := range file.Service {
		g.generateProtobufAnalytics(file, service)
	}

	resp := new(plugin.CodeGeneratorResponse_File)
	resp.Name = proto.String(goFileName(file))
	resp.Content = proto.String(g.output.String())
	g.output.Reset()

	return resp
}

func (g *generator) generateProtobufAnalytics(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	svcName := serviceName(service) + "Analytics"
	g.P("// ", svcName, " is a middleware for ", serviceName(service), " that collects timing and error rate data for servers.")
	g.P("type ", svcName, " struct {")
	g.P("\tNext ", serviceName(service))
	g.P("\tclient analytics.Client")
	g.P("}")
	g.P()

	g.P("func New", svcName, "(next ", serviceName(service), ", client analytics.Client) ", serviceName(service), " {")
	g.P("\tvar result ", svcName)
	g.P("\tresult.Next = next")
	g.P("\tresult.client = client")
	g.P("\treturn result")
	g.P("}")
	g.P()

	for _, method := range service.Method {
		methName := methodName(method)
		miType := methodInputName(method)
		moType := methodOutputName(method)
		g.P("func (i ", svcName, ") ", methName, "(ctx context.Context, input *", miType, ") (result *", moType, ", err error) {")
		g.P("\tvar track analytics.Track")
		g.P("\ttrack.Event = ", `"`, serviceName(service), " ", methName, `"`)
		g.P("\ttrack.UserId = ln.GetFFromContext(ctx)[\"x_forwarded_for\"].(string)")
		g.P("\tdefer func() {")
		g.P("\t\tif err != nil {")
		g.P("\t\t\ttrack.Event += ", `" Error"`)
		g.P("\t\t}")
		g.P("\t\terr = i.client.Enqueue(track)")
		g.P("\t\tif err != nil {")
		g.P("\t\t\tln.Error(ctx, err)")
		g.P("\t\t}")
		g.P("\t}()")
		g.P()
		g.P("\tresult, err = i.Next.", methName, "(ctx, input)")
		g.P("\treturn")
		g.P("}")
		g.P()
	}
}

func (g *generator) P(args ...string) {
	for _, v := range args {
		g.output.WriteString(v)
	}
	g.output.WriteByte('\n')
}

func (g *generator) printComments(comments typemap.DefinitionComments, prefix string) {
	text := strings.TrimSuffix(comments.Leading, "\n")
	for _, line := range strings.Split(text, "\n") {
		g.P(prefix, strings.TrimPrefix(line, " "))
	}
}

func serviceName(service *descriptor.ServiceDescriptorProto) string {
	return stringutils.CamelCase(service.GetName())
}

func clientName(service *descriptor.ServiceDescriptorProto) string {
	return serviceName(service) + "Client"
}

func fullServiceName(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	name := serviceName(service)
	if pkg := file.GetPackage(); pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func methodName(method *descriptor.MethodDescriptorProto) string {
	return method.GetName()
}

// methodOutputName returns the basename of the output type of a method.
func methodOutputName(meth *descriptor.MethodDescriptorProto) string {
	fullName := meth.GetOutputType()
	split := strings.Split(fullName, ".")
	return split[len(split)-1]
}

// methodInputName returns the basename of the input type of a method.
func methodInputName(meth *descriptor.MethodDescriptorProto) string {
	fullName := meth.GetInputType()
	split := strings.Split(fullName, ".")
	return split[len(split)-1]
}
