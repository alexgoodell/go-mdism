package main

import (
	"time"

	"github.com/cheggaaa/pb"
)

var beginTime = time.Now() //TODO: Test this [Issue: https://github.com/alexgoodell/go-mdism/issues/32]

// these are all global variables, which is why they are Capitalized
// current refers to the current cycle, which is used to calculate the next cycle

var Query Query_t

//var GlobalStatePopulations = []StatePopulation{}

var output_dir = "tmp"

// TODO: Capitalize global variables [Issue: https://github.com/alexgoodell/go-mdism/issues/46]
var numberOfPeople int
var numberOfPeopleStarting int
var numberOfIterations int
var numberOfPeopleEnteringPerYear int
var numberOfPeopleEntering int

var inputsPath string
var isProfile string
var runType string

var Inputs Input
var Outputs Output

var bar *pb.ProgressBar
