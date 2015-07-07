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
Discuss:
- How many cycles do we really need?
	- And do we need to start calculating DALYs et cetera after 0 or 1? (line 491)
	- And do we want to adjust the regression rate after cycle 1 or 2? (lne 1121)
- When are we letting new people enter and how are we going to make sure they aren't 21 already? (line 281)(add age 19?)
- Results (all seen from go point of view):
	- NAFLD steatosis is a little low
	- NAFLD mnatural mortality is high
	- NAFLD otherDeath is low
	- Too many people in the model
	- Not enough cycles are modeled (one too few)
	- T2D death is a little low
	- T2D otherDeath is a little high (these two are quite small)
	- CHD prevalence is low
	- (Therefore) CHD death is also low
	- CHD other death is a bit high
	- Normal weight and overweight are very much lower
	- Obesity is extremely high
	- BMI otherdeath is quite accurate - this probably indicates that the difference in the ither models stem from the fact that these people die from natural causes before they can die of the disease. The total death rate is quite accurate.
	- Age at uninitialized has 1 extra person? Why is that? Somewhere +1?
- Turn slices into arrays? Fixed length is faster?

# PSA
- [ ] Set up PSA reporting - natural events, PSA switch etc - Alex to complete
- [ ] Define borders and distributions for PSA. How to code this in GO? - Randomly change the input variables? + Also: what happens when a baseline transition probability changes? Do we adjust the rest accordingly or should we subtract/add from/to remaining? Discuss with Jim Tuesday 07-07.

# DALY, YLD and YLL
- [ ] Adjust YLD values - when somebody has 2 diseases simultaneously Dw = 1-PRODUCT(1-Dwx) Where Dwx is each disease specific Dw.
The same goes for YLL values, if somebody by chance dies in two models at the same time, you should not count those deaths twice. Attribute half to both diseases.
- [ ] Review YLD and YLL formula's for discounting - do we need to take age weighing into account? I suggest NO. Ask Jim tuesday. Also, we would need incidence and duration to calculate YLD properly. The prevalence way is just a measure for calculating healthy life expectancy in global disease studies, it cannot be used if you discount and/or use age-weighing.

# Outputs
- [ ] make output file that writes exactly all things we are interested in. If necessary can be combined with excel macro.
	- events need to be coded first
	- make output file for costs, calys, yld and yll and add these to the macro for prevalence - I have made a DALYstoExport file with some functions, I am not sure if it will work because I do not completely understand it (it is similar to the state_populations).
	I already have a macro to accurately access all state populations, turn them into graphs and compare those to excel.
- [ ] Discuss life expectancy table - use HALE or LE and should we (and how) extrapolate above age 80? - Discuss with Jim.
	I have extrapolated according to an article in lancet that suggested LE up until 105. Have already put that in.

# General
- [ ] Why does it say at inputs 0 (pre) until 25, but at outputs we have year 0 until 26. Those include -1 and -2? But then it seems that cycle 25 does not get simulated? For cycle inputs I have set 0 (2010) until 25 (2035), but it seems that he doesn't simulate 25?
	Is cycle 0 the same as unin1 and unin2? Or how does that work exactly?
- [ ] Interactions seem not to be functioning. Fix this.
- [ ] Entry of new people causes them to be 21 before they actualy start being simulated. Fix this.
- [ ] Introduce effects of Sugar->BMI->NAFLD->Diabetes->CHD? - Discuss with Jim
- [ ] Need to build in checks to see if decline mortality rate is actually functioning, same goes for other interactions.
- [ ] Starting prevalence of HCC should only be possible in the NAFLD population.
- [ ] What is the weird function after line 316 (just after time elapsed, excluding data import and export)

# Optimizations for later
- [ ] int to uint - later, works now as is, this is an optimization
- [ ] Make profiling simpler. Maybe have a makefile that allows for a simple `make profile-cpu` that automatically generates and opens the results in your browser
- [ ] Calculate prevalence per cycle, calculate costs and YLD per cycle as well. This could save time because costs are only added once per cycle, in stead of 68000 times
- [ ] Can leave for now but would be nice memory feature: make liver related death rate from HCC dependent on disease duration.
- [ ] fix TPs and "other deaths" - later, works as is now, just not elegant


# Done
- [X] Split DALYs, YLD, YLL by state - Rick to complete - finished + no more discounting/age weighing in DALYs
- [X] sensitivity analysis tools - built gamma, beta, and normal. To discuss when to use each
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
- [x] Extra people entering starts two years to soon. First we have 2 unin cycles, then cycle 0, then 25 normal cycles. Now they start entering at unin2, and therefore year 0 already has the 416 extra people, while they should only start filling at year 2.
- [x] Adjust the formula for DALY's to be split per group
- [x] Code outcomes for age - how many people of a certain age at each cycle during simulation => can be done from state populations with macro
- [x] We should actually split YLL between male and female, because the life expectancy table is different for each.
- [x] Health adjusted life expectancy for YLL calculations: extrapolate above age 80.
- [x] There should not be people switching between fructose risk groups, since this is a percentage of the people already in that group, it will be inequal and therefore make the groups change - I have changed this.
- [x] When new people are added to the model, he distributed them among ages according to the TPs. But they should all enter with age 20! -> changed entry state to 42 instead of 41 -> but now they are all 21 upon entering.
- [x] Check all values of interactions and baseline TPs - especially the age corrected ones. Why is the incidence of CHD so high? Also check if formula of declining rates functions properly. Check values in RAS file. While doing that:
	- [x] Introduce starting prevalence for HCC
	- [x] Introduce chance to die of liver death from NASH and Cirrhosis (+ subtract from total deaths?)