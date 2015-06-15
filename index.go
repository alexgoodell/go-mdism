// General to do
// * int to uint

package main

import (
	// "encoding/json"
	"flag"
	"fmt"
	// 	"github.com/alexgoodell/ghdmodel/models"
	//"io"
	// 	"net/http"
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"github.com/davecheney/profile"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	// "runtime/pprof"
	"github.com/spf13/nitro"
	"strconv"
	"time"
)

var beginTime = time.Now()

type State struct {
	Id                     int
	Model_id               int
	Name                   string
	Is_uninitialized_state bool
}

type Model struct {
	Id   int
	Name string
}

type MasterRecord struct {
	Cycle_id  int
	Person_id int
	State_id  int
	Model_id  int
}

type Cycle struct {
	Id   int
	Name string
}

type Person struct {
	Id int
}

type Interaction struct {
	Id                int
	In_state_id       int
	From_state_id     int
	To_state_id       int
	Adjustment        float64
	Effected_model_id int
}

type TransitionProbability struct {
	Id      int
	From_id int
	To_id   int
	Tp_base float64
}

type Query struct {
	State_id_by_cycle_and_person_and_model [][][]int
	States_ids_by_cycle_and_person         [][]int
	Tps_id_by_from_state                   [][]int
	Interactions_id_by_in_state_and_model  [][]int
}

type Input struct {
	CurrentCycle            int
	QueryData               Query
	Models                  []Model
	People                  []Person
	States                  []State
	TransitionProbabilities []TransitionProbability
	Interactions            []Interaction
	Cycles                  []Cycle
	MasterRecords           []MasterRecord
}

// these are all global variables, which is why they are Capitalized
// current refers to the current cycle, which is used to calculate the next cycle

var GlobalMasterRecords = []MasterRecord{}

var GlobalMasterRecordsByIPCM [][][][]int

var output_dir = "tmp"

var numberOfPeople int
var numberOfIterations int
var inputsPath string
var isProfile string

var Timer *nitro.B

func main() {

	Timer = nitro.Initalize()

	flag.IntVar(&numberOfPeople, "people", 1000, "number of people to run")
	flag.IntVar(&numberOfIterations, "iterations", 1, "number times to run")
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

	fmt.Println("and ", numberOfPeople, "individuals")
	fmt.Println("and ", numberOfIterations, "iterations")
	fmt.Println("and ", inputsPath, " as inputs")

	var Inputs Input
	Inputs = initializeInputs(Inputs, inputsPath)

	//set up queryData
	Inputs = setUpQueryData(Inputs, numberOfPeople)

	// create people will generate individuals and add their data to the master
	// records
	Inputs = createPeople(Inputs, numberOfPeople)

	setUpGlobalMasterRecordsByIPCM(Inputs)

	// table tests here

	concurrencyBy := "person"

	iterationChan := make(chan string)

	for i := 0; i < numberOfIterations; i++ {
		go runModel(Inputs, concurrencyBy, iterationChan)
	}

	for i := 0; i < numberOfIterations; i++ {
		toPrint := <-iterationChan
		fmt.Println(toPrint)
	}

	Timer.Step("main")

}

func runModel(Inputs Input, concurrencyBy string, iterationChan chan string) {

	switch concurrencyBy {

	case "person":

		masterRecordsToAdd := make(chan []MasterRecord)

		//create pointer to a new local set of inputs for each independent thread
		var localInputs Input
		localInputs = deepCopy(Inputs)

		for _, person := range Inputs.People { // foreach cycle
			go runModelWithConcurrentPeople(localInputs, person, masterRecordsToAdd)
		} // end foreach cycle

		for _, person := range Inputs.People {
			mRstoAdd := <-masterRecordsToAdd
			GlobalMasterRecords = append(GlobalMasterRecords, mRstoAdd...)
			for _, mRtoAdd := range mRstoAdd {
				//GlobalMasterRecordsByIPCM[0][mRtoAdd.Person_id][mRtoAdd.Cycle_id][mRtoAdd.Model_id] = mRtoAdd.State_id
				_ = mRtoAdd
			}

			_ = person
		}

		// case "person-within-cycle":

		// 	for _, cycle := range Cycles { // foreach cycle
		// 		for _, person := range People { // 	foreach person
		// 			runModelWithConcurrentPeopleWithinCycle(person, cycle)
		// 		}
		// 		CurrentCycle++
		// 	}

	} // end case

	fmt.Println("Time elapsed, excluding data export:", fmt.Sprint(time.Since(beginTime)))

	//outputs
	toCsv(output_dir+"/master.csv", GlobalMasterRecords[0], GlobalMasterRecords)
	toCsv(output_dir+"/states.csv", Inputs.States[0], Inputs.States)

	fmt.Println("Time elapsed, including data export:", fmt.Sprint(time.Since(beginTime)))

	iterationChan <- "Done"

}

