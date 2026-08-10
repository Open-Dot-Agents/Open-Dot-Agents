package adapterprotocol

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type Client struct {
	command *exec.Cmd
	stdin   io.WriteCloser
	reader  *bufio.Reader
	stderr  cappedBuffer
	mu      sync.Mutex
	nextID  int64
	workDir string
}

type cappedBuffer struct {
	bytes.Buffer
	limit     int
	truncated bool
}

func (buffer *cappedBuffer) Write(data []byte) (int, error) {
	written := len(data)
	remaining := buffer.limit - buffer.Len()
	if remaining <= 0 {
		buffer.truncated = true
		return written, nil
	}
	if len(data) > remaining {
		data = data[:remaining]
		buffer.truncated = true
	}
	_, _ = buffer.Buffer.Write(data)
	return written, nil
}

func (buffer *cappedBuffer) message() string {
	message := buffer.String()
	if buffer.truncated {
		message += "\n[adapter stderr truncated]"
	}
	return message
}

func Start(ctx context.Context, executable string, args ...string) (*Client, error) {
	workDir, err := os.MkdirTemp("", "dota-adapter-host-")
	if err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = workDir
	stdin, err := command.StdinPipe()
	if err != nil {
		os.RemoveAll(workDir)
		return nil, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		stdin.Close()
		os.RemoveAll(workDir)
		return nil, err
	}
	client := &Client{command: command, stdin: stdin, reader: bufio.NewReader(stdout), workDir: workDir, stderr: cappedBuffer{limit: 1 << 20}}
	command.Stderr = &client.stderr
	command.Env = []string{"LANG=C.UTF-8", "LC_ALL=C.UTF-8"}
	if err := command.Start(); err != nil {
		stdin.Close()
		os.RemoveAll(workDir)
		return nil, err
	}
	return client, nil
}

func (c *Client) Call(method string, params, result any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	req := map[string]any{"jsonrpc": "2.0", "id": c.nextID, "method": method}
	if params != nil {
		req["params"] = params
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if err := WriteFrame(c.stdin, payload); err != nil {
		return err
	}
	responsePayload, err := ReadFrame(c.reader)
	if err != nil {
		return c.processError(err)
	}
	var resp struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      int64           `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   *RPCError       `json:"error"`
	}
	if err := json.Unmarshal(responsePayload, &resp); err != nil {
		return fmt.Errorf("invalid adapter response: %w", err)
	}
	if resp.JSONRPC != "2.0" || resp.ID != c.nextID {
		return fmt.Errorf("invalid adapter response id")
	}
	if resp.Error != nil {
		return resp.Error
	}
	if result != nil && len(resp.Result) > 0 && string(resp.Result) != "null" {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("invalid adapter result: %w", err)
		}
	}
	return nil
}

func (c *Client) Initialize(hostVersion string) error {
	var result InitializeResult
	if err := c.Call("initialize", InitializeParams{HostName: "dota", HostVersion: hostVersion, ProtocolVersions: []string{ProtocolVersion}}, &result); err != nil {
		return err
	}
	if result.ProtocolVersion != ProtocolVersion {
		return fmt.Errorf("adapter selected unsupported protocol %q", result.ProtocolVersion)
	}
	return nil
}

func (c *Client) Describe() (AdapterDescription, error) {
	var result AdapterDescription
	err := c.Call("describe", nil, &result)
	return result, err
}

func (c *Client) Operation(method string, params SnapshotParams) (OperationResult, error) {
	var result OperationResult
	err := c.Call(method, params, &result)
	return result, err
}

func (c *Client) Close() error {
	defer os.RemoveAll(c.workDir)
	_ = c.Call("shutdown", nil, nil)
	_ = c.stdin.Close()
	if err := c.command.Wait(); err != nil {
		return c.processError(err)
	}
	return nil
}

func (c *Client) processError(err error) error {
	if c.stderr.Len() > 0 {
		return fmt.Errorf("%w: %s", err, c.stderr.message())
	}
	return err
}
