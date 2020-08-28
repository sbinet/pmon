# pmon-plot

`pmon-plot` reads monitoring data acquired with `pmon` and dumps plots.

## Example

```
$> pmon -o test.data ./run-cpu -timeout=1m
[...]

$> pmon-plot -o out.png ./test.data
pmon-plot: cmd:   ./run-cpu -timeout=1m
pmon-plot: freq:  1s
pmon-plot: start: 2020-08-28T17:41:28.940016793+02:00
pmon-plot: delta: 1m0.030008631s
pmon-plot: start: 2020-08-28T17:42:28.970025209+02:00
```

![plots](https://github.com/sbinet/pmon/raw/master/cmd/pmon-plot/testdata/out.png)

