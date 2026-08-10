// Package adapterprotocol defines the language-neutral dota adapter protocol.
package adapterprotocol

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	ProtocolVersion = "1.0"
	MaxMessageBytes = 32 << 20
)

type Diagnostic struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}

type Loss struct {
	Path      string `json:"path"`
	Reason    string `json:"reason"`
	Preserved string `json:"preservedAt,omitempty"`
	Severity  string `json:"severity"`
}

type File struct {
	Path     string `json:"path"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

type Plan struct {
	Files []File `json:"files"`
}

type SnapshotParams struct {
	Files []File `json:"files"`
}

type OperationResult struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	Losses      []Loss       `json:"losses"`
	Plan        *Plan        `json:"plan,omitempty"`
}

type InitializeParams struct {
	HostName         string   `json:"hostName"`
	HostVersion      string   `json:"hostVersion"`
	ProtocolVersions []string `json:"protocolVersions"`
}

type InitializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
}

type AdapterDescription struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	ProtocolVersion  string            `json:"protocolVersion"`
	Target           string            `json:"target"`
	Capabilities     []string          `json:"capabilities"`
	CategoryStatuses map[string]string `json:"categoryStatuses"`
	InputPatterns    []string          `json:"inputPatterns"`
	MaxSnapshotBytes int64             `json:"maxSnapshotBytes,omitempty"`
}

type Handler interface {
	Initialize(context.Context, InitializeParams) (InitializeResult, error)
	Describe(context.Context) (AdapterDescription, error)
	Validate(context.Context, SnapshotParams) (OperationResult, error)
	ExportPlan(context.Context, SnapshotParams) (OperationResult, error)
	ImportPlan(context.Context, SnapshotParams) (OperationResult, error)
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *RPCError) Error() string { return e.Message }

func ReadFrame(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	headerBytes := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		headerBytes += len(line)
		if headerBytes > 8<<10 {
			return nil, errors.New("protocol headers exceed limit")
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid protocol header %q", line)
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			if contentLength >= 0 {
				return nil, errors.New("duplicate Content-Length")
			}
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || parsed < 0 || parsed > MaxMessageBytes {
				return nil, errors.New("invalid Content-Length")
			}
			contentLength = parsed
		}
	}
	if contentLength < 0 {
		return nil, errors.New("missing Content-Length")
	}
	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func WriteFrame(writer io.Writer, payload []byte) error {
	if len(payload) > MaxMessageBytes {
		return errors.New("protocol message exceeds limit")
	}
	if _, err := fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return err
	}
	_, err := writer.Write(payload)
	return err
}

func Serve(ctx context.Context, input io.Reader, output io.Writer, handler Handler) error {
	reader := bufio.NewReader(input)
	for {
		payload, err := ReadFrame(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		var req request
		if err := json.Unmarshal(payload, &req); err != nil || req.JSONRPC != "2.0" || req.Method == "" {
			if err := writeResponse(output, response{JSONRPC: "2.0", Error: &RPCError{Code: -32600, Message: "invalid request"}}); err != nil {
				return err
			}
			continue
		}
		resp := response{JSONRPC: "2.0", ID: req.ID}
		var callErr error
		switch req.Method {
		case "initialize":
			var params InitializeParams
			if callErr = decodeParams(req.Params, &params); callErr == nil {
				resp.Result, callErr = handler.Initialize(ctx, params)
			}
		case "describe":
			resp.Result, callErr = handler.Describe(ctx)
		case "validate", "exportPlan", "importPlan":
			var params SnapshotParams
			if callErr = decodeParams(req.Params, &params); callErr == nil {
				switch req.Method {
				case "validate":
					resp.Result, callErr = handler.Validate(ctx, params)
				case "exportPlan":
					resp.Result, callErr = handler.ExportPlan(ctx, params)
				default:
					resp.Result, callErr = handler.ImportPlan(ctx, params)
				}
			}
		case "shutdown":
			resp.Result = nil
			if err := writeResponse(output, resp); err != nil {
				return err
			}
			return nil
		default:
			resp.Error = &RPCError{Code: -32601, Message: "method not found"}
		}
		if callErr != nil {
			code := -32000
			if strings.Contains(callErr.Error(), "parameters") {
				code = -32602
			}
			resp.Result = nil
			resp.Error = &RPCError{Code: code, Message: callErr.Error()}
		}
		if err := writeResponse(output, resp); err != nil {
			return err
		}
	}
}

func decodeParams(raw json.RawMessage, target any) error {
	if len(raw) == 0 {
		return errors.New("invalid parameters")
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}
	return nil
}

func writeResponse(output io.Writer, resp response) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return WriteFrame(output, payload)
}
