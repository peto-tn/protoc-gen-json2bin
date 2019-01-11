package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// parse the standard input from protoc
func parseReq(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var req plugin.CodeGeneratorRequest
	if err = proto.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// parse arguments from protoc
func parseArguments(req *plugin.CodeGeneratorRequest) (string, string, string, error) {
	parameter := req.GetParameter()
	if parameter == "" {
		return "", "", "", errors.New("missing required parameters")
	}

	var jsonPath string = ""
	var messageType string = ""
	var outputFile string = ""
	for _, p := range strings.Split(parameter, ",") {
		spec := strings.SplitN(p, "=", 2)
		if len(spec) == 1 {
			continue
		}
		name, value := spec[0], spec[1]
		if name == "json" {
			jsonPath = value
		} else if name == "message_type" {
			messageType = value
		} else if name == "output" {
			outputFile = value
		}
	}
	if jsonPath == "" {
		return "", "", "", errors.New("missing required parameter [json]")
	}
	if messageType == "" {
		return "", "", "", errors.New("missing required parameter [message_type]")
	}
	if outputFile == "" {
		fileName := filepath.Base(jsonPath[0 : len(jsonPath)-len(filepath.Ext(jsonPath))])
		outputFile = fileName + ".pb"
	}

	return jsonPath, messageType, outputFile, nil
}

// parse the protocol buffers defeinition
func parseProto(req *plugin.CodeGeneratorRequest, messageType string) (*dynamic.Message, error) {
	var md *desc.MessageDescriptor = nil
	for _, f := range req.ProtoFile {
		pb := proto.Clone(f).(*descriptor.FileDescriptorProto)
		pb.SourceCodeInfo = nil

		b, err := proto.Marshal(pb)
		if err != nil {
			return nil, err
		}

		var buff bytes.Buffer
		w, _ := gzip.NewWriterLevel(&buff, gzip.BestCompression)
		w.Write(b)
		w.Close()
		b = buff.Bytes()

		proto.RegisterFile(*f.Name, b)
		protoFd, err := desc.LoadFileDescriptor(*f.Name)
		if err != nil {
			return nil, err
		}
		if md == nil {
			md = protoFd.FindMessage(messageType)
		}
	}

	if md == nil {
		return nil, errors.New("not find message type [" + messageType + "]")
	}
	return dynamic.NewMessage(md), nil
}

// parse data in json format
func parseJson(dm *dynamic.Message, jsonPath string) error {
	// read json file
	js, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	err = dm.UnmarshalJSON([]byte(js))
	if err != nil {
		return err
	}

	return nil
}

// create a response for protoc
func createResp(dm *dynamic.Message, outputFile string) (*plugin.CodeGeneratorResponse, error) {
	buf, err := proto.Marshal(dm)
	if err != nil {
		return nil, err
	}

	var resp plugin.CodeGeneratorResponse
	resp.File = append(resp.File, &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(outputFile),
		Content: proto.String(string(buf)),
	})
	return &resp, nil
}

// emit response to standard output
func emitResp(resp *plugin.CodeGeneratorResponse) error {
	buf, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(buf)
	return err
}

func run() error {
	req, err := parseReq(os.Stdin)
	if err != nil {
		return err
	}

	jsonPath, messageType, outputFile, err := parseArguments(req)
	if err != nil {
		return err
	}

	md, err := parseProto(req, messageType)
	if err != nil {
		return err
	}

	err = parseJson(md, jsonPath)
	if err != nil {
		return err
	}

	resp, err := createResp(md, outputFile)
	if err != nil {
		return err
	}

	return emitResp(resp)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
