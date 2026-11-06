package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	// go run . run /bin/echo hello
	// go run . run /bin/sh
	if len(os.Args) < 3 || os.Args[1] != "run" {
		fmt.Println("Usage: go run . run <command> [args...]")
		os.Exit(1)
	}

	fmt.Printf("Parent PID: %d\n", os.Getpid())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// create new * namesapce for child process
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Error starting command:", err)
		os.Exit(1)
	}

	fmt.Printf("Child PID: %d\n", cmd.Process.Pid)
	
	if err := cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for command:", err)
		os.Exit(1)
	}

	fmt.Println("Child exits, parent process ends.")
	
}