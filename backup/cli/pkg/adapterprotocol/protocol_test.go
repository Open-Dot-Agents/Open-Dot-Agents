package adapterprotocol

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFrameRoundTrip(t *testing.T) {
	var buffer bytes.Buffer
	want := []byte(`{"jsonrpc":"2.0","id":1}`)
	if err := WriteFrame(&buffer, want); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFrame(bufio.NewReader(&buffer))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReadFrameRejectsMissingLength(t *testing.T) {
	_, err := ReadFrame(bufio.NewReader(strings.NewReader("X: y\r\n\r\n")))
	if err == nil {
		t.Fatal("expected missing length error")
	}
}

func TestReadFrameRejectsMalformedLengthsAndOversizedHeaders(t *testing.T) {
	tests := []string{
		"Content-Length: nope\r\n\r\n",
		"Content-Length: 1\r\nContent-Length: 1\r\n\r\nx",
		"Content-Length: 999999999\r\n\r\n",
		"X: " + strings.Repeat("a", 8<<10) + "\r\n\r\n",
	}
	for _, input := range tests {
		if _, err := ReadFrame(bufio.NewReader(strings.NewReader(input))); err == nil {
			t.Fatalf("expected framing error for %q", input[:min(len(input), 80)])
		}
	}
}

func TestClientReportsMalformedResponseAndCrash(t *testing.T) {
	for _, mode := range []string{"malformed", "crash"} {
		t.Run(mode, func(t *testing.T) {
			client := startHelperClient(t, context.Background(), mode)
			if err := client.Initialize("test"); err == nil {
				t.Fatal("expected adapter failure")
			}
			_ = client.Close()
		})
	}
}

func TestClientHonorsProcessDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	client := startHelperClient(t, ctx, "hang")
	started := time.Now()
	if err := client.Initialize("test"); err == nil {
		t.Fatal("expected deadline failure")
	}
	if elapsed := time.Since(started); elapsed > 3*time.Second {
		t.Fatalf("deadline took %s", elapsed)
	}
	_ = client.Close()
}

func TestClientRejectsProtocolMismatch(t *testing.T) {
	client := startHelperClient(t, context.Background(), "mismatch")
	defer client.Close()
	if err := client.Initialize("test"); err == nil || !strings.Contains(err.Error(), "unsupported protocol") {
		t.Fatalf("error = %v", err)
	}
}

func TestAdapterStderrBufferIsBounded(t *testing.T) {
	var buffer cappedBuffer
	buffer.limit = 8
	input := []byte("0123456789abcdef")
	written, err := buffer.Write(input)
	if err != nil || written != len(input) {
		t.Fatalf("write = %d, %v", written, err)
	}
	if buffer.Len() != 8 || !strings.Contains(buffer.message(), "truncated") {
		t.Fatalf("buffer = %q", buffer.message())
	}
}

func startHelperClient(t *testing.T, ctx context.Context, mode string) *Client {
	t.Helper()
	client, err := Start(ctx, os.Args[0], "-test.run=TestAdapterHelperProcess", "--", mode)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

type mismatchHandler struct{ testHandler }

func (mismatchHandler) Initialize(context.Context, InitializeParams) (InitializeResult, error) {
	return InitializeResult{ProtocolVersion: "9.0"}, nil
}

func TestAdapterHelperProcess(t *testing.T) {
	if len(os.Args) == 0 {
		return
	}
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "malformed":
		_, _ = ReadFrame(bufio.NewReader(os.Stdin))
		_, _ = fmt.Fprint(os.Stdout, "X: broken\r\n\r\n")
	case "crash":
		_, _ = ReadFrame(bufio.NewReader(os.Stdin))
		_, _ = fmt.Fprintln(os.Stderr, "adapter crashed")
		os.Exit(9)
	case "hang":
		time.Sleep(time.Hour)
	case "mismatch":
		if err := Serve(context.Background(), os.Stdin, os.Stdout, mismatchHandler{}); err != nil {
			os.Exit(10)
		}
	}
}

type testHandler struct{}

func (testHandler) Initialize(context.Context, InitializeParams) (InitializeResult, error) {
	return InitializeResult{ProtocolVersion: ProtocolVersion}, nil
}
func (testHandler) Describe(context.Context) (AdapterDescription, error) {
	return AdapterDescription{ID: "org.example.test", ProtocolVersion: ProtocolVersion}, nil
}
func (testHandler) Validate(context.Context, SnapshotParams) (OperationResult, error) {
	return OperationResult{}, nil
}
func (testHandler) ExportPlan(context.Context, SnapshotParams) (OperationResult, error) {
	return OperationResult{}, nil
}
func (testHandler) ImportPlan(context.Context, SnapshotParams) (OperationResult, error) {
	return OperationResult{}, nil
}
