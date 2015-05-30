package main

import (
	"fmt"
	"github.com/davecheney/profile"
	"os"
	"strconv"
	"testing"
)

var results_path = "tmp"
var person_count = 1000

func init() {
	var err error

	count := os.Getenv("PERSON_COUNT")
	if count != "" {
		person_count, err = strconv.Atoi(count)
	}

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("person_count=", person_count)
	fmt.Println("Profiling ...")
}

func BenchmarkMemoryProfile(b *testing.B) {
	createPeople(person_count)
	cfg := profile.Config{
		ProfilePath: results_path, // store profiles in current directory
		CPUProfile:  true,
	}

	defer profile.Start(&cfg).Stop()
	runModel()
}

func BenchmarkCPUProfile(b *testing.B) {
	createPeople(person_count)
	cfg := profile.Config{
		ProfilePath: results_path, // store profiles in current directory
		MemProfile:  true,
	}

	defer profile.Start(&cfg).Stop()
	runModel()
}
