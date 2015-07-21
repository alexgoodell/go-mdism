package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
)

func initializeInputs(inputsPath string) {
	// TODO: ALEX: Why do some have LEptr, and others only ptr? This does not cause trouble? [Issue: https://github.com/alexgoodell/go-mdism/issues/61]

	// ####################### Psa Inputs #######################

	filename := "inputs/" + inputsPath + "/psa.csv"
	numberOfRecords := getNumberOfRecords(filename)

	Inputs.PsaInputs = make([]PsaInput, numberOfRecords, numberOfRecords)
	var PsaPtrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		PsaPtrs = append(PsaPtrs, new(PsaInput))
	}
	PsaPtrs = fromCsv(filename, Inputs.PsaInputs[0], PsaPtrs)
	for i, ptr := range PsaPtrs {
		Inputs.PsaInputs[i] = *ptr.(*PsaInput)
	}

	// ####################### Interventions #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/interventions.csv"
	numberOfRecords = getNumberOfRecords(filename)
	Inputs.Interventions = make([]Intervention, numberOfRecords, numberOfRecords)
	var Iptrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		Iptrs = append(Iptrs, new(Intervention))
	}
	Iptrs = fromCsv(filename, Inputs.Interventions[0], Iptrs)
	for i, ptr := range Iptrs {
		Inputs.Interventions[i] = *ptr.(*Intervention)
	}
	//fmt.Println("complete")

	// ####################### Life Expectancy #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/life-expectancies.csv"
	numberOfRecords = getNumberOfRecords(filename)
	Inputs.LifeExpectancies = make([]LifeExpectancy, numberOfRecords, numberOfRecords)
	var LEptrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		LEptrs = append(LEptrs, new(LifeExpectancy))
	}
	LEptrs = fromCsv(filename, Inputs.LifeExpectancies[0], LEptrs)
	for i, ptr := range LEptrs {
		Inputs.LifeExpectancies[i] = *ptr.(*LifeExpectancy)
	}
	//fmt.Println("complete")

	// ####################### Models #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/models.csv"
	numberOfRecords = getNumberOfRecords(filename)
	Inputs.Models = make([]Model, numberOfRecords, numberOfRecords)
	var ptrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		ptrs = append(ptrs, new(Model))
	}
	ptrs = fromCsv(filename, Inputs.Models[0], ptrs)
	for i, ptr := range ptrs {
		Inputs.Models[i] = *ptr.(*Model)
	}
	//fmt.Println("complete")

	// ####################### States #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/states.csv"
	//fmt.Println(filename)
	numberOfRecords = getNumberOfRecords(filename)
	Inputs.States = make([]State, numberOfRecords, numberOfRecords)
	var statePtrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		statePtrs = append(statePtrs, new(State))
	}
	ptrs = fromCsv(filename, Inputs.States[0], statePtrs)
	for i, ptr := range statePtrs {
		Inputs.States[i] = *ptr.(*State)
	}

	// ####################### Transition Probabilities #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/transition-probabilities.csv"
	numberOfRecords = getNumberOfRecords(filename)

	Inputs.TransitionProbabilities = make([]TransitionProbability, numberOfRecords, numberOfRecords)
	var tpPtrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		tpPtrs = append(tpPtrs, new(TransitionProbability))
	}
	ptrs = fromCsv(filename, Inputs.TransitionProbabilities[0], tpPtrs)
	for i, ptr := range tpPtrs {
		Inputs.TransitionProbabilities[i] = *ptr.(*TransitionProbability)
	}

	// ####################### Interactions #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/interactions.csv"
	numberOfRecords = getNumberOfRecords(filename)

	Inputs.Interactions = make([]Interaction, numberOfRecords, numberOfRecords)
	var interactionPtrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		interactionPtrs = append(interactionPtrs, new(Interaction))
	}
	ptrs = fromCsv(filename, Inputs.Interactions[0], interactionPtrs)
	for i, ptr := range interactionPtrs {
		Inputs.Interactions[i] = *ptr.(*Interaction)
	}

	// ####################### Cycles #######################

	//initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/cycles.csv"
	numberOfRecords = getNumberOfRecords(filename)

	Inputs.Cycles = make([]Cycle, numberOfRecords, numberOfRecords)
	var cyclePtrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		cyclePtrs = append(cyclePtrs, new(Cycle))
	}
	ptrs = fromCsv(filename, Inputs.Cycles[0], cyclePtrs)
	for i, ptr := range cyclePtrs {
		Inputs.Cycles[i] = *ptr.(*Cycle)
	}

	// ####################### TPs By RAS #######################

	filename = "inputs/" + inputsPath + "/ras.csv"
	numberOfRecords = getNumberOfRecords(filename)

	Inputs.TPByRASs = make([]TPByRAS, numberOfRecords, numberOfRecords)
	var tpbrsPtr []interface{}
	for i := 0; i < numberOfRecords; i++ {
		tpbrsPtr = append(tpbrsPtr, new(TPByRAS))
	}
	ptrs = fromCsv(filename, Inputs.TPByRASs[0], tpbrsPtr)
	for i, ptr := range tpbrsPtr {
		Inputs.TPByRASs[i] = *ptr.(*TPByRAS)
	}

	// ####################### Costs #######################

	filename = "inputs/" + inputsPath + "/costs.csv"
	numberOfRecords = getNumberOfRecords(filename)

	Inputs.Costs = make([]Cost, numberOfRecords, numberOfRecords)
	var costsPtr []interface{}
	for i := 0; i < numberOfRecords; i++ {
		costsPtr = append(costsPtr, new(Cost))
	}
	ptrs = fromCsv(filename, Inputs.Costs[0], costsPtr)
	for i, ptr := range costsPtr {
		Inputs.Costs[i] = *ptr.(*Cost)
	}

	// ####################### Disability Weights #######################

	filename = "inputs/" + inputsPath + "/disability-weights.csv"
	numberOfRecords = getNumberOfRecords(filename)

	Inputs.DisabilityWeights = make([]DisabilityWeight, numberOfRecords, numberOfRecords)
	var dwPtrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		dwPtrs = append(dwPtrs, new(DisabilityWeight))
	}
	ptrs = fromCsv(filename, Inputs.DisabilityWeights[0], dwPtrs)
	for i, ptr := range dwPtrs {
		Inputs.DisabilityWeights[i] = *ptr.(*DisabilityWeight)
	}

	// ####################### Interventions #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/interventionvalues.csv"
	numberOfRecords = getNumberOfRecords(filename)
	Inputs.InterventionValues = make([]InterventionValue, numberOfRecords, numberOfRecords)
	var Ivptrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		Ivptrs = append(Ivptrs, new(InterventionValue))
	}
	Ivptrs = fromCsv(filename, Inputs.InterventionValues[0], Ivptrs)
	for i, ptr := range Ivptrs {
		Inputs.InterventionValues[i] = *ptr.(*InterventionValue)
	}

	// ####################### Regression Rates #######################

	filename = "inputs/" + inputsPath + "/regression-rates.csv"
	numberOfRecords = getNumberOfRecords(filename)
	Inputs.RegressionRates = make([]RegressionRate, numberOfRecords, numberOfRecords)
	var RRptrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		RRptrs = append(RRptrs, new(RegressionRate))
	}
	RRptrs = fromCsv(filename, Inputs.RegressionRates[0], RRptrs)
	for i, ptr := range RRptrs {
		Inputs.RegressionRates[i] = *ptr.(*RegressionRate)
	}

}

