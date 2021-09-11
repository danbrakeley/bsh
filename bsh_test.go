package bsh

import (
	"bytes"
	"os"
	"testing"

	"github.com/magefile/mage/mg"
)

const altEnvVarName = "BSH_VERBOSE_TEST"

func Test_EchoFilter(t *testing.T) {
	var b bytes.Buffer
	sh := Bsh{DisableColor: true, Stdout: &b}
	sh.PushEchoFilter("llama")
	sh.Echo("alpha llama gopher")
	actual := b.String()
	expected := "alpha ****** gopher\n"
	if actual != expected {
		t.Errorf(`expected: "%s", but got "%s"`, expected, actual)
	}

	b.Reset()
	sh.PopEchoFilter()
	sh.Echo("alpha llama gopher")
	actual = b.String()
	expected = "alpha llama gopher\n"
	if actual != expected {
		t.Errorf(`expected: "%s", but got "%s"`, expected, actual)
	}

	b.Reset()
	sh.PushEchoFilter("alpha")
	sh.PushEchoFilter("bravo")
	sh.PushEchoFilter("llama")
	sh.Echo("alpha llama gopher")
	actual = b.String()
	expected = "****** ****** gopher\n"
	if actual != expected {
		t.Errorf(`expected: "%s", but got "%s"`, expected, actual)
	}

	b.Reset()
	sh.PopEchoFilter()
	sh.Echo("alpha llama gopher")
	actual = b.String()
	expected = "****** llama gopher\n"
	if actual != expected {
		t.Errorf(`expected: "%s", but got "%s"`, expected, actual)
	}
}

func Test_IsVerbose(t *testing.T) {
	sh := Bsh{}

	os.Unsetenv(altEnvVarName)
	os.Unsetenv(mageVerboseEnvVar)

	if mg.Verbose() {
		t.Errorf(`expected mg.Verbose() to be false when "%s" is unset`, mageVerboseEnvVar)
	}
	if sh.IsVerbose() {
		t.Errorf(`expected IsVerbose() to be false when "%s" is unset`, mageVerboseEnvVar)
	}

	os.Setenv(mageVerboseEnvVar, "true")

	if !mg.Verbose() {
		t.Errorf(`unable to make mg.Verbose() return true by setting "%s" to "true"`, mageVerboseEnvVar)
	}
	if !sh.IsVerbose() {
		t.Errorf(`unable to make IsVerbose() return true by setting "%s" to "true"`, mageVerboseEnvVar)
	}

	os.Unsetenv(mageVerboseEnvVar)

	if mg.Verbose() {
		t.Errorf(`unable to clear mg.Verbose() when "%s" is unset`, mageVerboseEnvVar)
	}
	if sh.IsVerbose() {
		t.Errorf(`unable to clear IsVerbose() when "%s" is unset`, mageVerboseEnvVar)
	}

	sh.SetVerbose(true)

	if !mg.Verbose() {
		t.Errorf(`unable to make mg.Verbose() return true by calling SetVerbose(true)`)
	}
	if !sh.IsVerbose() {
		t.Errorf(`unable to make IsVerbose() return true by calling SetVerbose(true)`)
	}

	sh.SetVerbose(false)

	if mg.Verbose() {
		t.Errorf(`unable to make mg.Verbose() return false by calling SetVerbose(false)`)
	}
	if sh.IsVerbose() {
		t.Errorf(`unable to make IsVerbose() return false by calling SetVerbose(false)`)
	}

	sh.SetVerbose(true)
	sh.SetVerboseEnvVarName(altEnvVarName)

	if !mg.Verbose() {
		t.Errorf(`expected mg.Verbose() to still be true after calling SetVerboseEnvVarName("%s")`, altEnvVarName)
	}
	if !sh.IsVerbose() {
		t.Errorf(`expected IsVerbose() to still return true after calling SetVerboseEnvVarName`)
	}

	os.Unsetenv(mageVerboseEnvVar)

	if mg.Verbose() {
		t.Errorf(`expected mg.Verbose() to be false after calling SetVerboseEnvVarName("%s") and then unsetting "%s"`, altEnvVarName, mageVerboseEnvVar)
	}
	if !sh.IsVerbose() {
		t.Errorf(`expected IsVerbose() after SetVerboseEnvVarName("%s") to still return true after unsetting "%s"`, altEnvVarName, mageVerboseEnvVar)
	}

	sh.SetVerbose(false)

	if mg.Verbose() {
		t.Errorf(`expected mg.Verbose() to still be false`)
	}
	if sh.IsVerbose() {
		t.Errorf(`expected IsVerbose() to be false after SetVerbose(false)`)
	}
}