func deepCopy(Inputs Input) Input {

	var mod bytes.Buffer
	enc := gob.NewEncoder(&mod)
	dec := gob.NewDecoder(&mod)

	err := enc.Encode(Inputs)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	var cpy Input
	err = dec.Decode(&cpy)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	return cpy
}

func runModelWithConcurrentPeople(localInputs Input, person Person, masterRecordsToAdd chan []MasterRecord) {

	localInputsPointer := &localInputs

	mrSize := len(localInputsPointer.Cycles) * len(localInputsPointer.Models)
	theseMasterRecordsToAdd := make([]MasterRecord, mrSize, mrSize)
	mrIndex := 0
	//fmt.Println("Person:", person.Id)
	for _, cycle := range localInputsPointer.Cycles { // foreach cycle
		//fmt.Println("Cycle: ", cycle.Name)
		//shuffled := shuffle(localInputsPointer.Models) // randomize the order of the models //TODO place back in not sure why broken.
		for _, model := range localInputsPointer.Models { // foreach model
			//fmt.Println(model.Name)

			// get the current state of the person in this model (should be
			// the uninitialized state for cycle 0)
			currentStateInThisModel := person.get_state_by_model(localInputsPointer, model)
			//stateToReturnId := localInputs.QueryData.State_id_by_cycle_and_person_and_model[localInputs.CurrentCycle][person.Id][model.Id]

			//fmt.Println("Current state in this model: ", currentStateInThisModel.Id)

			// get the transition probabilities from the given state
			transitionProbabilities := currentStateInThisModel.get_destination_probabilites(localInputsPointer)

			check_sum(transitionProbabilities) // will throw error if sum isn't 1

			// get all states this person is in in current cycle
			//fmt.Println("now get all states")
			states := person.get_states(localInputsPointer)

			//fmt.Println("All states this person is in: ", states)

			// get any interactions that will effect the transtion from
			// the persons current states based on all states that they are
			// in - it is a method of their current state in this model,
			// and accepts an array of all currents states they occupy
			interactions := currentStateInThisModel.get_relevant_interactions(localInputsPointer, states)

			if len(interactions) > 0 { // if there are interactions

				for _, interaction := range interactions { // foreach interaction
					// apply the interactions to the transition probabilities
					transitionProbabilities = adjust_transitions(localInputsPointer, transitionProbabilities, interaction)
				} // end foreach interaction

			} // end if there are interactions

			check_sum(transitionProbabilities) // will throw error if sum isn't 1

			// using  final transition probabilities, assign new state to person
			new_state := pickState(localInputsPointer, transitionProbabilities)
			//fmt.Println("New state is", new_state.Id)

			if new_state.Id < 1 {
				fmt.Println("No new state!")
				os.Exit(1)
			}

			if localInputsPointer.CurrentCycle != cycle.Id {
				fmt.Println("cycle mismatch!")
				os.Exit(1)
			}

			// store new state in master object
			err := add_master_record(localInputsPointer, cycle, person, new_state)
			localInputsPointer.QueryData.State_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id] = new_state.Id

			check_new_state_id := localInputsPointer.QueryData.State_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id]

			if check_new_state_id != new_state.Id {
				fmt.Println("Was not correctly assigned... bug")
				os.Exit(1)
			}

			var newMasterRecord MasterRecord
			newMasterRecord.Cycle_id = cycle.Id + 1
			newMasterRecord.Person_id = person.Id
			newMasterRecord.State_id = new_state.Id
			newMasterRecord.Model_id = model.Id

			//fmt.Println("setting c p m", newMasterRecord.Cycle_id, newMasterRecord.Person_id, newMasterRecord.Model_id, "to", newMasterRecord.State_id)

			theseMasterRecordsToAdd[mrIndex] = newMasterRecord
			mrIndex++

			if err != false {
				fmt.Println("problem adding master record")
				os.Exit(1)
			} else {
				//fmt.Println("master updated")
			}

		} // end foreach model
		localInputsPointer.CurrentCycle++

	} //end foreach cycle

	//Timer := nitro.Initialize()

	masterRecordsToAdd <- theseMasterRecordsToAdd

}

