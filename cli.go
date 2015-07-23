package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/codegangsta/cli"
	"github.com/davecheney/profile"
)

func main() {

	app := cli.NewApp()
	app.Name = "Go Mdism - mmmSugar edition"
	app.Usage = "go-mdism help for help"
	app.Version = "dev-0.1"

	// TODO_LATER: Build flag system out
	globalFlags := []cli.Flag{
		cli.IntFlag{
			Name:  "people",
			Value: 22400,
			Usage: "number of people",
		},
	}

	app.Action = func(c *cli.Context) {
		show_greeting()
		println("go-mdism help for help")
		fmt.Println(c.String("lang"))
	}

	app.Commands = []cli.Command{
		{
			Name:    "single",
			Aliases: []string{"a"},
			Usage:   "Run a single simulation",
			Flags:   globalFlags,
			Action: func(c *cli.Context) {
				processFlags(c)
				startRunWithSingle()
			},
		},
		{
			Name:    "psa",
			Aliases: []string{"a"},
			Usage:   "Run a PSA until user quits (control+C)",
			Flags:   globalFlags,
			Action: func(c *cli.Context) {
				processFlags(c)
				startRunWithPsa()
			},
		},
		{
			Name:    "dsa",
			Aliases: []string{"a"},
			Usage:   "Run a DSA",
			Flags:   globalFlags,
			Action: func(c *cli.Context) {
				processFlags(c)
				startRunWithDsa()
			},
		},
	}

	app.Run(os.Args)
}

func processFlags(c *cli.Context) {
	//TODO_LATER
}

func startRunWithSingle() {
	runType = "single"
	initialize()
	initializeInputs(inputsPath)
	Query.setUp()
	runInterventions()
}

func startRunWithPsa() {
	runType = "psa"
	initialize()
	generateAllPsaValues()
	runPsa()

	// create people will generate individuals and add their data to the master
	// records

	fmt.Println("Intialization complete, time elapsed:", fmt.Sprint(time.Since(beginTime)))

	// table tests here

	for true {
		initializeInputs(inputsPath)
		Query.setUp()
		generateAllPsaValues()
		runPsa()
		randomLetters = randSeq(10)
		runInterventions()
	}

}

func startRunWithDsa() {
	runType = "dsa"
	initialize()

	for i := 0; i < 76; i++ {
		for p := 1; p < 6; p++ {
			initializeInputs(inputsPath)
			Query.setUp()
			runNewDsaValue(i, p)
			runDsa()
			randomLetters = randSeq(10)
			runInterventions()
		}
	}

	// create people will generate individuals and add their data to the master
	// records

	fmt.Println("Intialization complete, time elapsed:", fmt.Sprint(time.Since(beginTime)))

	// table tests here
}

func initialize() {

	show_greeting()

	flag.IntVar(&numberOfPeopleStarting, "people", 22400, "number of people to run")
	flag.IntVar(&numberOfIterations, "iterations", 1, "number times to run")
	// TODO: index error if number of people entering is <15000 [Issue: https://github.com/alexgoodell/go-mdism/issues/33]
	flag.IntVar(&numberOfPeopleEnteringPerYear, "entering", 416, "number of people that will enter the run(s)")
	flag.StringVar(&inputsPath, "inputs", "example", "folder that stores input csvs")
	flag.StringVar(&isProfile, "profile", "false", "cpu, mem, or false")
	flag.Parse()

	if isProfile != "false" {
		fmt.Println("Enabling profiler")

		if isProfile == "cpu" {
			cfg := profile.Config{
				ProfilePath: ".", // store profiles in current directory
				CPUProfile:  true,
			}
			defer profile.Start(&cfg).Stop()
		} else if isProfile == "mem" {
			cfg := profile.Config{
				ProfilePath: ".", // store profiles in current directory
				MemProfile:  true,
			}
			defer profile.Start(&cfg).Stop()
		}
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("using ", runtime.NumCPU(), " cores")
	// Seed the random function

	// TODO: remove hardcoded cycles [Issue: https://github.com/alexgoodell/go-mdism/issues/40]
	numberOfPeopleEntering = numberOfPeopleEnteringPerYear * (26 + 1)
	numberOfPeople = numberOfPeopleEntering + numberOfPeopleStarting

	fmt.Println("and ", numberOfPeopleStarting, "initial individuals")
	fmt.Println("and ", numberOfPeopleEntering, "individuals entering")
	fmt.Println("and ", numberOfIterations, "iterations")
	fmt.Println("and ", inputsPath, " as inputs")

}
