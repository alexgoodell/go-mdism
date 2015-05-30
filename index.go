// General to do
// * int to uint

package main

import (
	// "encoding/json"
	// "flag"
	"fmt"
	// 	"github.com/alexgoodell/ghdmodel/models"
	// 	"io/ioutil"
	// 	"net/http"
	// 	"strconv"
	"encoding/csv"
	"github.com/davecheney/profile"
	"math"
	"math/rand"
	"os"
	"reflect"

	"time"
)

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
	// generate a hash key for a map, allows easy access to states
	// by hashing cycle, person and model.
	Hash_key string
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
	//State_by_cycle_and_person_and_model
	State_id_by_cycle_and_person_and_model map[string]int
	States_ids_by_cycle_and_person         map[string]int
	Tps_id_by_from_state                   map[string]int
	Interactions_id_by_in_state_and_model  map[string]int
}

// these are all global variables, which is why they are Capitalized
// current refers to the current cycle, which is used to calculate the next cycle
var CurrentCycle = 1

var QueryData = Query{}

var Models = []Model{
	Model{1, "HIV"},
	Model{2, "TB"}}

var People []Person

var States = []State{
	State{1, 1, "Uninit", true},
	State{2, 1, "HIV-", false},
	State{3, 1, "HIV+", false},
	State{4, 2, "Uninit", true},
	State{5, 2, "TB-", false},
	State{6, 2, "TB+", false}}

var TransitionProbabilities = []TransitionProbability{
	TransitionProbability{1, 1, 1, 0},
	TransitionProbability{2, 1, 2, 0.99},
	TransitionProbability{3, 1, 3, 0.01},
	TransitionProbability{4, 2, 1, 0},
	TransitionProbability{5, 2, 2, 0.95},
	TransitionProbability{6, 2, 3, 0.05},
	TransitionProbability{7, 3, 1, 0},
	TransitionProbability{8, 3, 2, 0},
	TransitionProbability{9, 3, 3, 1},
	TransitionProbability{10, 4, 4, 0},
	TransitionProbability{11, 4, 5, 0.8},
	TransitionProbability{12, 4, 6, 0.2},
	TransitionProbability{13, 5, 4, 0},
	TransitionProbability{14, 5, 5, 0.9},
	TransitionProbability{15, 5, 6, 0.1},
	TransitionProbability{16, 6, 4, 0},
	TransitionProbability{17, 6, 5, 0},
	TransitionProbability{18, 6, 6, 1}}

var Interactions = []Interaction{Interaction{1, 3, 5, 6, 2, 2}}

var Cycles = []Cycle{
	Cycle{1, "Pre-initialization"},
	Cycle{2, "2015"},
	Cycle{3, "2016"},
	Cycle{4, "2017"},
	Cycle{5, "2018"}}

var MasterRecords = []MasterRecord{}

func main() {

	cfg := profile.Config{
		MemProfile:     false,
		ProfilePath:    ".",  // store profiles in current directory
		NoShutdownHook: true, // do not hook SIGINT
		CPUProfile:     true,
	}

	defer profile.Start(&cfg).Stop()

	// create people will generate individuals and add their data to the master
	// records
	createPeople(1000)

	// Seed the random function
	rand.Seed(time.Now().UTC().UnixNano())

	// table tests here

	for _, cycle := range Cycles { // foreach cycle
		fmt.Println("Cycle: ", cycle.Name)
		for _, person := range People { // 	foreach person
			fmt.Println("Person:", person.Id)
			shuffled := shuffle(Models)      // randomize the order of the models
			for _, model := range shuffled { // foreach model

				runPersonCycleModel(person, cycle, model)

			} // end foreach model
		} // end foreach person

		fmt.Println(MasterRecords)
		toCsv("master.csv", MasterRecords[0], MasterRecords)

		toCsv("states.csv", States[0], States)

		CurrentCycle++ //move to next cycle in global variable
		//fmt.Println("Debugging stop")
		//os.Exit(1)
	} // end foreach cycle

}