// Exports sets of data to CSVs. I particular, it will print any array of structs
// and automatically uses the struct field names as headers! wow.
// It takes a filename, as well one copy of the struct, and the array of structs
// itself.
func toCsv(filename string, record interface{}, records interface{}) error {
	fmt.Println("Beginning export to ", filename)
	//create or open file
	os.Create(filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// new Csv wriier
	writer := csv.NewWriter(file)
	// use the single record to determine the fields of the struct
	val := reflect.Indirect(reflect.ValueOf(record))
	numberOfFields := val.Type().NumField()
	var fieldNames []string
	for i := 0; i < numberOfFields; i++ {
		fieldNames = append(fieldNames, val.Type().Field(i).Name)
	}
	// print field names of struct
	err = writer.Write(fieldNames)
	// print the values from the array of structs
	val2 := reflect.ValueOf(records)
	for i := 0; i < val2.Len(); i++ {
		var line []string
		for p := 0; p < numberOfFields; p++ {
			//convert interface to string
			line = append(line, fmt.Sprintf("%v", val2.Index(i).Field(p).Interface()))
		}
		err = writer.Write(line)
	}
	if err != nil {
		fmt.Println("error")
		os.Exit(1)
	}
	//fmt.Println("Exported to ", filename)
	writer.Flush()
	return err
}
