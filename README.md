pmon
====

`pmon` is a process wrapper for (simple) resource monitoring, on Linux and Darwin.

## Installation

```sh
$ go get github.com/sbinet/pmon
```

## Examples

```sh
$ pmon my-command arg1 arg2

$ pmon -freq=1s -o=pmon.data my-command arg1 arg2

$ cat pmon.data
# pmon: /home/binet/dev/go/gocode/bin/run-cpu run-cpu
# freq: 1s
# format: pmon.Infos{CPU:0, UTime:0, STime:0, VMem:0, RSS:0, Threads:0, Rchar:0, Wchar:0, Rdisk:0, Wdisk:0}
# start: 2015-01-29 12:09:44.695571123 +0100 CET
0.000000 0.000000 0.000000 4416 848 2 0 0 0 0
1000.000000 1000.000000 0.000000 4416 848 3 0 0 0 0
2000.000000 2000.000000 0.000000 4416 848 3 0 0 0 0
[...]
16970.000000 16970.000000 0.000000 4416 848 3 0 0 0 0
17950.000000 17950.000000 0.000000 4416 848 3 0 0 0 0
18950.000000 18950.000000 0.000000 4416 848 3 0 0 0 0
# elapsed: 20.01032601s
# stop: 2015-01-29 12:10:04.705895667 +0100 CET
```

## Limitations

- `I/O` monitoring data isn't captured on Darwin
