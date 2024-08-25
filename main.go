package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os/exec"
	"slices"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func findRemoteIDA(ida_debugger_name string) (*process.Process, error) {

	processes, err := process.Processes()
	if err != nil {
		return &process.Process{}, fmt.Errorf("Error getting processes: %v", err)
	}

	for _, process := range processes {
		name, _ := process.Name()
		if name == ida_debugger_name {
			return process, nil
		}
	}
	return &process.Process{}, fmt.Errorf("IDA remote debugger process not found")
}

func getChildern(proc *process.Process) ([]process.Process, error) {
	procs := make([]process.Process, 1)
	childern, err := proc.Children()
	if err != nil {
		return nil, fmt.Errorf("No childern found: %v", err)
	}
	for _, child := range childern {
		name, err := child.Name()
		if err != nil {
			fmt.Printf("Error getting process name: %v\n\n", err)
			break
		}
		if name != "conhost.exe" {
			procs = append(procs, *child)
		}
	}
	return procs, nil
}

func inject(pid int32, injector string, hookdll string) error {
	//command := fmt.Sprintf("\"%v\" pid:%v \"%v\" nowait", injector, pid, hookdll)
	stderr, err := execute(injector, hookdll, fmt.Sprintf("pid:%v", pid))

	if stderr != "" {
		return fmt.Errorf("%v", stderr)
	}

	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func execute(injector string, hookdll string, pid string) (string, error) {
	stderr := new(bytes.Buffer)

	cmd := exec.Command(injector, pid, hookdll, "nowait")
	cmd.Stderr = stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(stdout)

	err = cmd.Start()
	if err != nil {
		return "", err
	}
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if scanner.Err() != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return "", scanner.Err()
	}
	return stderr.String(), cmd.Wait()
}

func main() {

	var injector string
	flag.StringVar(&injector, "injector", "", "The Scylla CLI Injector")

	var hookdll string
	flag.StringVar(&hookdll, "hookdll", "", "The Scylla HookDll")

	var ida_remote_debugger string
	flag.StringVar(&ida_remote_debugger, "debugger", "win64_remote64.exe", "The process name of IDA's remote debugger")

	var sleep_val int
	flag.IntVar(&sleep_val, "sleep", 2, "The number of seconds to sleep between looking for new ida debugger child processes to inject into.")

	var delay_diff int
	flag.IntVar(&delay_diff, "delay", 5, "The number of seconds to wait before injecting the scylla dll into the process once found. This is to make sure the executable gets a chance to load properly before the injection occurs.")

	var monitor_for_ida bool
	flag.BoolVar(&monitor_for_ida, "monitor", false, "Keep running, monitoring for IDA's debugger versus exiting when not found.")

	flag.Parse()

	if injector == "" {
		fmt.Println("injector parameter not passed")
		return
	}

	if hookdll == "" {
		fmt.Println("hookdll parameter not passed")
		return
	}

	fmt.Printf("\nLooking for %v\n", ida_remote_debugger)

	pids := make([]int32, 1)
	ida_pid := int32(0)

	for {
		proc, err := findRemoteIDA(ida_remote_debugger)
		if err != nil {
			if fmt.Sprint(err) == "IDA remote debugger process not found" && !monitor_for_ida {
				fmt.Printf("%v\nExiting...", err)
				break
			} else {
				fmt.Printf("%v\nRetrying in %v seconds...\n\n", err, sleep_val)
				time.Sleep(time.Duration(sleep_val) * time.Second)
				continue
			}
		}

		if proc.Pid != ida_pid {
			fmt.Printf("Found %v (%v)\nSilently monitoring for child processes.\n\n", ida_remote_debugger, proc.Pid)
			ida_pid = proc.Pid
		}

		childern, err := getChildern(proc)
		if err != nil {
			fmt.Printf("Error getting childern: %v\n\n", err)
		}

		for _, child := range childern {
			if !slices.Contains(pids, child.Pid) {

				time.Sleep(time.Duration(delay_diff) * time.Second)

				err := inject(child.Pid, injector, hookdll)
				if err != nil {
					fmt.Println(err)
				} else {
					name, err := child.Name()
					if err != nil {
						name = "unknown"
					}
					fmt.Printf("Injected into process %v (%v)\n\n", name, child.Pid)
					pids = append(pids, child.Pid)
				}
			}
		}
		time.Sleep(time.Duration(sleep_val) * time.Second)
	}
}
