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
	"math/rand"
	"os"
	"reflect"
	"time"
)

type State struct {
	Id       int
	Model_id int
	Name     string
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
	Id            int
	In_state_id   int
	From_state_id int
	To_state_id   int
	Adjustment    float32
}

type TransitionProbability struct {
	Id      int
	From_id int
	To_id   int
	Tp_base float32
}

// these are all global variables, which is why they are Capitalized

var CurrentCycle = 0

var Models = []Model{
	Model{1, "HIV"},
	Model{2, "TB"}}

var People = []Person{
	Person{1},
	Person{2},
	Person{3},
	Person{4},
	Person{5},
	Person{6},
	Person{7},
	Person{8},
	Person{9},
	Person{10}}

var States = []State{
	State{1, 1, "Uninit"},
	State{2, 1, "HIV-"},
	State{3, 1, "HIV+"},
	State{4, 2, "Uninit"},
	State{5, 2, "TB-"},
	State{6, 2, "TB+"}}

var TransitionProbabilities = []TransitionProbability{
	TransitionProbability{1, 1, 1, 0},
	TransitionProbability{2, 1, 2, 0.99},
	TransitionProbability{3, 1, 3, 0.1},
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

var Interactions = []Interaction{Interaction{1, 3, 5, 6, 2}}

var Cycles = []Cycle{
	Cycle{0, "Pre-initialization"},
	Cycle{1, "2015"},
	Cycle{2, "2016"},
	Cycle{3, "2017"},
	Cycle{4, "2018"},
	Cycle{5, "2019"}}

var MasterRecords = []MasterRecord{
	MasterRecord{0, 1, 1, 1},
	MasterRecord{0, 1, 4, 2},
	MasterRecord{0, 2, 1, 1},
	MasterRecord{0, 2, 4, 2},
	MasterRecord{0, 3, 1, 1},
	MasterRecord{0, 3, 4, 2},
	MasterRecord{0, 4, 1, 1},
	MasterRecord{0, 4, 4, 2},
	MasterRecord{0, 5, 1, 1},
	MasterRecord{0, 5, 4, 2},
	MasterRecord{0, 6, 1, 1},
	MasterRecord{0, 6, 4, 2},
	MasterRecord{0, 7, 1, 1},
	MasterRecord{0, 7, 4, 2},
	MasterRecord{0, 8, 1, 1},
	MasterRecord{0, 8, 4, 2},
	MasterRecord{0, 9, 1, 1},
	MasterRecord{0, 9, 4, 2},
	MasterRecord{0, 10, 1, 1},
	MasterRecord{0, 10, 4, 2}}

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	// table tests here

	// population initial cycle

	// current refers to the current cycle, which is used to calculate the next cycel

	for _, cycle := range Cycles { // foreach cycle
		fmt.Println("Cycle: ", cycle.Name)
		for _, person := range People { // 	foreach person
			fmt.Println("Person:", person.Id)
			shuffled := shuffle(Models)      // randomize the order of the models
			for _, model := range shuffled { // foreach model

				fmt.Println(model.Name)

				_ = model
				_ = person
				_ = States
				_ = TransitionProbabilities
				_ = Interactions
				_ = cycle
				_ = MasterRecords

				// get the current state of the person in this model (should be the uninitialized state for cycle 0)
				current_state_in_this_model := person.get_state_by_model(model)

				fmt.Println("Current state in this model: ", current_state_in_this_model.Id)

				// get the transition probabilities from the given state
				transition_probabilities := current_state_in_this_model.get_destination_probabilites()
				_ = transition_probabilities

				// get all states this person is in in current cycle
				states := person.get_states()

				fmt.Println("All states this person is in: ", states)

				// // get any interactions that will effect this model from the persons current states
				// interactions := model.get_interactions_by_states(states)

				// if len(interactions) > 0 { // if there are interactions
				// 	for _, interaction := range interactions { // foreach interaction
				// 		// apply the interactions to the transition probabilities
				// 		transition_probability = transition_probability.adjust_transitions(interaction)
				// 	} // end foreach interaction
				// } // end if there are interactions

				// using  final transition probabilities, assign new state to person
				new_state := pickState(transition_probabilities)
				fmt.Println("New state is", new_state.Id)

				// store new state in master object
				err := add_master_record(cycle, person, new_state)
				if err != false {
					fmt.Println("problem adding master record")
					os.Exit(1)
				} else {
					fmt.Println("master updated")
				}

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

// func pickState(states []State) State {

// }

// functions needed

// // non-methods

func shuffle(models []Model) []Model {
	//randomize order of models
	for i := range models {
		j := rand.Intn(i + 1)
		models[i], models[j] = models[j], models[i]
	}
	return models
}

// //methods

// //person

// get the current state of the person in this model (should be the uninitialized state for cycle 0)
func (thisPerson *Person) get_state_by_model(thisModel Model) State {
	thisModelId := thisModel.Id
	var stateToReturn State
	for _, masterRecord := range MasterRecords {
		if masterRecord.Model_id == thisModelId && masterRecord.Cycle_id == CurrentCycle && masterRecord.Person_id == thisPerson.Id {
			stateToReturn = get_state_by_id(masterRecord.State_id)
			return stateToReturn
		}
	}
	fmt.Println("Cannot find state via get_state_by_model")
	os.Exit(1)
	return stateToReturn
}

// get state by id
func get_state_by_id(stateId int) State {
	var state State
	for _, state := range States {
		if state.Id == stateId {
			return state
		}
	}
	fmt.Println("Cannot find state by id ", stateId)
	os.Exit(1)
	return state
}

// get all states this person is in at the current cycle
func (thisPerson *Person) get_states() []State {
	thisPersonId := thisPerson.Id
	var statesToReturn []State
	for _, masterRecord := range MasterRecords {
		if masterRecord.Person_id == thisPersonId && masterRecord.Cycle_id == CurrentCycle {
			stateToReturn := get_state_by_id(masterRecord.State_id)
			statesToReturn = append(statesToReturn, stateToReturn)
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

// //model

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

// model.get_interactions_by_states(states) // get any interactions that will effect this model from the persons current states

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
	var probs []float32
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
func pick(probabilities []float32) int {
	random := rand.Float32()
	sum := float32(0.0)
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

// exports sets of int data to CSVs
func toCsv(filename string, record interface{}, records interface{}) error {

	fmt.Println("Beginning export process to... ", filename)
	//os.Exit(1)

	os.Create(filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// New Csv wriier
	writer := csv.NewWriter(file)

	val := reflect.Indirect(reflect.ValueOf(record))
	numberOfFields := val.Type().NumField()
	var fieldNames []string
	for i := 0; i < numberOfFields; i++ {
		fieldNames = append(fieldNames, val.Type().Field(i).Name)
	}
	err = writer.Write(fieldNames)

	val2 := reflect.ValueOf(records)
	for i := 0; i < val2.Len(); i++ {
		var line []string
		for p := 0; p < numberOfFields; p++ {
			//convert interface to string
			line = append(line, fmt.Sprintf("%v", val2.Index(i).Field(p).Interface()))
		}
		err = writer.Write(line)
	}

	//val := reflect.Indirect(reflect.ValueOf(a))
	//    fmt.Println(val..Type().Name())

	//fmt.Println(nval)

	if err != nil {
		fmt.Println("error")
	}
	writer.Flush()

	return err

}
