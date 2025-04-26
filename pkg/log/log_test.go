package log

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	// capture standard output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// reset verbose and level settings
	verbose = false
	level = INFO

	// test different levels of logs
	Info("This is an info")
	Infof("This is an info with %s", "format")
	Debug("This should not be printed")
	Debugf("This should not be printed with %s", "format")
	Error("This is an error")
	Errorf("This is an error with %s", "format")

	// switch to verbose mode
	SetVerbose(true)
	if !IsVerbose() {
		t.Error("IsVerbose should return true after SetVerbose(true)")
	}
	if GetLevel() != DEBUG {
		t.Errorf("Level should be DEBUG when verbose is true, got %v", GetLevel())
	}

	// test Debug log should be visible in verbose mode
	Debug("This should be printed in verbose mode")
	Debugf("This should be printed in verbose mode with %s", "format")

	// restore standard output and get output content
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// check log output
	t.Run("Info logs are printed", func(t *testing.T) {
		if !strings.Contains(output, "This is an info") {
			t.Error("Info log should be printed")
		}
		if !strings.Contains(output, "This is an info with format") {
			t.Error("Formatted info log should be printed")
		}
	})

	t.Run("Debug logs are not printed without verbose", func(t *testing.T) {
		if strings.Contains(output, "This should not be printed") {
			t.Error("Debug log should not be printed when verbose is false")
		}
		if strings.Contains(output, "This should not be printed with format") {
			t.Error("Formatted debug log should not be printed when verbose is false")
		}
	})

	t.Run("Error logs are printed", func(t *testing.T) {
		if !strings.Contains(output, "This is an error") {
			t.Error("Error log should be printed")
		}
		if !strings.Contains(output, "This is an error with format") {
			t.Error("Formatted error log should be printed")
		}
	})

	t.Run("Debug logs are printed with verbose", func(t *testing.T) {
		if !strings.Contains(output, "This should be printed in verbose mode") {
			t.Error("Debug log should be printed when verbose is true")
		}
		if !strings.Contains(output, "This should be printed in verbose mode with format") {
			t.Error("Formatted debug log should be printed when verbose is true")
		}
	})
}

func TestLogLevel(t *testing.T) {
	// test log level settings
	SetLevel(ERROR)
	if GetLevel() != ERROR {
		t.Errorf("Level should be ERROR, got %v", GetLevel())
	}

	// test verbose overrides log level settings
	SetVerbose(true)
	if GetLevel() != DEBUG {
		t.Errorf("Level should be DEBUG when verbose is true, got %v", GetLevel())
	}

	// test manual level settings can override verbose settings
	SetLevel(ERROR)
	if GetLevel() != ERROR {
		t.Errorf("Level should be ERROR, got %v", GetLevel())
	}
}

// TestFatalStackTrace tests the stack trace functionality of Fatal and Fatalf functions
func TestFatalStackTrace(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Capture stdout
	oldStdout := os.Stdout
	ro, wo, _ := os.Pipe()
	os.Stdout = wo

	// Mock os.Exit
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		if code != 1 {
			t.Errorf("Expected exit code 1, got %d", code)
		}
	}

	// 启用栈跟踪
	oldStackTraceEnabled := IsStackTraceEnabled()
	defer func() { EnableStackTrace(oldStackTraceEnabled) }()
	EnableStackTrace(true)

	// Call Fatal with a test message
	Fatal("Test fatal error")

	// Restore stdout and stderr
	w.Close()
	wo.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout

	// Read stderr output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	stderrOutput := buf.String()

	// Read stdout output
	var bufOut bytes.Buffer
	io.Copy(&bufOut, ro)
	stdoutOutput := bufOut.String()

	// Verify exit was called
	if !exitCalled {
		t.Error("os.Exit was not called")
	}

	// Verify stderr contains stack trace
	if !strings.Contains(stderrOutput, "Stack trace:") {
		t.Error("Stack trace not found in stderr output")
	}

	if !strings.Contains(stderrOutput, "goroutine") {
		t.Error("Stack trace does not contain goroutine information")
	}

	// Verify stdout contains log message
	if !strings.Contains(stdoutOutput, "Test fatal error") {
		t.Error("Fatal log message not found in stdout output")
	}
}

// TestColorOutput tests color output functionality
func TestColorOutput(t *testing.T) {
	// 保存原始设置以便测试结束后恢复
	originalVerbose := IsVerbose()
	originalLevel := GetLevel()
	defer func() {
		SetVerbose(originalVerbose)
		SetLevel(originalLevel)
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 确保所有级别的日志都可见
	SetLevel(DEBUG)

	// Test with colors enabled
	EnableColor(true)
	if !IsColorEnabled() {
		t.Error("Color should be enabled")
	}

	// Output logs with colors
	Info("Colored info")
	Error("Colored error")
	Debug("Colored debug")

	// Disable colors
	EnableColor(false)
	if IsColorEnabled() {
		t.Error("Color should be disabled")
	}

	// Output logs without colors
	Info("Non-colored info")
	Error("Non-colored error")
	Debug("Non-colored debug")

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	t.Logf("Test output: %s", output)

	// With colors enabled, ANSI sequences should be present
	if !strings.Contains(output, ColorGreen) {
		t.Error("Green color code not found in colored output")
	}

	if !strings.Contains(output, ColorRed) {
		t.Error("Red color code not found in colored output")
	}

	if !strings.Contains(output, ColorCyan) {
		t.Error("Cyan color code not found in colored output")
	}

	if !strings.Contains(output, ColorReset) {
		t.Error("Color reset code not found in colored output")
	}

	// After disabling colors, new logs should not have color codes
	// Find the position of "Non-colored" to check only the part after disabling colors
	nonColoredPos := strings.Index(output, "Non-colored")
	if nonColoredPos == -1 {
		t.Error("Non-colored log message not found")
		return
	}

	outputAfterDisable := output[nonColoredPos:]
	if strings.Contains(outputAfterDisable, ColorGreen) ||
		strings.Contains(outputAfterDisable, ColorRed) ||
		strings.Contains(outputAfterDisable, ColorReset) {
		t.Error("Color codes found after disabling colors")
	}
}
