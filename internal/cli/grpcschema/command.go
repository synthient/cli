package grpcschema

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/go-synthient/v2"
	"go.mattglei.ch/timber"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type serviceSummary struct {
	Name    string
	Methods []methodSummary
}

type methodSummary struct {
	Name   string
	Input  string
	Output string
}

var Command = &cobra.Command{
	Use:   "grpc",
	Short: "Inspect gRPC services and protobuf schemas",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			timber.Fatal(err, "output help")
		}
	},
}

var SchemaCommand = &cobra.Command{
	Use:     "schema [symbol...]",
	Aliases: []string{"proto", "protobuf", "protoset"},
	Short:   "Output protobuf descriptors using gRPC reflection",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !slices.Contains(formats, flags.format) {
			timber.FatalMsg("invalid output format", timber.A("value", flags.format), timber.A("valid", strings.Join(formats, "|")))
		}

		config, err := conf.Read()
		if err != nil {
			app.Fatal(err, "failed to read configuration file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			app.Fatal(err, "failed to create synthient client")
		}
		config.ApplyToClient(&client)

		endpoint := strings.TrimSpace(flags.endpoint)
		if endpoint == "" {
			endpoint = config.GRPCEndpoint(synthient.DefaultGRPCEndpoint)
		}

		ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
		defer cancel()

		result, err := client.GRPCSchema(ctx, &synthient.GRPCSchemaOptions{
			Endpoint:          endpoint,
			Plaintext:         flags.plaintext,
			IncludeReflection: flags.includeReflection,
			Symbols:           args,
		})
		if err != nil {
			if msg := synthient.ExplainGRPCError(err); msg != "" {
				timber.FatalMsg(msg)
			}
			app.Fatal(err, "failed to fetch protobuf schemas")
		}

		out, closeOut := output.Open(flags.output)
		defer closeOut()

		switch flags.format {
		case "text":
			writeText(out, output.NewStyles(out), result)
		case "json":
			data, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(result.DescriptorSet)
			if err != nil {
				app.Fatal(err, "failed to marshal descriptor set")
			}
			writeBytes(out, append(data, '\n'))
		case "binpb":
			data, err := proto.Marshal(result.DescriptorSet)
			if err != nil {
				app.Fatal(err, "failed to marshal descriptor set")
			}
			writeBytes(out, data)
		}
	},
}

func init() {
	Command.AddCommand(SchemaCommand)
}

func writeBytes(out *os.File, data []byte) {
	_, err := out.Write(data)
	if err != nil {
		timber.Fatal(err, "failed to write output")
	}
}

func writeText(out *os.File, styles output.Styles, result synthient.GRPCSchemaResult) {
	output.Header(out, styles, "Protobuf Schema", result.Endpoint)
	output.Divider(out, styles)
	blocks := []output.Block{
		{
			Name: "Reflection",
			Values: []output.BlockValue{
				{Key: "Endpoint", Value: result.Endpoint},
				{Key: "Symbols", Value: result.Symbols},
				{Key: "Files", Value: len(result.DescriptorSet.File)},
			},
		},
	}

	files := output.Block{Name: "Files"}
	for _, file := range result.DescriptorSet.File {
		files.Values = append(files.Values, output.BlockValue{Key: file.GetName(), Value: file.GetPackage()})
	}
	blocks = append(blocks, files)

	services := summarizeServices(result.DescriptorSet)
	if len(services) > 0 {
		serviceBlock := output.Block{Name: "Services"}
		methodBlock := output.Block{Name: "Methods"}
		for _, service := range services {
			serviceBlock.Values = append(serviceBlock.Values, output.BlockValue{Key: service.Name, Value: fmt.Sprintf("%d methods", len(service.Methods))})
			for _, method := range service.Methods {
				methodBlock.Values = append(methodBlock.Values, output.BlockValue{
					Key:   fmt.Sprintf("%s.%s", service.Name, method.Name),
					Value: fmt.Sprintf("%s -> %s", method.Input, method.Output),
				})
			}
		}
		blocks = append(blocks, serviceBlock, methodBlock)
	}

	for i, block := range blocks {
		block.Output(out, styles, 0, i+1 == len(blocks))
		if i+1 != len(blocks) {
			output.Divider(out, styles)
		}
	}
}

func summarizeServices(set *descriptorpb.FileDescriptorSet) []serviceSummary {
	services := []serviceSummary{}
	for _, file := range set.File {
		pkg := file.GetPackage()
		for _, service := range file.Service {
			name := service.GetName()
			if pkg != "" {
				name = pkg + "." + name
			}
			summary := serviceSummary{Name: name}
			for _, method := range service.Method {
				summary.Methods = append(summary.Methods, methodSummary{
					Name:   method.GetName(),
					Input:  trimType(method.GetInputType()),
					Output: trimType(method.GetOutputType()),
				})
			}
			services = append(services, summary)
		}
	}
	sort.Slice(services, func(i int, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services
}

func trimType(name string) string {
	return strings.TrimPrefix(name, ".")
}