// func runModelWithConcurrentPeopleWithinCycle(person Person, cycle Cycle) {
// 	fmt.Println("Person:", person.Id)
// 	fmt.Println("Cycle: ", cycle.Name)
// 	shuffled := shuffle(Models) // randomize the order of the models
// 	for _, model := range shuffled {        // foreach model
// 		runPersonCycleModel(person, cycle, model)
// 	} // end foreach model
// }

func setUpQueryData(Inputs Input, numberOfPeople int) Input {
	// Need to have lengths to be able to access them
	//Cycles
	Inputs.QueryData.State_id_by_cycle_and_person_and_model = make([][][]int, len(Inputs.Cycles)+1, len(Inputs.Cycles)+1)
	for i, _ := range Inputs.QueryData.State_id_by_cycle_and_person_and_model {
		//People
		Inputs.QueryData.State_id_by_cycle_and_person_and_model[i] = make([][]int, numberOfPeople, numberOfPeople)
		for p, _ := range Inputs.QueryData.State_id_by_cycle_and_person_and_model[i] {
			Inputs.QueryData.State_id_by_cycle_and_person_and_model[i][p] = make([]int, len(Inputs.Models), len(Inputs.Models))
		}
	}

	//Cycles
	//Inputs.QueryData.States_ids_by_cycle_and_person = make([][]int, 1000000, 1000000)

	Inputs.QueryData.Tps_id_by_from_state = make([][]int, len(Inputs.States), len(Inputs.States))
	for i, _ := range Inputs.QueryData.Tps_id_by_from_state {
		var tPIdsToReturn []int
		for _, transitionProbability := range Inputs.TransitionProbabilities {
			if transitionProbability.From_id == i {
				tPIdsToReturn = append(tPIdsToReturn, transitionProbability.Id)
			}
		}
		Inputs.QueryData.Tps_id_by_from_state[i] = tPIdsToReturn
	}

	Inputs.QueryData.Interactions_id_by_in_state_and_model = make([][]int, len(Inputs.States), len(Inputs.States))
	for i, _ := range Inputs.QueryData.Interactions_id_by_in_state_and_model {
		Inputs.QueryData.Interactions_id_by_in_state_and_model[i] = make([]int, len(Inputs.Models), len(Inputs.Models))
		for r := 0; r < len(Inputs.Models); r++ {
			Inputs.QueryData.Interactions_id_by_in_state_and_model[i][r] = 99999999 // placeholder value to represent no interaction
		}
	}

	for _, interaction := range Inputs.Interactions {
		// if person is in a state with an interaction that effects current model
		Inputs.QueryData.Interactions_id_by_in_state_and_model[interaction.In_state_id][interaction.Effected_model_id] = interaction.Id
	}

	Timer.Step("set up query data")
	return Inputs
}

func setUpGlobalMasterRecordsByIPCM(Inputs Input) {

	GlobalMasterRecordsByIPCM = make([][][][]int, numberOfIterations, numberOfIterations)
	for i := 0; i < numberOfIterations; i++ {
		GlobalMasterRecordsByIPCM[i] = make([][][]int, numberOfPeople, numberOfPeople)
		for p := 0; p < numberOfPeople; p++ {
			GlobalMasterRecordsByIPCM[i][p] = make([][]int, len(Inputs.Cycles), len(Inputs.Cycles))
			for q := 0; q < len(Inputs.Cycles); q++ {
				GlobalMasterRecordsByIPCM[i][p][q] = make([]int, len(Inputs.Models), len(Inputs.Models))
			}
		}
	}

	Timer.Step("set up master data")
}

// ----------- non-methods

func shuffle(models []Model) []Model {
	//randomize order of models
	for i := range models {
		j := rand.Intn(i + 1)
		models[i], models[j] = models[j], models[i]
	}
	return models
}

