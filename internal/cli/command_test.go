package cli

import (
	"bytes"
	"strings"
	"testing"
)

// newTestCtx returns a Context with buffered output for assertions.
func newTestCtx() (*Context, *bytes.Buffer, *bytes.Buffer) {
	var out, errb bytes.Buffer
	return &Context{Stdout: &out, Stderr: &errb}, &out, &errb
}

func TestDispatchRunsLeaf(t *testing.T) {
	ctx, out, _ := newTestCtx()
	root := &Command{Name: "gw", Sub: []*Command{
		{Name: "ping", Run: func(ctx *Context, args []string) error {
			ctx.Stdout.Write([]byte("pong"))
			return nil
		}},
	}}
	if err := root.dispatch(ctx, nil, []string{"ping"}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if got := out.String(); got != "pong" {
		t.Fatalf("got %q, want %q", got, "pong")
	}
}

func TestDispatchUnknownCommandIsError(t *testing.T) {
	ctx, _, _ := newTestCtx()
	root := buildRoot()
	err := root.dispatch(ctx, nil, []string{"bogus"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	var ce *Error
	if !asError(err, &ce) || ce.Code != "unknown_command" {
		t.Fatalf("want unknown_command error, got %v", err)
	}
}

func TestGroupWithoutSubcommandPrintsHelp(t *testing.T) {
	ctx, out, _ := newTestCtx()
	root := buildRoot()
	if err := root.dispatch(ctx, nil, []string{"ticket"}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	for _, want := range []string{"create", "transition", "tree"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("ticket help missing %q:\n%s", want, out.String())
		}
	}
}

func TestHelpArgPrintsHelp(t *testing.T) {
	ctx, out, _ := newTestCtx()
	root := buildRoot()
	if err := root.dispatch(ctx, nil, []string{"help"}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if !strings.Contains(out.String(), "Commands:") {
		t.Fatalf("help output missing command list:\n%s", out.String())
	}
}

// asError is a tiny errors.As shim kept local to avoid importing errors in the
// test for a single call.
func asError(err error, target **Error) bool {
	e, ok := err.(*Error)
	if ok {
		*target = e
	}
	return ok
}
