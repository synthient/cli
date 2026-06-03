package grpcschema

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const DefaultEndpoint = "grpc.synthient.com:443"

type Options struct {
	Endpoint          string
	APIKey            string
	Plaintext         bool
	IncludeReflection bool
	Symbols           []string
}

type Result struct {
	Endpoint      string
	Symbols       []string
	DescriptorSet *descriptorpb.FileDescriptorSet
}

func Fetch(ctx context.Context, options Options) (Result, error) {
	endpoint, host, err := NormalizeEndpoint(options.Endpoint)
	if err != nil {
		return Result{}, err
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32 << 20)),
	}
	if options.Plaintext {
		dialOptions[0] = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.NewClient(endpoint, dialOptions...)
	if err != nil {
		return Result{}, fmt.Errorf("creating grpc client: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if strings.TrimSpace(options.APIKey) != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", options.APIKey)
	}

	client := reflectpb.NewServerReflectionClient(conn)
	symbols := cleanSymbols(options.Symbols)
	if len(symbols) == 0 {
		symbols, err = listServices(ctx, client, host, options.IncludeReflection)
		if err != nil {
			return Result{}, err
		}
	}
	if len(symbols) == 0 {
		return Result{}, errors.New("no protobuf services returned by reflection")
	}

	files := map[string]*descriptorpb.FileDescriptorProto{}
	for _, symbol := range symbols {
		rawFiles, err := fileContainingSymbol(ctx, client, host, symbol)
		if err != nil {
			return Result{}, err
		}
		err = addFiles(files, rawFiles)
		if err != nil {
			return Result{}, err
		}
	}
	err = resolveDependencies(ctx, client, host, files)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Endpoint:      endpoint,
		Symbols:       symbols,
		DescriptorSet: &descriptorpb.FileDescriptorSet{File: orderedFiles(files)},
	}, nil
}

func NormalizeEndpoint(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = DefaultEndpoint
	}
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", "", fmt.Errorf("parsing grpc endpoint: %w", err)
		}
		raw = parsed.Host
		if raw == "" {
			raw = strings.Trim(parsed.Path, "/")
		}
	}
	raw = strings.TrimSuffix(raw, "/")
	if raw == "" {
		return "", "", errors.New("empty grpc endpoint")
	}

	host, _, err := net.SplitHostPort(raw)
	if err == nil {
		return raw, strings.Trim(host, "[]"), nil
	}
	if !strings.Contains(raw, ":") {
		return net.JoinHostPort(raw, "443"), raw, nil
	}
	return raw, strings.Trim(raw, "[]"), nil
}

func Explain(err error) string {
	statusValue, ok := status.FromError(err)
	if !ok {
		return ""
	}
	switch statusValue.Code() {
	case codes.Unauthenticated:
		return "Missing or invalid API key. Run `synthient auth`, set SYNTHIENT_API_KEY, or add SYNTHIENT_API_KEY to .env."
	case codes.PermissionDenied:
		return "This API key is authenticated but does not have permission to inspect gRPC schemas."
	case codes.NotFound:
		return "No protobuf schema was found for the requested symbol. Run `synthient grpc schema` without arguments to discover exported services."
	case codes.Unimplemented:
		return "The gRPC endpoint does not expose server reflection."
	case codes.Unavailable:
		return "Could not reach the gRPC endpoint. Check --endpoint, network access, and TLS settings."
	case codes.DeadlineExceeded:
		return "Timed out while reading protobuf schemas from gRPC reflection. Increase --timeout or retry."
	default:
		return ""
	}
}

func listServices(ctx context.Context, client reflectpb.ServerReflectionClient, host string, includeReflection bool) ([]string, error) {
	response, err := reflectionRequest(ctx, client, host, &reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{ListServices: ""},
	})
	if err != nil {
		return nil, err
	}
	list := response.GetListServicesResponse()
	if list == nil {
		return nil, errors.New("reflection response did not include service list")
	}
	services := []string{}
	for _, service := range list.Service {
		name := strings.TrimSpace(service.GetName())
		if name == "" {
			continue
		}
		if !includeReflection && strings.HasPrefix(name, "grpc.reflection.") {
			continue
		}
		services = append(services, name)
	}
	sort.Strings(services)
	return services, nil
}