// create people will generate individuals and add their data to the master
// records
func createPeople(Inputs Input, number int) Input {
	for i := 0; i < number; i++ {
		Inputs.People = append(Inputs.People, Person{i})
	}

	for _, person := range Inputs.People {
		for _, model := range Inputs.Models {
			uninitializedState := model.get_uninitialized_state(Inputs)
			var mr MasterRecord
			mr.Cycle_id = 0
			mr.State_id = uninitializedState.Id
			mr.Model_id = model.Id
			mr.Person_id = person.Id
			// generate a hash key for a map, allows easy access to states
			// by hashing cycle, person and model.
			qd := Inputs.QueryData.State_id_by_cycle_and_person_and_model

			qd[mr.Cycle_id][mr.Person_id][mr.Model_id] = mr.State_id

			Inputs.MasterRecords = append(Inputs.MasterRecords, mr)

			// fmt.Println("setting c p m", mr.Cycle_id, mr.Person_id, mr.Model_id, "to", Inputs.QueryData.State_id_by_cycle_and_person_and_model[mr.Cycle_id][mr.Person_id][mr.Model_id])

			//State_id_by_cycle_and_person_and_model
			//States_ids_by_cycle_and_person

		}
	}

	Timer.Step("set up people")

	return Inputs

}

// get state by id
func get_state_by_id(localInputs *Input, stateId int) State {

	theState := localInputs.States[stateId]

	if theState.Id == stateId {
		return theState
	}

	fmt.Println("Cannot find state by id ", stateId)
	os.Exit(1)
	return theState
	// var state State
	// for _, state := range States {
	// 	if state.Id == stateId {
	// 		return state
	// 	}
	// }

	// return state
}

// ------------------------------------------- methods

// --------------- transition probabilities

func adjust_transitions(localInputs *Input, theseTPs []TransitionProbability, interaction Interaction) []TransitionProbability {
	// TODO if these ever change to pointerss, you'll need to deference them
	adjustmentFactor := interaction.Adjustment
	for i, _ := range theseTPs {
		tp := &theseTPs[i] // TODO don't really understand why this works
		originalTpBase := tp.Tp_base
		if tp.From_id == interaction.From_state_id && tp.To_id == interaction.To_state_id {
			tp.Tp_base = tp.Tp_base * adjustmentFactor
			if tp.Tp_base == originalTpBase && adjustmentFactor != 1 {
				fmt.Println("error adjusting transition probabilities in adjust_transitions()")
				fmt.Println("interaction id is: ", interaction.Id)
				os.Exit(1)
			}
		}
	}
	// now, we need to make sure everything adds to one. to do so, we find what
	// it currently sums to, and make a new adjustment factor. We can then
	// adjust every transition probability by that amount.
	sum := get_sum(theseTPs)
	newAdjFactor := float64(1) / sum

	for i, _ := range theseTPs {
		tp := &theseTPs[i] // TODO don't really understand why this works
		tp.Tp_base = tp.Tp_base * newAdjFactor
	}
	return theseTPs
}

func check_sum(theseTPs []TransitionProbability) {
	sum := get_sum(theseTPs)

	if !equalFloat(sum, 1.0, 0.00000001) {
		fmt.Println("sum does not equal 1 !")
		os.Exit(1)
	}
}

func get_sum(theseTPs []TransitionProbability) float64 {
	sum := float64(0.0)
	for _, tp := range theseTPs {
		sum += tp.Tp_base
	}
	return sum
}

// EqualFloat() returns true if x and y are approximately equal to the
// given limit. Pass a limit of -1 to get the greatest accuracy the machine
// can manage.
func equalFloat(x float64, y float64, limit float64) bool {

	if limit <= 0.0 {
		limit = math.SmallestNonzeroFloat64
	}

	return math.Abs(x-y) <= (limit * math.Min(math.Abs(x), math.Abs(y)))
}

func pause() {
	time.Sleep(1000000000)
}

// --------------- person

// get the current state of the person in this model (should be the uninitialized state for cycle 0)
func (thisPerson *Person) get_state_by_model(localInputs *Input, thisModel Model) State {
	thisModelId := thisModel.Id
	var stateToReturn State
	var stateToReturnId int

	stateToReturnId = localInputs.QueryData.State_id_by_cycle_and_person_and_model[localInputs.CurrentCycle][thisPerson.Id][thisModelId]

	if localInputs.CurrentCycle != 0 && stateToReturnId == 0 {
		fmt.Println("unint state after cycle 0!")
	}

	stateToReturn = localInputs.States[stateToReturnId]
	if stateToReturn.Id == stateToReturnId {
		return stateToReturn
	}
	fmt.Println("Cannot find state via get_state_by_model, error 2")
	os.Exit(1)
	return stateToReturn
}

