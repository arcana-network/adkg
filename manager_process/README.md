# Notes

## How to start main/child processes
```
go build -o myapp main.go
```

```
go build -o starter ./manager_process
```
```
./starter
```

Should show n+1 results, if n processes were started. (Because grep itself is counted as well)
```
ps aux | grep myapp
```

## Current status

Main can start 2 child processes with adkg codebase in dev-multiprocess branch.
Main can send messages to both child processes via stdin and the process echo the message back.

## Start 1 process manually (directly run binary instead of through main)

To test how a correct startup could be and see what logging looks like. 

In `arcana_impl_adkg`:

```
./myapp start --config local-setup-data/config.local.1.json --secret-config local-setup-data/config.local.1.json --ip-address 127.0.0.1
```