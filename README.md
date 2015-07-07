# go-mdism

## Profiling

Profiles can be generated in the google [perftools](https://code.google.com/p/gperftools/) format (.pprof).
Profiles are generated using the go benchmarking facility:

```
go test -bench cpu
go test -bench mem
```

See `benchmark_test.go` to view the different profiles. By default profile results are output to `tmp`.

To visualize the profiles, you will need to install perftools (`brew install google-perftools` on osx).

```
go tool pprof --pdf (which go-mdism) tmp/cpu.pprof  > test.pdf
```