// get all states this person is in at the current cycle
func (thisPerson *Person) get_states(localInputs *Input) []State {
	thisPersonId := thisPerson.Id

	//fmt.Println("getting all states of cycle and person", localInputs.CurrentCycle, thisPersonId)

	statesToReturnIds := localInputs.QueryData.State_id_by_cycle_and_person_and_model[localInputs.CurrentCycle][thisPersonId]

	statesToReturn := make([]State, len(statesToReturnIds), len(statesToReturnIds))

	for i, statesToReturnId := range statesToReturnIds {
		if localInputs.States[statesToReturnId].Id == statesToReturnId {
			statesToReturn[i] = localInputs.States[statesToReturnId]
		} else {
			fmt.Println("cannot find states via get_states, cycle & person id =", localInputs.CurrentCycle, thisPersonId)
			fmt.Println("looking for id", statesToReturnId, "but found", localInputs.States[statesToReturnId].Id)
			os.Exit(1)
		}
	}

	if len(statesToReturn) > 0 {
		return statesToReturn
	} else {
		fmt.Println("cannot find states via get_states")
		os.Exit(1)
		return statesToReturn
	}

}

//  --------------- model

// gets the uninitialized state for a model (the state individuals start in)
func (model *Model) get_uninitialized_state(Inputs Input) State {
	modelId := model.Id
	for _, state := range Inputs.States {
		if state.Model_id == modelId && state.Is_uninitialized_state == true {
			return state
		}
	}
	fmt.Println("cannot find uninitialized state by get_uninitialized_state for model ", model.Id)
	os.Exit(1)
	return State{}
}

//  --------------- state

// get the transition probabilities *from* the given state. It's called
// destination because we're finding the chances of moving to each destination
func (state *State) get_destination_probabilites(localInputs *Input) []TransitionProbability {
	var tPIdsToReturn []int
	tPIdsToReturn = localInputs.QueryData.Tps_id_by_from_state[state.Id]
	tPsToReturn := make([]TransitionProbability, len(tPIdsToReturn), len(tPIdsToReturn))
	for i, id := range tPIdsToReturn {
		tPsToReturn[i] = localInputs.TransitionProbabilities[id]
	}
	if len(tPsToReturn) > 0 {
		return tPsToReturn
	} else {
		///fmt.Println("cannot find destination probabilities via get_destination_probabilites")
		os.Exit(1)
		return tPsToReturn
	}
}

// get any interactions that will effect the transtion from
// the persons current states based on all states that they are
// in - it is a method of their current state in this model,
// and accepts an array of all currents states they occupy
func (inState *State) get_relevant_interactions(localInputs *Input, allStates []State) []Interaction {
	modelId := inState.Model_id

	var relevantInteractions []Interaction
	for _, alsoInState := range allStates {
		relevantInteractionId := localInputs.QueryData.Interactions_id_by_in_state_and_model[alsoInState.Id][modelId]
		if relevantInteractionId != 99999999 {
			if relevantInteractionId == localInputs.Interactions[relevantInteractionId].Id {
				relevantInteractions = append(relevantInteractions, localInputs.Interactions[relevantInteractionId])
			} else {
				fmt.Println("off-by-one error or similar in get_relevant_interactions")
			}
		}
	}

	return relevantInteractions
}

// //transition probabilities
// transition_probability.adjust_transitions(interaction) // apply the interactions to the transition probabilities

// //master

// store new state in master object for n+1 cycle (note that the cycle is
// auto - incremented within this function)
func add_master_record(localInputs *Input, cycle Cycle, person Person, newState State) bool {
	ogLen := len(localInputs.MasterRecords)
	var newMasterRecord MasterRecord
	newMasterRecord.Cycle_id = cycle.Id + 1
	newMasterRecord.Person_id = person.Id
	newMasterRecord.State_id = newState.Id
	newMasterRecord.Model_id = newState.Model_id

	localInputs.QueryData.State_id_by_cycle_and_person_and_model[newMasterRecord.Cycle_id][newMasterRecord.Person_id][newMasterRecord.Model_id] = newMasterRecord.State_id

	_ = ogLen

	return false
}

// Using  the final transition probabilities, pickState assigns a new state to
// a person. It is given many states and returns one.
func pickState(localInputs *Input, tPs []TransitionProbability) State {
	probs := make([]float64, len(tPs), len(tPs))
	for i, tP := range tPs {
		probs[i] = tP.Tp_base
	}

	chosenIndex := pick(probs)
	stateId := tPs[chosenIndex].To_id
	if stateId == 0 {
		fmt.Println("error!! ")
		os.Exit(1)
	}

	state := get_state_by_id(localInputs, stateId)

	if &state != nil {
		return state
	} else {
		fmt.Println("cannot pick state with pickState")
		os.Exit(1)
		return state
	}

}

