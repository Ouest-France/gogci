package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

func Run(name string, args []string) (stdout, stderr []byte, code int, err error) {

	// Create command
	cmd := exec.Command(name, args...)

	// Pipe stdout/stderr to display and capture command output
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdoutMW := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderrMW := io.MultiWriter(os.Stderr, &stderrBuf)

	// Launch command
	err = cmd.Start()
	if err != nil {
		return stdout, stderr, cmd.ProcessState.ExitCode(), fmt.Errorf("command start failed: %s", err)
	}

	// Create goroutines for stdout/stderr capture
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdoutMW, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderrMW, stderrIn)
	wg.Wait()

	// Wait end of command execution
	err = cmd.Wait()
	if err != nil {
		return stdout, stderr, cmd.ProcessState.ExitCode(), fmt.Errorf("command execution failed: %s", err)
	}

	// Return stdout/stderr
	if errStdout != nil {
		return stdout, stderr, cmd.ProcessState.ExitCode(), fmt.Errorf("failed to capture stdout : %s", errStdout)
	}
	if errStderr != nil {
		return stdout, stderr, cmd.ProcessState.ExitCode(), fmt.Errorf("failed to capture stderr: %s", errStderr)
	}

	return stdoutBuf.Bytes(), stderrBuf.Bytes(), cmd.ProcessState.ExitCode(), nil
}
