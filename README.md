## Why

Wanted a tool that would auto inject Scylla Hide into whatever process being debugged using IDA's remote debugger without having to manually re-inject each time the debuggeed instance gets restarted.

## How Does It Work

1. Finds the running instance of IDA's remote debugger by process name
2. Looks for child processes that are not conhost.exe
3. Waits 5 seconds after identifiying a process
4. Injects Scylla Hide into the processing using ScyllaHide's injector executable
5. Sleeps for 2 seconds before repeating the process

## Build

Tested from a Linux machine with golang version `1.21.3`.

```
GOOS=windows GOARCH=amd64 go build -o bin/remote_ida_scylla_inject.exe main.go
```

Might need to run `go get github.com/xorhex/remote_ida_scylla_inject` once it if complains about missing `go.sum`.

Requirements

The following components built from ![ScyllaHide](https://github.com/x64dbg/ScyllaHide) or found in the ![release build](https://github.com/x64dbg/ScyllaHide/releases/tag/v1.4) on Github:

- InjectorCLIx86.exe
- HookLibraryx86.dll
- InjectorCLIx64.exe
- HookLibraryx64.dll

## Options

```
Usage of remote_ida_scylla_inject:
  -debugger string
    	The process name of IDA's remote debugger (default "win64_remote64.exe")
  -delay int
    	The number of seconds to wait before injecting the scylla dll into the process once found. This is to make sure the executable gets a chance to load properly before the injection occurs. (default 5)
  -hookdll string
    	The Scylla HookDll
  -injector string
    	The Scylla CLI Injector
  -monitor
    	Keep running, monitoring for IDA's debugger versus exiting when not found.
  -sleep int
    	The number of seconds to sleep between looking for new ida debugger child processes to inject into. (default 2)
```

## Running

```
remote_ida_scylla_inject.exe -injector <PATH TO>\InjectorCLIx86.exe -hookdll <PATH TO>\HookLibraryx86.dll
```

The injector and hooklibrary need to match the bittness of the binary being executed by IDA (not the bittness of IDA's remote debugger).

To start `remote_ida_scylla_inject` before starting the remote IDA debugger, pass the `-monitor` flag as well.