func fileContainingSymbol(ctx context.Context, client reflectpb.ServerReflectionClient, host string, symbol string) ([][]byte, error) {
	response, err := reflectionRequest(ctx, client, host, &reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_FileContainingSymbol{FileContainingSymbol: symbol},
	})
	if err != nil {
		return nil, err
	}
	descriptorResponse := response.GetFileDescriptorResponse()
	if descriptorResponse == nil {
		return nil, fmt.Errorf("reflection response did not include descriptors for %s", symbol)
	}
	return descriptorResponse.FileDescriptorProto, nil
}

func fileByName(ctx context.Context, client reflectpb.ServerReflectionClient, host string, name string) ([][]byte, error) {
	response, err := reflectionRequest(ctx, client, host, &reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{FileByFilename: name},
	})
	if err != nil {
		return nil, err
	}
	descriptorResponse := response.GetFileDescriptorResponse()
	if descriptorResponse == nil {
		return nil, fmt.Errorf("reflection response did not include descriptors for %s", name)
	}
	return descriptorResponse.FileDescriptorProto, nil
}

func reflectionRequest(ctx context.Context, client reflectpb.ServerReflectionClient, host string, request *reflectpb.ServerReflectionRequest) (*reflectpb.ServerReflectionResponse, error) {
	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}
	request.Host = host
	err = stream.Send(request)
	if err != nil {
		return nil, err
	}
	response, err := stream.Recv()
	_ = stream.CloseSend()
	if err != nil {
		return nil, err
	}
	if errorResponse := response.GetErrorResponse(); errorResponse != nil {
		return nil, status.Error(codes.Code(errorResponse.ErrorCode), errorResponse.ErrorMessage)
	}
	return response, nil
}

func addFiles(files map[string]*descriptorpb.FileDescriptorProto, rawFiles [][]byte) error {
	for _, rawFile := range rawFiles {
		file := descriptorpb.FileDescriptorProto{}
		err := proto.Unmarshal(rawFile, &file)
		if err != nil {
			return fmt.Errorf("decoding file descriptor: %w", err)
		}
		if file.GetName() == "" {
			return errors.New("reflection returned a file descriptor without a name")
		}
		if _, ok := files[file.GetName()]; !ok {
			files[file.GetName()] = &file
		}
	}
	return nil
}

func resolveDependencies(ctx context.Context, client reflectpb.ServerReflectionClient, host string, files map[string]*descriptorpb.FileDescriptorProto) error {
	for {
		missing := missingDependencies(files)
		if len(missing) == 0 {
			return nil
		}
		for _, name := range missing {
			rawFiles, err := fileByName(ctx, client, host, name)
			if err != nil {
				return err
			}
			err = addFiles(files, rawFiles)
			if err != nil {
				return err
			}
			if _, ok := files[name]; !ok {
				return fmt.Errorf("reflection did not return missing dependency %s", name)
			}
		}
	}
}

func missingDependencies(files map[string]*descriptorpb.FileDescriptorProto) []string {
	seen := map[string]bool{}
	missing := []string{}
	for _, file := range files {
		for _, dependency := range file.Dependency {
			if _, ok := files[dependency]; ok || seen[dependency] {
				continue
			}
			seen[dependency] = true
			missing = append(missing, dependency)
		}
	}
	sort.Strings(missing)
	return missing
}

func orderedFiles(files map[string]*descriptorpb.FileDescriptorProto) []*descriptorpb.FileDescriptorProto {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	ordered := []*descriptorpb.FileDescriptorProto{}
	visited := map[string]bool{}
	visiting := map[string]bool{}
	var visit func(string)
	visit = func(name string) {
		if visited[name] || visiting[name] {
			return
		}
		file, ok := files[name]
		if !ok {
			return
		}
		visiting[name] = true
		dependencies := append([]string{}, file.Dependency...)
		sort.Strings(dependencies)
		for _, dependency := range dependencies {
			visit(dependency)
		}
		visiting[name] = false
		visited[name] = true
		ordered = append(ordered, file)
	}
	for _, name := range names {
		visit(name)
	}
	return ordered
}

func cleanSymbols(symbols []string) []string {
	seen := map[string]bool{}
	cleaned := []string{}
	for _, symbol := range symbols {
		symbol = strings.TrimSpace(symbol)
		if symbol == "" || seen[symbol] {
			continue
		}
		seen[symbol] = true
		cleaned = append(cleaned, symbol)
	}
	return cleaned
}
