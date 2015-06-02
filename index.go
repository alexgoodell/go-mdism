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
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"runtime"
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
	// generate a hash key for a map, allows easy access to states
	// by hashing cycle, person and model.
	//Hash_key string
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
	State_id_by_cycle_and_person_and_model [][][]int
	States_ids_by_cycle_and_person         [][]int
	Tps_id_by_from_state                   []int
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

var output_dir = "tmp"

func main() {

	var Inputs Input
	Inputs = initializeInputs(Inputs)

	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("using ", runtime.NumCPU(), " cores")
	pause()
	pause()
	// Seed the random function

	numberOfPeople := 5000

	//set up queryData
	Inputs = setUpQueryData(Inputs, numberOfPeople)

	// create people will generate individuals and add their data to the master
	// records
	Inputs = createPeople(Inputs, numberOfPeople)

	// table tests here

	concurrencyBy := "person"
	runModel(Inputs, concurrencyBy)

}

func initializeInputs(Inputs Input) Input {

	Inputs.CurrentCycle = 0

	Inputs.Models = []Model{
		Model{0, "HIV"},
		Model{1, "TB"}}

	Inputs.States = []State{
		State{0, 0, "Uninit HIV", true},
		State{1, 0, "HIV-", false},
		State{2, 0, "HIV+", false},
		State{3, 1, "Uninit TB", true},
		State{4, 1, "TB-", false},
		State{5, 1, "TB+", false}}

	Inputs.TransitionProbabilities = []TransitionProbability{
		TransitionProbability{0, 0, 0, 0},
		TransitionProbability{1, 0, 1, 0.99},
		TransitionProbability{2, 0, 2, 0.01},
		TransitionProbability{3, 1, 0, 0},
		TransitionProbability{4, 1, 1, 0.95},
		TransitionProbability{5, 1, 2, 0.05},
		TransitionProbability{6, 2, 0, 0},
		TransitionProbability{7, 2, 1, 0},
		TransitionProbability{8, 2, 2, 1},
		TransitionProbability{9, 3, 3, 0},
		TransitionProbability{10, 3, 4, 0.8},
		TransitionProbability{11, 3, 5, 0.2},
		TransitionProbability{12, 4, 3, 0},
		TransitionProbability{13, 4, 4, 0.9},
		TransitionProbability{14, 4, 5, 0.1},
		TransitionProbability{15, 5, 3, 0},
		TransitionProbability{16, 5, 4, 0},
		TransitionProbability{17, 5, 5, 1}}

	Inputs.Interactions = []Interaction{Interaction{0, 2, 4, 5, 2, 1}}

	Inputs.Cycles = []Cycle{
		Cycle{0, "Pre-initialization"},
		Cycle{1, "2015"},
		Cycle{2, "2016"},
		Cycle{3, "2017"},
		Cycle{4, "2018"},
		Cycle{5, "2018"},
		Cycle{6, "2018"},
		Cycle{7, "2018"},
		Cycle{8, "2018"},
		Cycle{9, "2018"},
		Cycle{10, "2018"},
		Cycle{11, "2018"},
		Cycle{12, "2018"},
		Cycle{13, "2018"},
		Cycle{14, "2018"},
		Cycle{15, "2018"},
		Cycle{16, "2018"},
		Cycle{17, "2018"},
		Cycle{18, "2018"},
		Cycle{19, "2018"}}

	return Inputs
}

