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
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run . run <command> [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		run(os.Args[2:])
	case "init":
		fmt.Printf("I am child process, pid = %d\n", os.Getpid())
		fmt.Printf("Prepare the cmd to be executed: %v\n", os.Args[2:])

		args := os.Args[2:]

		// MOUNT
		err := syscall.Mount("", "/", "", uintptr(syscall.MS_PRIVATE|syscall.MS_REC), "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error mounting root filesystem:", err)
			os.Exit(1)
		}
		fmt.Println("Mount propagation is now private")

		// mount a new proc filesystem under current "/proc"
		err = syscall.Mount(
			"proc",  // source
			"/proc", // target
			"proc",  // filesystem type
			uintptr(syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV),  // mount flags
			"" )
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error mounting proc:", err)
			os.Exit(1)
		}
		fmt.Println("New /proc is ready")

		// EXEC
		// find the program 
		path, err := exec.LookPath(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error finding command:", err)
			os.Exit(1)
		}

		// replace the current progress with the program
		if err := syscall.Exec(path, args, os.Environ()); err != nil {
			fmt.Fprintln(os.Stderr, "Error executing command:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

func run(args []string) {
	fmt.Printf("Parent PID: %d\n", os.Getpid())

	childArgs := append([]string{"init"}, args...)
	// /proc/self/exec points to the current my docker executable
	// cmd becomes "mydocker init /bin/echo hello"
	cmd := exec.Command("/proc/self/exe", childArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// create new * namesapce for child process
		Cloneflags: 
			syscall.CLONE_NEWUTS | 
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS,
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
