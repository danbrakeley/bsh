package bsh

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	mageVerboseEnvVar = "MAGEFILE_VERBOSE"
)

type Bsh struct {
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
	DisableColor bool

	fnErr       func(error)
	echoFilters []string
	// defaults to Mage's verbose flag, since this package was original written to be used in Magefiles.
	// However, if you want to use your own VERBOSE flag here, just call SetVerboseEnvVarName.
	verboseEnvVar string
}

// ensureStdin returns Stdin or os.Stdin (never nil, unless os.Stdin is nil)
func (b *Bsh) ensureStdin() io.Reader {
	if b.Stdin == nil {
		return os.Stdin
	}
	return b.Stdin
}

// ensureStdout returns Stdout or os.Stdout (never nil, unless os.Stdout is nil)
func (b *Bsh) ensureStdout() io.Writer {
	if b.Stdout == nil {
		return os.Stdout
	}
	return b.Stdout
}

// ensureStderr returns Stderr or os.Stderr (never nil, unless os.Stderr is nil)
func (b *Bsh) ensureStderr() io.Writer {
	if b.Stderr == nil {
		return os.Stderr
	}
	return b.Stderr
}

// SetErrorHandler sets the behavior when an error is encountered while running most commands.
// The default behavior is to panic.
func (b *Bsh) SetErrorHandler(fnErr func(error)) {
	b.Verbose("Error handler changed")
	b.fnErr = fnErr
}

// Panic is called internally any time there's an unhandled error. It will in turn call any
// error handler set by SetErrorHandler, or panic() if no error handler was set.
func (b *Bsh) Panic(err error) {
	if b.fnErr != nil {
		b.fnErr(err)
	} else {
		panic(err)
	}
}

// filter secrets from the output

func (b *Bsh) PushEchoFilter(str string) {
	b.echoFilters = append(b.echoFilters, str)
}

func (b *Bsh) PopEchoFilter() {
	b.echoFilters = b.echoFilters[:len(b.echoFilters)-1]
}

// Echo writes to stdout, and ensures the last character written is a newline.

func (b *Bsh) Echo(str string) {
	b.echo(str, ensureNewline, colorEcho)
}

func (b *Bsh) Echof(format string, args ...interface{}) {
	b.echo(fmt.Sprintf(format, args...), ensureNewline, colorEcho)
}

// SetVerboseEnvVarName allows changing the name of the environment variable that is used to
// decide if we are in Verbose mode. This function creates the new env var immediately,
// setting its value to true or false based on the value of the old env var name.
func (b *Bsh) SetVerboseEnvVarName(s string) {
	wasVerbose := b.IsVerbose()
	b.verboseEnvVar = s
	b.SetVerbose(wasVerbose)
}

func (b *Bsh) SetVerbose(v bool) {
	if len(b.verboseEnvVar) == 0 {
		b.verboseEnvVar = mageVerboseEnvVar
	}
	os.Setenv(b.verboseEnvVar, strconv.FormatBool(v))
}

func (b *Bsh) IsVerbose() bool {
	if len(b.verboseEnvVar) == 0 {
		b.verboseEnvVar = mageVerboseEnvVar
	}
	v, _ := strconv.ParseBool(os.Getenv(b.verboseEnvVar))
	return v
}

func (b *Bsh) Verbose(str string) {
	if !b.IsVerbose() {
		return
	}
	b.echo(str, ensureNewline, colorVerbose)
}

func (b *Bsh) Verbosef(format string, args ...interface{}) {
	if !b.IsVerbose() {
		return
	}
	b.echo(fmt.Sprintf(format, args...), ensureNewline, colorVerbose)
}

func (b *Bsh) Warn(str string) {
	b.echo(str, ensureNewline, colorWarn)
}

func (b *Bsh) Warnf(format string, args ...interface{}) {
	b.echo(fmt.Sprintf(format, args...), ensureNewline, colorWarn)
}

type echoOpt byte

const (
	ensureNewline echoOpt = iota
	ignoreFilter  echoOpt = iota
	colorEcho     echoOpt = iota
	colorVerbose  echoOpt = iota
	colorAsk      echoOpt = iota
	colorWarn     echoOpt = iota
)

func (b *Bsh) echo(str string, opts ...echoOpt) {
	newline := false
	filter := true
	var color string
	for _, v := range opts {
		switch v {
		case ensureNewline:
			newline = true
		case ignoreFilter:
			filter = false
		case colorEcho:
			color = ansiWhite
		case colorVerbose:
			color = ansiCyan
		case colorAsk:
			color = ansiBlue
		case colorWarn:
			color = ansiYellow
		}
	}

	if filter {
		for _, v := range b.echoFilters {
			str = strings.ReplaceAll(str, v, "******")
		}
	}

	if newline && str[len(str)-1] != '\n' {
		str += "\n"
	}

	if !b.DisableColor && len(color) > 0 {
		_, exists := os.LookupEnv("NO_COLOR")
		if !exists {
			str = color + str + ansiReset
		}
	}

	fmt.Fprint(b.ensureStdout(), str)
}

// ScanLine reads from default stdin until a newline is encountered

func (b *Bsh) ScanLine() string {
	str, err := b.ScanLineErr()
	if err != nil {
		b.Panic(err)
	}
	return str
}

func (b *Bsh) ScanLineErr() (string, error) {
	r := bufio.NewReader(b.ensureStdin())
	str, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(str, "\n"), nil
}

// Ask is a combination of echo and scanline

func (b *Bsh) Ask(msg string) string {
	b.echo(msg, colorAsk)
	return b.ScanLine()
}

func (b *Bsh) Askf(format string, args ...interface{}) string {
	b.echo(fmt.Sprintf(format, args...), colorAsk)
	return b.ScanLine()
}

// ansi color helpers

const (
	ansiCSI   = "\u001b[" // Control Sequence Introducer
	ansiReset = ansiCSI + "39m"

	ansiBlack       = ansiCSI + "30m"
	ansiDarkRed     = ansiCSI + "31m"
	ansiDarkGreen   = ansiCSI + "32m"
	ansiDarkYellow  = ansiCSI + "33m"
	ansiDarkBlue    = ansiCSI + "34m"
	ansiDarkMagenta = ansiCSI + "35m"
	ansiDarkCyan    = ansiCSI + "36m"
	ansiLightGray   = ansiCSI + "37m"

	ansiDarkGray = ansiCSI + "90m"
	ansiRed      = ansiCSI + "91m"
	ansiGreen    = ansiCSI + "92m"
	ansiYellow   = ansiCSI + "93m"
	ansiBlue     = ansiCSI + "94m"
	ansiMagenta  = ansiCSI + "95m"
	ansiCyan     = ansiCSI + "96m"
	ansiWhite    = ansiCSI + "97m"
)