func runModel(Inputs Input, concurrencyBy string) {

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
			mRtoAdd := <-masterRecordsToAdd
			GlobalMasterRecords = append(GlobalMasterRecords, mRtoAdd...)
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

	fmt.Println("Time elapsed:", fmt.Sprint(time.Since(beginTime)))

	//outputs
	//toCsv(output_dir+"/master.csv", GlobalMasterRecords[0], GlobalMasterRecords)
	//toCsv(output_dir+"/states.csv", Inputs.States[0], Inputs.States)

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
	var theseMasterRecordsToAdd []MasterRecord
	//fmt.Println("Person:", person.Id)
	for _, cycle := range localInputsPointer.Cycles { // foreach cycle
		//fmt.Println("Cycle: ", cycle.Name)
		//shuffled := shuffle(localInputsPointer.Models) // randomize the order of the models //TODO place back in not sure why broken.
		for _, model := range localInputsPointer.Models { // foreach model
			//fmt.Println(model.Name)

			// get the current state of the person in this model (should be
			// the uninitialized state for cycle 0)
			currentStateInThisModel := person.get_state_by_model(localInputsPointer, model)
			stateToReturnId := localInputs.QueryData.State_id_by_cycle_and_person_and_model[localInputs.CurrentCycle][person.Id][model.Id]

			if currentStateInThisModel.Id != stateToReturnId {
				fmt.Println("state to return function broken")
				os.Exit(1)
			}

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

			if currentStateInThisModel.Id == 2 && new_state.Id == 1 {
				fmt.Println("cured of HIV?... bug")
				os.Exit(1)
			}

			if localInputsPointer.CurrentCycle != cycle.Id {
				fmt.Println("cycle mismatch!")
				os.Exit(1)
			}

			if new_state.Id == 3 || new_state.Id == 0 {
				fmt.Println("assigning unit state... bug")
				os.Exit(1)
			}

			if localInputsPointer.CurrentCycle != 0 && currentStateInThisModel.Id == 0 {
				fmt.Println("unit state after cycle 0... bug. person, cycle, state: ", person.Id, localInputsPointer.CurrentCycle, currentStateInThisModel.Id)
				//os.Exit(1)
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

			theseMasterRecordsToAdd = append(theseMasterRecordsToAdd, newMasterRecord)

			if err != false {
				fmt.Println("problem adding master record")
				os.Exit(1)
			} else {
				fmt.Println("master updated")
			}

		} // end foreach model
		localInputsPointer.CurrentCycle++

	} //end foreach cycle

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

// generate a hash key for a map, allows easy access to states
// by hashing cycle, person and model.
// func makeHashByCyclePersonModel(cycle Cycle, person Person, model Model) string {
// 	Hash_key := fmt.Sprintf("%010d%010d%010d", cycle.Id, person.Id, model.Id)
// 	return Hash_key
// }

// func makeHashByCyclePerson(cycle Cycle, person Person) string {
// 	Hash_key := fmt.Sprintf("%010d%010d", cycle.Id, person.Id)
// 	return Hash_key
// }

// func makeHashByTpFromState(state State) string {

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
	Inputs.QueryData.States_ids_by_cycle_and_person = make([][]int, 1000000, 1000000)
	Inputs.QueryData.Interactions_id_by_in_state_and_model = make([][]int, 1000000, 1000000)
	Inputs.QueryData.Tps_id_by_from_state = make([]int, 1000000, 1000000)

	return Inputs
}

// func runPersonCycleModel(person Person, cycle Cycle, model Model) {

// 	fmt.Println(model.Name)

// 	// get the current state of the person in this model (should be
// 	// the uninitialized state for cycle 0)
// 	currentStateInThisModel := person.get_state_by_model(model)

// 	fmt.Println("Current state in this model: ", currentStateInThisModel.Id)

// 	// get the transition probabilities from the given state
// 	transitionProbabilities := currentStateInThisModel.get_destination_probabilites()

// 	check_sum(transitionProbabilities) // will throw error if sum isn't 1

// 	// get all states this person is in in current cycle
// 	fmt.Println("now get all states")
// 	states := person.get_states()

// 	fmt.Println("All states this person is in: ", states)

// 	// get any interactions that will effect the transtion from
// 	// the persons current states based on all states that they are
// 	// in - it is a method of their current state in this model,
// 	// and accepts an array of all currents states they occupy
// 	interactions := currentStateInThisModel.get_relevant_interactions(states)

// 	if len(interactions) > 0 { // if there are interactions

// 		for _, interaction := range interactions { // foreach interaction
// 			// apply the interactions to the transition probabilities
// 			transitionProbabilities = adjust_transitions(transitionProbabilities, interaction)
// 		} // end foreach interaction

// 	} // end if there are interactions

// 	check_sum(transitionProbabilities) // will throw error if sum isn't 1

// 	// using  final transition probabilities, assign new state to person
// 	new_state := pickState(transitionProbabilities)
// 	fmt.Println("New state is", new_state.Id)

// 	// store new state in master object
// 	err := add_master_record(cycle, person, new_state)
// 	if err != false {
// 		fmt.Println("problem adding master record")
// 		os.Exit(1)
// 	} else {
// 		fmt.Println("master updated")
// 	}
// }

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

			//fmt.Println("setting c p m", mr.Cycle_id, mr.Person_id, mr.Model_id, "to", Inputs.QueryData.State_id_by_cycle_and_person_and_model[mr.Cycle_id][mr.Person_id][mr.Model_id])

			//State_id_by_cycle_and_person_and_model
			//States_ids_by_cycle_and_person

		}
	}

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
func (thisPerson *Person) get_state_by_model(localInputs *Input, thisModel Model) State {
	thisModelId := thisModel.Id
	var stateToReturn State
	var stateToReturnId int
	//var bestGuess MasterRecord
	// MasterRecords is organized as such: Cycle_id, Person_id, Model_id
	//fmt.Println("looking for cycle, person, model", localInputs.CurrentCycle, thisPerson.Id, thisModelId)
	// this is tricky; because models are run in a random order, and they place
	// their results into MasterRecords in the order in which they are run, the
	// end part of MasterResults is unpredictable. Therefore, we just grab all
	// the MasterRecords which may fit, and test them. TODO: perhaps design a
	// system by which add_to_master_record would lay down results in model-order
	// as opposed to the order in which they are run.

	stateToReturnId = localInputs.QueryData.State_id_by_cycle_and_person_and_model[localInputs.CurrentCycle][thisPerson.Id][thisModelId]

	if localInputs.CurrentCycle != 0 && stateToReturnId == 0 {
		fmt.Println("unint state after cycle 0!")
	}

	// bestGuessStartingIndex := (CurrentCycle-1)*(len(People)*len(Models)) + (thisPerson.Id-1)*len(Models)
	// for i, _ := range Models {
	// 	if MasterRecords[bestGuessStartingIndex+i].Model_id == thisModelId && MasterRecords[bestGuessStartingIndex+i].Cycle_id == CurrentCycle && MasterRecords[bestGuessStartingIndex+i].Person_id == thisPerson.Id {
	// 		bestGuess = MasterRecords[bestGuessStartingIndex+i]

	// 	}
	// 	fmt.Println(bestGuessStartingIndex + i)
	// 	fmt.Printf("%+v\n", bestGuess)
	// }

	// if bestGuess.Model_id == thisModelId && bestGuess.Cycle_id == CurrentCycle && bestGuess.Person_id == thisPerson.Id {
	// 	stateToReturnId = bestGuess.State_id
	// } else {
	// 	fmt.Println("Cannot find state via get_state_by_model, error 1")
	// 	os.Exit(1)
	// }
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
	var statesToReturn []State

	//fmt.Println("getting all states of cycle and person", localInputs.CurrentCycle, thisPersonId)

	statesToReturnIds := localInputs.QueryData.State_id_by_cycle_and_person_and_model[localInputs.CurrentCycle][thisPersonId]

	for _, statesToReturnId := range statesToReturnIds {
		if localInputs.States[statesToReturnId].Id == statesToReturnId {
			statesToReturn = append(statesToReturn, localInputs.States[statesToReturnId])
		} else {
			fmt.Println("cannot find states via get_states")
			os.Exit(1)
		}
	}

	// MasterRecords is organized as such: Cycle_id, Person_id, Model_id
	// bestGuessStartingIndex := (CurrentCycle-1)*(len(People)*len(Models)) + (thisPersonId-1)*len(Models)

	// for i, _ := range Models {
	// 	if MasterRecords[bestGuessStartingIndex+i].Person_id == thisPersonId && MasterRecords[bestGuessStartingIndex+i].Cycle_id == CurrentCycle {
	// 		bestGuessIds = append(bestGuessIds, MasterRecords[bestGuessStartingIndex+i].State_id)
	// 	} else {
	// 		fmt.Println("Cannot find master records via get_states")
	// 		os.Exit(1)
	// 	}
	// }

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
	fmt.Println("cannot find uninitialized state by get_uninitialized_state")
	os.Exit(1)
	return State{}
}