func runPersonCycleModel(person Person, cycle Cycle, model Model) {

	fmt.Println(model.Name)

	// get the current state of the person in this model (should be
	// the uninitialized state for cycle 0)
	currentStateInThisModel := person.get_state_by_model(model)

	fmt.Println("Current state in this model: ", currentStateInThisModel.Id)

	// get the transition probabilities from the given state
	transitionProbabilities := currentStateInThisModel.get_destination_probabilites()

	check_sum(transitionProbabilities) // will throw error if sum isn't 1

	// get all states this person is in in current cycle
	states := person.get_states()

	fmt.Println("All states this person is in: ", states)

	// get any interactions that will effect the transtion from
	// the persons current states based on all states that they are
	// in - it is a method of their current state in this model,
	// and accepts an array of all currents states they occupy
	interactions := currentStateInThisModel.get_relevant_interactions(states)

	if len(interactions) > 0 { // if there are interactions

		for _, interaction := range interactions { // foreach interaction
			// apply the interactions to the transition probabilities
			transitionProbabilities = adjust_transitions(transitionProbabilities, interaction)
		} // end foreach interaction

	} // end if there are interactions

	check_sum(transitionProbabilities) // will throw error if sum isn't 1

	// using  final transition probabilities, assign new state to person
	new_state := pickState(transitionProbabilities)
	fmt.Println("New state is", new_state.Id)

	// store new state in master object
	err := add_master_record(cycle, person, new_state)
	if err != false {
		fmt.Println("problem adding master record")
		os.Exit(1)
	} else {
		fmt.Println("master updated")
	}
}

// ------------------------------------------- functions

// func createQueryData() {

// 	for _, cycle := range Cycles { // foreach cycle
// 		for _, person := range People { // 	foreach person
// 			for _, model := range Models { // foreach model

// QueryData.State_by_cycle_and_person_and_model

// type Query struct {
// 	State_by_cycle_and_person_and_model [][][]int
// 	States_by_cycle_and_person          [][]int
// 	Tps_by_from_state                   []int
// 	Interactions_by_in_state_and_model  [][]int
// }

// state = state_by_cycle_and_person_and_model[0][person][model]
// states = states_by_cycle_and_person[0][person]
// tps_by_from_state[from_state]
// interactions = interactions_by_in_state_and_model[in_state][model]

// }

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
func createPeople(number int) {
	for i := 0; i < number; i++ {
		People = append(People, Person{i + 1})
	}

	for _, person := range People {
		for _, model := range Models {
			uninitializedState := model.get_uninitialized_state()
			var mr MasterRecord
			mr.Cycle_id = 1
			mr.State_id = uninitializedState.Id
			mr.Model_id = model.Id
			mr.Person_id = person.Id
			// generate a hash key for a map, allows easy access to states
			// by hashing cycle, person and model.
			mr.Hash_key = fmt.Sprintf("%010d%010d%010d", mr.Cycle_id, mr.Person_id, mr.Model_id)
			MasterRecords = append(MasterRecords, mr)

		}
	}

}