// iterates over array of potential states and uses a random value to find
// where new state is. returns new state id.
func pick(probabilities []float64) int {
	random := rand.Float64()
	sum := float64(0.0)
	for i, prob := range probabilities { //for i := 0; i < len(probabilities); i++ {
		sum += prob
		if random <= sum {
			return i
		}
	}
	// TODO(alex): figure this out - needed error of something
	fmt.Println("problem with pick")
	os.Exit(1)
	return 0
}

// Exports sets of data to CSVs. I particular, it will print any array of structs
// and automatically uses the struct field names as headers! wow.
// It takes a filename, as well one copy of the struct, and the array of structs
// itself.
func toCsv(filename string, record interface{}, records interface{}) error {
	fmt.Println("Beginning export process to ", filename)
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
	fmt.Println("Exported to ", filename)
	writer.Flush()
	return err
}

func getNumberOfRecords(filename string) int {
	csvFile, err := os.Open(filename)
	r := csv.NewReader(csvFile)
	lines, err := r.ReadAll()
	if err != nil {
		log.Fatalf("error reading all lines: %v", err)
	}
	return len(lines) - 1
}

// fromCSV accepts a filename of a properly-formatted CSV and write the content
// of that CSV into pointers. it returns an array of pointers in an interface
// which are used later to convert them to their correct structures.
// adapted from http://stackoverflow.com/questions/20768511/unmarshal-csv-record-into-struct-in-go
func fromCsv(filename string, record interface{}, recordPtrs []interface{}) []interface{} {

	fmt.Println("Beginning import process from ", filename)

	//open file
	csvFile, err := os.Open(filename)
	r := csv.NewReader(csvFile)
	lines, err := r.ReadAll()
	if err != nil {
		log.Fatalf("error reading all lines: %v", err)
	}
	// use the single record to determine the fields of the struct
	val := reflect.Indirect(reflect.ValueOf(record))
	numberOfFields := val.Type().NumField()
	var fieldNames []string
	for i := 0; i < numberOfFields; i++ {
		fieldNames = append(fieldNames, val.Type().Field(i).Name)
	}
	//check to make sure header CSV and structs use the same order
	for i, _ := range lines[0] {
		if lines[0][i] != fieldNames[i] {
			fmt.Println("fatal: CSV fields in wrong order", filename)
			fmt.Println(lines[0][i], fieldNames[i])
			os.Exit(1)
		}
	}
	// toReturn is where all the pointers will go
	var toReturn []interface{}
	for q, line := range lines {
		// skip first row, just the headers. use q-1 to reference the
		// pointer in the array, because of the difference in indicies
		// (there is no header row in the pointer data, just the CSV)
		if q > 0 {
			for i := 0; i < numberOfFields; i++ {
				f := reflect.ValueOf(recordPtrs[q-1]).Elem().Field(i)
				switch f.Type().String() {
				case "string":
					f.SetString(line[i])
					//f.SetString(line[i])
				case "int":
					ival, err := strconv.ParseInt(line[i], 10, 0)
					if err != nil {
						fmt.Println("error converting to int!", err)
						os.Exit(1)
					}
					f.SetInt(ival)
				case "float64":
					ival, err := strconv.ParseFloat(line[i], 64)
					if err != nil {
						fmt.Println("error converting to float!", err)
						os.Exit(1)
					}
					f.SetFloat(ival)
				case "bool":
					ival, err := strconv.ParseBool(line[i])
					if err != nil {
						fmt.Println("error converting to bool!", err)
						os.Exit(1)
					}
					f.SetBool(ival)
				default:
					fmt.Println("error with import - not acceptable type")
					os.Exit(1)
				}
			}
			toReturn = append(toReturn, recordPtrs[q-1])
		}
	}
	return toReturn
}

func initializeInputs(Inputs Input, inputsPath string) Input {

	//get the correct csvs
	Inputs.CurrentCycle = 0

	// ####################### Models #######################

	// initialize inputs, needed for fromCsv function
	filename := "inputs/" + inputsPath + "/models.csv"
	numberOfRecords := getNumberOfRecords(filename)
	Inputs.Models = make([]Model, numberOfRecords, numberOfRecords)
	var ptrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		ptrs = append(ptrs, new(Model))
	}
	ptrs = fromCsv(filename, Inputs.Models[0], ptrs)
	for i, ptr := range ptrs {
		Inputs.Models[i] = *ptr.(*Model)
	}

	// ####################### States #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/states.csv"
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

	return Inputs
}
