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

## Todo


0.4 <- varied, then sum was 1.1
0.3 <- use correction factor (minus - 3/7 * 0.1)
0.4 <- use correction factor (minus - 4/7 * 0.1)


General to do
- [ ] Split DALYs, YLD, YLL by state - Rick to complete
- [ ] Set up PSA reporting - natural events, PSA switch etc - Alex to complete
- [] sensitivity analysis tools - built gamma, beta, and normal. To discuss when to use each
- [x] costs - rick is working on
- [x] discounting - began (6/22/15), waiting on response from Jim
- [x] open cohort - done (6/20/15)
- [x] add CHD into risk factor issue - done  (6/23/15)
- [x] add death-age reporting - done  (6/22/15)
- [x] death sync - done  (6/22/15)
- [x] flask display for charts - done  (6/21/15)
- [x] stack chart - done  (6/22/15)
- [x] intervention - done (6/23/15)
- [x] report costs (6/24/15)
- [ ] fix TPs and "other deaths" - later, works as is now, just not elegant
- [ ] int to uint - later, works now as is, this is an optimization
- [ ] Make profiling simpler. Maybe have a makefile that allows for a simple `make profile-cpu` that automatically generates and opens the results in your browser
- [ ] Calculate prevalence per cycle, calculate costs and YLD per cycle as well. This could save time because costs are only added once per cycle, in stead of 68000 times