// get state by id
func get_state_by_id(stateId int) State {

	theState := States[stateId-1]

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

func adjust_transitions(theseTPs []TransitionProbability, interaction Interaction) []TransitionProbability {
	// TODO if these ever change to pointerss, you'll need to deference them
	adjustmentFactor := interaction.Adjustment
	for i, _ := range theseTPs {
		tp := &theseTPs[i] // TODO don't really understand why this works
		originalTpBase := tp.Tp_base
		if tp.From_id == interaction.From_state_id && tp.To_id == interaction.To_state_id {
			tp.Tp_base = tp.Tp_base * adjustmentFactor
			if tp.Tp_base == originalTpBase {
				fmt.Println("error adjusting transition probabilities in adjust_transitions()")
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

	if !equalFloat(sum, 1.0, 0) {
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
func (thisPerson *Person) get_state_by_model(thisModel Model) State {
	thisModelId := thisModel.Id
	var stateToReturn State
	var stateToReturnId int
	var bestGuess MasterRecord
	// MasterRecords is organized as such: Cycle_id, Person_id, Model_id
	fmt.Println("looking for cycle, person, model", CurrentCycle, thisPerson.Id, thisModelId)
	// this is tricky; because models are run in a random order, and they place
	// their results into MasterRecords in the order in which they are run, the
	// end part of MasterResults is unpredictable. Therefore, we just grab all
	// the MasterRecords which may fit, and test them. TODO: perhaps design a
	// system by which add_to_master_record would lay down results in model-order
	// as opposed to the order in which they are run.
	bestGuessStartingIndex := (CurrentCycle-1)*(len(People)*len(Models)) + (thisPerson.Id-1)*len(Models)
	for i, _ := range Models {
		if MasterRecords[bestGuessStartingIndex+i].Model_id == thisModelId && MasterRecords[bestGuessStartingIndex+i].Cycle_id == CurrentCycle && MasterRecords[bestGuessStartingIndex+i].Person_id == thisPerson.Id {
			bestGuess = MasterRecords[bestGuessStartingIndex+i]

		}
		fmt.Println(bestGuessStartingIndex + i)
		fmt.Printf("%+v\n", bestGuess)
	}

	if bestGuess.Model_id == thisModelId && bestGuess.Cycle_id == CurrentCycle && bestGuess.Person_id == thisPerson.Id {
		stateToReturnId = bestGuess.State_id
	} else {
		fmt.Println("Cannot find state via get_state_by_model, error 1")
		os.Exit(1)
	}
	stateToReturn = States[stateToReturnId-1]
	if stateToReturn.Id == stateToReturnId {
		return stateToReturn
	}
	fmt.Println("Cannot find state via get_state_by_model, error 2")
	os.Exit(1)
	return stateToReturn
}

// get all states this person is in at the current cycle
func (thisPerson *Person) get_states() []State {
	thisPersonId := thisPerson.Id
	var bestGuessIds []int
	var statesToReturn []State
	// MasterRecords is organized as such: Cycle_id, Person_id, Model_id
	bestGuessStartingIndex := (CurrentCycle-1)*(len(People)*len(Models)) + (thisPersonId-1)*len(Models)

	for i, _ := range Models {
		if MasterRecords[bestGuessStartingIndex+i].Person_id == thisPersonId && MasterRecords[bestGuessStartingIndex+i].Cycle_id == CurrentCycle {
			bestGuessIds = append(bestGuessIds, MasterRecords[bestGuessStartingIndex+i].State_id)
		} else {
			fmt.Println("Cannot find master records via get_states")
			os.Exit(1)
		}
	}

	for _, bestGuessId := range bestGuessIds {
		if States[bestGuessId-1].Id == bestGuessId {
			statesToReturn = append(statesToReturn, States[bestGuessId-1])
		} else {
			fmt.Println("cannot find states via get_states")
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
func (model *Model) get_uninitialized_state() State {
	modelId := model.Id
	for _, state := range States {
		if state.Model_id == modelId && state.Is_uninitialized_state == true {
			return state
		}
	}
	fmt.Println("cannot find uninitialized state by get_uninitialized_state")
	os.Exit(1)
	return State{}
}

//  --------------- state

// get the transition probabilities *from* the given state. It's called
// destination because we're finding the chances of moving to each destination
func (state *State) get_destination_probabilites() []TransitionProbability {
	var tPsToReturn []TransitionProbability

	for _, transitionProbability := range TransitionProbabilities {
		if transitionProbability.From_id == state.Id {
			tPsToReturn = append(tPsToReturn, transitionProbability)
		}
	}

	if len(tPsToReturn) > 0 {
		return tPsToReturn
	} else {
		fmt.Println("cannot find destination probabilities via get_destination_probabilites")
		os.Exit(1)
		return tPsToReturn
	}

}

// get any interactions that will effect the transtion from
// the persons current states based on all states that they are
// in - it is a method of their current state in this model,
// and accepts an array of all currents states they occupy
func (inState *State) get_relevant_interactions(allStates []State) []Interaction {
	var relevantInteractions []Interaction
	for _, alsoInState := range allStates {
		for _, interaction := range Interactions {
			// if person is in a state with an interaction that effects current model
			if interaction.From_state_id == inState.Id && interaction.In_state_id == alsoInState.Id {
				relevantInteractions = append(relevantInteractions, interaction)
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
func add_master_record(cycle Cycle, person Person, newState State) bool {
	ogLen := len(MasterRecords)
	var newMasterRecord MasterRecord
	newMasterRecord.Cycle_id = cycle.Id + 1
	newMasterRecord.Person_id = person.Id
	newMasterRecord.State_id = newState.Id
	newMasterRecord.Model_id = newState.Model_id

	MasterRecords = append(MasterRecords, newMasterRecord)
	newLen := len(MasterRecords)
	if (newLen - ogLen) == 1 { //added one record
		return false //no error
	} else {
		return true //error
	}
}

// Using  the final transition probabilities, pickState assigns a new state to
// a person. It is given many states and returns one.
func pickState(tPs []TransitionProbability) State {
	var probs []float64
	for _, tP := range tPs {
		probs = append(probs, tP.Tp_base)
	}

	chosenIndex := pick(probs)
	stateId := tPs[chosenIndex].To_id
	state := get_state_by_id(stateId)

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
