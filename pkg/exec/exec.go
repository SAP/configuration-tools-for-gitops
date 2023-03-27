package exec

import (
	"context"
	"fmt"
	"os"
	ex "os/exec"
	"strings"
)

type Exec interface {
	Command(name string, arg ...Input) Command
}

type Command interface {
	CombinedOutput() ([]byte, error)
	Output() ([]byte, error)
	Run() error
	RunDynamic() error
}

type Input interface {
	Parse() string
	Debug() string
	IsPrivate() bool
}

type Logger interface {
	Debugf(template string, args ...interface{})
	Debug(msg ...interface{})
}

func Private(value string, debugOutput ...string) Input {
	return input{value, true, strings.Join(debugOutput, " ")}
}

func Public(values ...string) []Input {
	res := make([]Input, 0, len(values))
	for _, v := range values {
		res = append(res, input{value: v})
	}
	return res
}
func Public1(value string) Input {
	return input{value: value}
}

type execute struct {
	ctx        context.Context
	workingDir string
	envVars    []Input
	logger     Logger
}

func New(l Logger, ctx context.Context, workingDir string, envVars ...Input) Exec {
	return execute{
		ctx,
		workingDir,
		envVars,
		l,
	}
}

func NewHome(l Logger, ctx context.Context, workingDir string, envVars ...Input) (Exec, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return New(l, ctx, workingDir, append(envVars, Public1(fmt.Sprintf("HOME=%s", home)))...), nil
}

func (e execute) DebugOutput(name string, args []Input) string {
	privateEnv := false
	for _, e := range e.envVars {
		if e.IsPrivate() {
			privateEnv = true
			break
		}
	}
	var cmd string
	if privateEnv {
		cmd = fmt.Sprintf("sh -c '%s %s'", name, strings.Join(debug(args), " "))
	} else {
		cmd = fmt.Sprintf("%s %s", name, strings.Join(debug(args), " "))
	}
	return fmt.Sprintf("workdir: %q -- exec: \"%+v   %s\"",
		e.workingDir,
		strings.Join(debug(e.envVars), " "),
		cmd,
	)
}

func (e execute) Command(name string, args ...Input) Command {
	e.logger.Debug(e.DebugOutput(name, args))

	// (https://github.com/securego/gosec/blob/ea6d49d1b5ae4945cdd856f80e52e3ebba216019/rules/subproc.go)
	// not an issue here since the parse function simply casts the Input type back to string
	cmd := ex.CommandContext(e.ctx, name, parse(args)...) //nolint:gosec // (arg is a function call) see description above
	if len(e.workingDir) > 0 {
		cmd.Dir = e.workingDir
	}
	cmd.Env = parse(e.envVars)
	return command{cmd}
}

type command struct {
	cmd *ex.Cmd
}

func (c command) RunDynamic() error {
	c.cmd.Stderr = os.Stderr
	c.cmd.Stdin = os.Stdin
	c.cmd.Stdout = os.Stdout
	return c.cmd.Run()
}

func (c command) Run() error {
	return c.cmd.Run()
}

func (c command) Output() ([]byte, error) {
	return c.cmd.Output()
}

func (c command) CombinedOutput() ([]byte, error) {
	return c.cmd.CombinedOutput()
}

func debug(inputs []Input) []string {
	res := make([]string, 0, len(inputs))

	for _, e := range inputs {
		res = append(res, e.Debug())
	}
	return res
}
func parse(inputs []Input) []string {
	res := make([]string, 0, len(inputs))
	for _, e := range inputs {
		res = append(res, e.Parse())
	}
	return res
}

type input struct {
	value       string
	private     bool
	debugOutput string
}

func (i input) Parse() string {
	return i.value
}
func (i input) Debug() string {
	if i.debugOutput != "" {
		return i.debugOutput
	}
	if i.private {
		return "xxx"
	}
	return i.value
}

func (i input) IsPrivate() bool {
	if i.debugOutput != "" || i.private {
		return true
	}
	return false
}