//  --------------- state

// get the transition probabilities *from* the given state. It's called
// destination because we're finding the chances of moving to each destination
func (state *State) get_destination_probabilites(localInputs *Input) []TransitionProbability {
	var tPsToReturn []TransitionProbability

	for _, transitionProbability := range localInputs.TransitionProbabilities {
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
func (inState *State) get_relevant_interactions(localInputs *Input, allStates []State) []Interaction {
	var relevantInteractions []Interaction
	for _, alsoInState := range allStates {
		for _, interaction := range localInputs.Interactions {
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
func add_master_record(localInputs *Input, cycle Cycle, person Person, newState State) bool {
	ogLen := len(localInputs.MasterRecords)
	var newMasterRecord MasterRecord
	newMasterRecord.Cycle_id = cycle.Id + 1
	newMasterRecord.Person_id = person.Id
	newMasterRecord.State_id = newState.Id
	newMasterRecord.Model_id = newState.Model_id

	localInputs.QueryData.State_id_by_cycle_and_person_and_model[newMasterRecord.Cycle_id][newMasterRecord.Person_id][newMasterRecord.Model_id] = newMasterRecord.State_id

	//localInputs.MasterRecords = append(localInputs.MasterRecords, newMasterRecord)
	// newLen := len(localInputs.MasterRecords)
	// if (newLen - ogLen) == 1 { //added one record
	// 	return false //no error
	// } else {
	// 	return true //error
	// }

	_ = ogLen

	return false
}

// Using  the final transition probabilities, pickState assigns a new state to
// a person. It is given many states and returns one.
func pickState(localInputs *Input, tPs []TransitionProbability) State {
	var probs []float64
	for _, tP := range tPs {
		probs = append(probs, tP.Tp_base)
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
