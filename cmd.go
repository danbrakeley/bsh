package bsh

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/danbrakeley/commandline"
)

// Command is intended to be built via the Cmd() funcs.
// Generally you call a "starter", which returns a *Command,
// then call zero or more "modifiers" to tweak the *Command,
// then call a "runner" to actually run the *Command.

type Command struct {
	raw        string
	dir        string
	env        []string
	in         io.Reader // the stdin to attach to this process
	out        io.Writer // the stdout to attach to this process
	err        io.Writer // the stderr to attach to this process
	exitStatus *int      // exit status code

	// copied from Bsh at creation
	b *Bsh
}

// Command starters

func (b *Bsh) Cmd(command string) *Command {
	return &Command{
		raw: command,
		in:  b.ensureStdin(),
		out: b.ensureStdout(),
		err: b.ensureStderr(),
		b:   b,
	}
}

func (b *Bsh) Cmdf(format string, args ...interface{}) *Command {
	return b.Cmd(fmt.Sprintf(format, args...))
}

// Command methods

func (c *Command) StdIn() io.Reader {
	return c.in
}

func (c *Command) StdOut() io.Writer {
	return c.out
}

func (c *Command) StdErr() io.Writer {
	return c.err
}

func (c *Command) In(r io.Reader) *Command {
	c.in = r
	return c
}

func (c *Command) Out(w io.Writer) *Command {
	c.out = w
	return c
}

func (c *Command) Err(w io.Writer) *Command {
	c.err = w
	return c
}

func (c *Command) OutErr(w io.Writer) *Command {
	c.out = w
	c.err = w
	return c
}

func (c *Command) ExitStatus(n *int) *Command {
	c.exitStatus = n
	return c
}

// ExpandEnv calls os.ExpandEnv on the command string before it is parsed and passed to exec.Cmd.
func (c *Command) ExpandEnv() *Command {
	c.raw = os.ExpandEnv(c.raw)
	return c
}

// Env adds environment variables in the form "KEY=VALUE", to be set on exec.Cmd.Env.
// Note: these env vars are not seen by ExpandEnv.
func (c *Command) Env(vars ...string) *Command {
	c.env = append(c.env, vars...)
	return c
}

// Dir sets the working directory
func (c *Command) Dir(dir string) *Command {
	c.dir = dir
	return c
}

// Command runners

func (c *Command) Run() {
	if err := c.run(); err != nil {
		c.b.Warnf("unexpected error in %s", c.raw)
		c.b.Panic(err)
	}
}

func (c *Command) RunStr() string {
	var b strings.Builder
	c.out = &b
	c.err = &b
	if err := c.run(); err != nil {
		c.b.Warnf("unexpected error in %s", c.raw)
		c.b.Panic(err)
	}
	return b.String()
}

func (c *Command) RunErr() error {
	return c.run()
}

func (c *Command) RunExitStatus() int {
	n, err := extractExitStatus(c.run())
	if err != nil {
		c.b.Warnf("unexpected error in %s", c.raw)
		c.b.Panic(err)
	}
	return n
}

func (c *Command) Bash() {
	if err := c.bash(); err != nil {
		c.b.Warnf("unexpected error in bash -c %s", c.raw)
		c.b.Panic(err)
	}
}

func (c *Command) BashStr() string {
	var b strings.Builder
	c.out = &b
	c.err = &b
	if err := c.bash(); err != nil {
		c.b.Warnf("unexpected error in bash -c %s", c.raw)
		c.b.Panic(err)
	}
	return b.String()
}

func (c *Command) BashErr() error {
	return c.bash()
}

func (c *Command) BashExitStatus() int {
	n, err := extractExitStatus(c.bash())
	if err != nil {
		c.b.Warnf("unexpected error in bash -c %s", c.raw)
		c.b.Panic(err)
	}
	return n
}

// helpers

func (c *Command) run() error {
	args, err := commandline.Parse(c.raw)
	if err != nil {
		return err
	}
	c.b.Verbosef("Exec: %s", c.raw)
	cmd := exec.Command(args[0], args[1:]...)
	if len(c.env) > 0 {
		c.b.Verbosef("+Env: %v", c.env)
		cmd.Env = append(os.Environ(), c.env...)
	}
	cmd.Dir = c.dir
	cmd.Stdin = c.in
	cmd.Stdout = c.out
	cmd.Stderr = c.err
	err = cmd.Run()
	if c.exitStatus != nil {
		n, e := extractExitStatus(err)
		if e == nil {
			*c.exitStatus = n
		}
	}
	return err
}

func (c *Command) bash() error {
	c.b.Verbosef("Bash: %s", c.raw)
	cmd := exec.Command("bash", "-c", c.raw)
	if len(c.env) > 0 {
		c.b.Verbosef("+Env: %v", c.env)
		cmd.Env = append(os.Environ(), c.env...)
	}
	cmd.Dir = c.dir
	cmd.Stdin = c.in
	cmd.Stdout = c.out
	cmd.Stderr = c.err
	err := cmd.Run()
	if c.exitStatus != nil {
		n, e := extractExitStatus(err)
		if e == nil {
			*c.exitStatus = n
		}
	}
	return err
}

func extractExitStatus(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		return -1, err
	}
	return ee.ExitCode(), nil
}
