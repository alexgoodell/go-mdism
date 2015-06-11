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
	//"github.com/davecheney/profile"
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

var output_dir = "tmp"

func main() {

	// cfg := profile.Config{
	// 	ProfilePath: ".", // store profiles in current directory
	// 	CPUProfile:  true,
	// }

	// defer profile.Start(&cfg).Stop()

	var Inputs Input
	Inputs = initializeInputs(Inputs)

	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("using ", runtime.NumCPU(), " cores")
	// Seed the random function

	numberOfPeople := 100
	numberOfIterations := 2

	fmt.Println("and ", numberOfPeople, "individuals")

	//set up queryData
	Inputs = setUpQueryData(Inputs, numberOfPeople)

	// create people will generate individuals and add their data to the master
	// records
	Inputs = createPeople(Inputs, numberOfPeople)

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
			mRtoAdd := <-masterRecordsToAdd
			GlobalMasterRecords = append(GlobalMasterRecords, mRtoAdd...)
			_ = person
		}

		iterationChan <- "Done"

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

			theseMasterRecordsToAdd = append(theseMasterRecordsToAdd, newMasterRecord)

			if err != false {
				fmt.Println("problem adding master record")
				os.Exit(1)
			} else {
				//fmt.Println("master updated")
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
	}

	for _, interaction := range Inputs.Interactions {
		// if person is in a state with an interaction that effects current model
		Inputs.QueryData.Interactions_id_by_in_state_and_model[interaction.In_state_id][interaction.Effected_model_id] = interaction.Id
	}

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

			// fmt.Println("setting c p m", mr.Cycle_id, mr.Person_id, mr.Model_id, "to", Inputs.QueryData.State_id_by_cycle_and_person_and_model[mr.Cycle_id][mr.Person_id][mr.Model_id])

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
			fmt.Println("cannot find states via get_states, cycle & person id =", localInputs.CurrentCycle, thisPersonId)
			fmt.Println("looking for id", statesToReturnId, "but found", localInputs.States[statesToReturnId].Id)
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
	fmt.Println("cannot find uninitialized state by get_uninitialized_state for model ", model.Id)
	os.Exit(1)
	return State{}
}

//  --------------- state

// get the transition probabilities *from* the given state. It's called
// destination because we're finding the chances of moving to each destination
func (state *State) get_destination_probabilites(localInputs *Input) []TransitionProbability {
	var tPsToReturn []TransitionProbability
	var tPIdsToReturn []int

	tPIdsToReturn = localInputs.QueryData.Tps_id_by_from_state[state.Id]

	for _, id := range tPIdsToReturn {
		tPsToReturn = append(tPsToReturn, localInputs.TransitionProbabilities[id])
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
		if relevantInteractionId == localInputs.Interactions[relevantInteractionId].Id {
			relevantInteractions = append(relevantInteractions, localInputs.Interactions[relevantInteractionId])
		} else {
			fmt.Println("off-by-one error or similar in get_relevant_interactions")
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

func initializeInputs(Inputs Input) Input {

	Inputs.CurrentCycle = 0

	Inputs.Models = []Model{
		Model{0, "NAFLD"},
		Model{1, "CHD"},
		Model{2, "T2DM"},
		Model{3, "BMI"},
		Model{4, "Ethnicity"},
		Model{5, "Sex"},
		Model{6, "Physicial activity"},
		Model{7, "Fructose"},
		Model{8, "Trans Fat"},
		Model{9, "N3-PUFA"},
		Model{10, "Age"}}

	Inputs.States = []State{
		State{0, 0, "Unitialized1", true},
		State{1, 0, "Unitialized2", false},
		State{2, 0, "No NAFLD", false},
		State{3, 0, "Steatosis", false},
		State{4, 0, "NASH", false},
		State{5, 0, "Cirrhosis", false},
		State{6, 0, "HCC", false},
		State{7, 0, "Liver death", false},
		State{8, 0, "Natural death", false},
		State{9, 0, "Other death", false},
		State{10, 1, "Unitialized1", true},
		State{11, 1, "Unitialized2", false},
		State{12, 1, "No CHD", false},
		State{13, 1, "CHD", false},
		State{14, 1, "CHD Death", false},
		State{15, 1, "Other death", false},
		State{16, 2, "Unitialized1", true},
		State{17, 2, "Unitialized2", false},
		State{18, 2, "No T2DM", false},
		State{19, 2, "T2DM", false},
		State{20, 2, "T2DM Death", false},
		State{21, 2, "Other death", false},
		State{22, 3, "Unitialized1", true},
		State{23, 3, "Unitialized2", false},
		State{24, 3, "Healthy weight", false},
		State{25, 3, "Overweight", false},
		State{26, 3, "Obese", false},
		State{27, 3, "Other death", false},
		State{28, 4, "Unitialized1", true},
		State{29, 4, "Non-hispanic white", false},
		State{30, 4, "Non-hispanic black", false},
		State{31, 4, "Hispanic", false},
		State{32, 4, "Other death", false},
		State{33, 5, "Unitialized1", true},
		State{34, 5, "Male", false},
		State{35, 5, "Female", false},
		State{36, 5, "Other death", false},
		State{37, 6, "Unitialized1", true},
		State{38, 6, "Unitialized2", false},
		State{39, 6, "Active", false},
		State{40, 6, "Inactive", false},
		State{41, 6, "Other death", false},
		State{42, 7, "Unitialized1", true},
		State{43, 7, "Increase risk", false},
		State{44, 7, "No increased risk", false},
		State{45, 7, "Other death", false},
		State{46, 8, "Unitialized1", true},
		State{47, 8, "Increase risk", false},
		State{48, 8, "No increased risk", false},
		State{49, 8, "Other death", false},
		State{50, 9, "Unitialized1", true},
		State{51, 9, "Decreased risk", false},
		State{52, 9, "No decreased risk", false},
		State{53, 9, "Other death", false},
		State{54, 10, "Unitialized1", true},
		State{55, 10, "Age of 20", false},
		State{56, 10, "Age of 21", false},
		State{57, 10, "Age of 22", false},
		State{58, 10, "Age of 23", false},
		State{59, 10, "Age of 24", false},
		State{60, 10, "Age of 25", false},
		State{61, 10, "Age of 26", false},
		State{62, 10, "Age of 27", false},
		State{63, 10, "Age of 28", false},
		State{64, 10, "Age of 29", false},
		State{65, 10, "Age of 30", false},
		State{66, 10, "Age of 31", false},
		State{67, 10, "Age of 32", false},
		State{68, 10, "Age of 33", false},
		State{69, 10, "Age of 34", false},
		State{70, 10, "Age of 35", false},
		State{71, 10, "Age of 36", false},
		State{72, 10, "Age of 37", false},
		State{73, 10, "Age of 38", false},
		State{74, 10, "Age of 39", false},
		State{75, 10, "Age of 40", false},
		State{76, 10, "Age of 41", false},
		State{77, 10, "Age of 42", false},
		State{78, 10, "Age of 43", false},
		State{79, 10, "Age of 44", false},
		State{80, 10, "Age of 45", false},
		State{81, 10, "Age of 46", false},
		State{82, 10, "Age of 47", false},
		State{83, 10, "Age of 48", false},
		State{84, 10, "Age of 49", false},
		State{85, 10, "Age of 50", false},
		State{86, 10, "Age of 51", false},
		State{87, 10, "Age of 52", false},
		State{88, 10, "Age of 53", false},
		State{89, 10, "Age of 54", false},
		State{90, 10, "Age of 55", false},
		State{91, 10, "Age of 56", false},
		State{92, 10, "Age of 57", false},
		State{93, 10, "Age of 58", false},
		State{94, 10, "Age of 59", false},
		State{95, 10, "Age of 60", false},
		State{96, 10, "Age of 61", false},
		State{97, 10, "Age of 62", false},
		State{98, 10, "Age of 63", false},
		State{99, 10, "Age of 64", false},
		State{100, 10, "Age of 65", false},
		State{101, 10, "Age of 66", false},
		State{102, 10, "Age of 67", false},
		State{103, 10, "Age of 68", false},
		State{104, 10, "Age of 69", false},
		State{105, 10, "Age of 70", false},
		State{106, 10, "Age of 71", false},
		State{107, 10, "Age of 72", false},
		State{108, 10, "Age of 73", false},
		State{109, 10, "Age of 74", false},
		State{110, 10, "Age of 75", false},
		State{111, 10, "Age of 76", false},
		State{112, 10, "Age of 77", false},
		State{113, 10, "Age of 78", false},
		State{114, 10, "Age of 79", false},
		State{115, 10, "Age of 80", false},
		State{116, 10, "Age of 81", false},
		State{117, 10, "Age of 82", false},
		State{118, 10, "Age of 83", false},
		State{119, 10, "Age of 84", false},
		State{120, 10, "Age of 85", false},
		State{121, 10, "Age of 86", false},
		State{122, 10, "Age of 87", false},
		State{123, 10, "Age of 88", false},
		State{124, 10, "Age of 89", false},
		State{125, 10, "Age of 90", false},
		State{126, 10, "Age of 91", false},
		State{127, 10, "Age of 92", false},
		State{128, 10, "Age of 93", false},
		State{129, 10, "Age of 94", false},
		State{130, 10, "Age of 95", false},
		State{131, 10, "Age of 96", false},
		State{132, 10, "Age of 97", false},
		State{133, 10, "Age of 98", false},
		State{134, 10, "Age of 99", false},
		State{135, 10, "Age of 100", false},
		State{136, 10, "Age of 101", false},
		State{137, 10, "Age of 102", false},
		State{138, 10, "Age of 103", false},
		State{139, 10, "Age of 104", false},
		State{140, 10, "Age of 105", false},
		State{141, 10, "Age of 106", false},
		State{142, 10, "Age of 107", false},
		State{143, 10, "Age of 108", false},
		State{144, 10, "Age of 109", false},
		State{145, 10, "Age of 110", false}}

	// Id      int
	// From_id int
	// To_id   int
	// Tp_base float64

	Inputs.TransitionProbabilities = []TransitionProbability{
		TransitionProbability{0, 0, 0, 0},
		TransitionProbability{1, 0, 1, 1},
		TransitionProbability{2, 0, 2, 0},
		TransitionProbability{3, 0, 3, 0},
		TransitionProbability{4, 0, 4, 0},
		TransitionProbability{5, 0, 5, 0},
		TransitionProbability{6, 0, 6, 0},
		TransitionProbability{7, 0, 7, 0},
		TransitionProbability{8, 0, 8, 0},
		TransitionProbability{9, 0, 9, 0},
		TransitionProbability{10, 1, 0, 0},
		TransitionProbability{11, 1, 1, 0},
		TransitionProbability{12, 1, 2, 0.7},
		TransitionProbability{13, 1, 3, 0.234},
		TransitionProbability{14, 1, 4, 0.06},
		TransitionProbability{15, 1, 5, 0.006},
		TransitionProbability{16, 1, 6, 0},
		TransitionProbability{17, 1, 7, 0},
		TransitionProbability{18, 1, 8, 0},
		TransitionProbability{19, 1, 9, 0},
		TransitionProbability{20, 2, 0, 0},
		TransitionProbability{21, 2, 1, 0},
		TransitionProbability{22, 2, 2, 0.9797},
		TransitionProbability{23, 2, 3, 0.01},
		TransitionProbability{24, 2, 4, 0.0003},
		TransitionProbability{25, 2, 5, 0},
		TransitionProbability{26, 2, 6, 0},
		TransitionProbability{27, 2, 7, 0},
		TransitionProbability{28, 2, 8, 0.01},
		TransitionProbability{29, 2, 9, 0},
		TransitionProbability{30, 3, 0, 0},
		TransitionProbability{31, 3, 1, 0},
		TransitionProbability{32, 3, 2, 0.02},
		TransitionProbability{33, 3, 3, 0.9638},
		TransitionProbability{34, 3, 4, 0.006},
		TransitionProbability{35, 3, 5, 0.0002},
		TransitionProbability{36, 3, 6, 0},
		TransitionProbability{37, 3, 7, 0},
		TransitionProbability{38, 3, 8, 0.01},
		TransitionProbability{39, 3, 9, 0},
		TransitionProbability{40, 4, 0, 0},
		TransitionProbability{41, 4, 1, 0},
		TransitionProbability{42, 4, 2, 0.001},
		TransitionProbability{43, 4, 3, 0.02},
		TransitionProbability{44, 4, 4, 0.9669},
		TransitionProbability{45, 4, 5, 0.002},
		TransitionProbability{46, 4, 6, 0.0001},
		TransitionProbability{47, 4, 7, 0},
		TransitionProbability{48, 4, 8, 0.01},
		TransitionProbability{49, 4, 9, 0},
		TransitionProbability{50, 5, 0, 0},
		TransitionProbability{51, 5, 1, 0},
		TransitionProbability{52, 5, 2, 0},
		TransitionProbability{53, 5, 3, 0},
		TransitionProbability{54, 5, 4, 0},
		TransitionProbability{55, 5, 5, 0.97},
		TransitionProbability{56, 5, 6, 0.02},
		TransitionProbability{57, 5, 7, 0},
		TransitionProbability{58, 5, 8, 0.01},
		TransitionProbability{59, 5, 9, 0},
		TransitionProbability{60, 6, 0, 0},
		TransitionProbability{61, 6, 1, 0},
		TransitionProbability{62, 6, 2, 0},
		TransitionProbability{63, 6, 3, 0},
		TransitionProbability{64, 6, 4, 0},
		TransitionProbability{65, 6, 5, 0},
		TransitionProbability{66, 6, 6, 0.5},
		TransitionProbability{67, 6, 7, 0.5},
		TransitionProbability{68, 6, 8, 0},
		TransitionProbability{69, 6, 9, 0},
		TransitionProbability{70, 7, 0, 0},
		TransitionProbability{71, 7, 1, 0},
		TransitionProbability{72, 7, 2, 0},
		TransitionProbability{73, 7, 3, 0},
		TransitionProbability{74, 7, 4, 0},
		TransitionProbability{75, 7, 5, 0},
		TransitionProbability{76, 7, 6, 0},
		TransitionProbability{77, 7, 7, 1},
		TransitionProbability{78, 7, 8, 0},
		TransitionProbability{79, 7, 9, 0},
		TransitionProbability{80, 8, 0, 0},
		TransitionProbability{81, 8, 1, 0},
		TransitionProbability{82, 8, 2, 0},
		TransitionProbability{83, 8, 3, 0},
		TransitionProbability{84, 8, 4, 0},
		TransitionProbability{85, 8, 5, 0},
		TransitionProbability{86, 8, 6, 0},
		TransitionProbability{87, 8, 7, 0},
		TransitionProbability{88, 8, 8, 1},
		TransitionProbability{89, 8, 9, 0},
		TransitionProbability{90, 9, 0, 0},
		TransitionProbability{91, 9, 1, 0},
		TransitionProbability{92, 9, 2, 0},
		TransitionProbability{93, 9, 3, 0},
		TransitionProbability{94, 9, 4, 0},
		TransitionProbability{95, 9, 5, 0},
		TransitionProbability{96, 9, 6, 0},
		TransitionProbability{97, 9, 7, 0},
		TransitionProbability{98, 9, 8, 0},
		TransitionProbability{99, 9, 9, 1},
		TransitionProbability{100, 10, 10, 0},
		TransitionProbability{101, 10, 11, 1},
		TransitionProbability{102, 10, 12, 0},
		TransitionProbability{103, 10, 13, 0},
		TransitionProbability{104, 10, 14, 0},
		TransitionProbability{105, 10, 15, 0},
		TransitionProbability{106, 11, 10, 0},
		TransitionProbability{107, 11, 11, 0},
		TransitionProbability{108, 11, 12, 0.95},
		TransitionProbability{109, 11, 13, 0.05},
		TransitionProbability{110, 11, 14, 0},
		TransitionProbability{111, 11, 15, 0},
		TransitionProbability{112, 12, 10, 0},
		TransitionProbability{113, 12, 11, 0},
		TransitionProbability{114, 12, 12, 0.995},
		TransitionProbability{115, 12, 13, 0.005},
		TransitionProbability{116, 12, 14, 0},
		TransitionProbability{117, 12, 15, 0},
		TransitionProbability{118, 13, 10, 0},
		TransitionProbability{119, 13, 11, 0},
		TransitionProbability{120, 13, 12, 0},
		TransitionProbability{121, 13, 13, 0.99},
		TransitionProbability{122, 13, 14, 0.01},
		TransitionProbability{123, 13, 15, 0},
		TransitionProbability{124, 14, 10, 0},
		TransitionProbability{125, 14, 11, 0},
		TransitionProbability{126, 14, 12, 0},
		TransitionProbability{127, 14, 13, 0},
		TransitionProbability{128, 14, 14, 1},
		TransitionProbability{129, 14, 15, 0},
		TransitionProbability{130, 15, 10, 0},
		TransitionProbability{131, 15, 11, 0},
		TransitionProbability{132, 15, 12, 0},
		TransitionProbability{133, 15, 13, 0},
		TransitionProbability{134, 15, 14, 0},
		TransitionProbability{135, 15, 15, 1},
		TransitionProbability{136, 16, 16, 0},
		TransitionProbability{137, 16, 17, 1},
		TransitionProbability{138, 16, 18, 0},
		TransitionProbability{139, 16, 19, 0},
		TransitionProbability{140, 16, 20, 0},
		TransitionProbability{141, 16, 21, 0},
		TransitionProbability{142, 17, 16, 0},
		TransitionProbability{143, 17, 17, 0},
		TransitionProbability{144, 17, 18, 0.9},
		TransitionProbability{145, 17, 19, 0.1},
		TransitionProbability{146, 17, 20, 0},
		TransitionProbability{147, 17, 21, 0},
		TransitionProbability{148, 18, 16, 0},
		TransitionProbability{149, 18, 17, 0},
		TransitionProbability{150, 18, 18, 0.99},
		TransitionProbability{151, 18, 19, 0.01},
		TransitionProbability{152, 18, 20, 0},
		TransitionProbability{153, 18, 21, 0},
		TransitionProbability{154, 19, 16, 0},
		TransitionProbability{155, 19, 17, 0},
		TransitionProbability{156, 19, 18, 0},
		TransitionProbability{157, 19, 19, 0.99},
		TransitionProbability{158, 19, 20, 0.01},
		TransitionProbability{159, 19, 21, 0},
		TransitionProbability{160, 20, 16, 0},
		TransitionProbability{161, 20, 17, 0},
		TransitionProbability{162, 20, 18, 0},
		TransitionProbability{163, 20, 19, 0},
		TransitionProbability{164, 20, 20, 1},
		TransitionProbability{165, 20, 21, 0},
		TransitionProbability{166, 21, 16, 0},
		TransitionProbability{167, 21, 17, 0},
		TransitionProbability{168, 21, 18, 0},
		TransitionProbability{169, 21, 19, 0},
		TransitionProbability{170, 21, 20, 0},
		TransitionProbability{171, 21, 21, 1},
		TransitionProbability{172, 22, 22, 0},
		TransitionProbability{173, 22, 23, 1},
		TransitionProbability{174, 22, 24, 0},
		TransitionProbability{175, 22, 25, 0},
		TransitionProbability{176, 22, 26, 0},
		TransitionProbability{177, 22, 27, 0},
		TransitionProbability{178, 23, 22, 0},
		TransitionProbability{179, 23, 23, 0},
		TransitionProbability{180, 23, 24, 0.35},
		TransitionProbability{181, 23, 25, 0.35},
		TransitionProbability{182, 23, 26, 0.3},
		TransitionProbability{183, 23, 27, 0},
		TransitionProbability{184, 24, 22, 0},
		TransitionProbability{185, 24, 23, 0},
		TransitionProbability{186, 24, 24, 0.944},
		TransitionProbability{187, 24, 25, 0.05},
		TransitionProbability{188, 24, 26, 0.006},
		TransitionProbability{189, 24, 27, 0},
		TransitionProbability{190, 25, 22, 0},
		TransitionProbability{191, 25, 23, 0},
		TransitionProbability{192, 25, 24, 0.03},
		TransitionProbability{193, 25, 25, 0.946},
		TransitionProbability{194, 25, 26, 0.024},
		TransitionProbability{195, 25, 27, 0},
		TransitionProbability{196, 26, 22, 0},
		TransitionProbability{197, 26, 23, 0},
		TransitionProbability{198, 26, 24, 0.002},
		TransitionProbability{199, 26, 25, 0.02},
		TransitionProbability{200, 26, 26, 0.978},
		TransitionProbability{201, 26, 27, 0},
		TransitionProbability{202, 27, 22, 0},
		TransitionProbability{203, 27, 23, 0},
		TransitionProbability{204, 27, 24, 0},
		TransitionProbability{205, 27, 25, 0},
		TransitionProbability{206, 27, 26, 0},
		TransitionProbability{207, 27, 27, 1},
		TransitionProbability{208, 28, 28, 0},
		TransitionProbability{209, 28, 29, 0.743771},
		TransitionProbability{210, 28, 30, 0.115852},
		TransitionProbability{211, 28, 31, 0.140377},
		TransitionProbability{212, 28, 32, 0},
		TransitionProbability{213, 29, 28, 0},
		TransitionProbability{214, 29, 29, 1},
		TransitionProbability{215, 29, 30, 0},
		TransitionProbability{216, 29, 31, 0},
		TransitionProbability{217, 29, 32, 0},
		TransitionProbability{218, 30, 28, 0},
		TransitionProbability{219, 30, 29, 0},
		TransitionProbability{220, 30, 30, 1},
		TransitionProbability{221, 30, 31, 0},
		TransitionProbability{222, 30, 32, 0},
		TransitionProbability{223, 31, 28, 0},
		TransitionProbability{224, 31, 29, 0},
		TransitionProbability{225, 31, 30, 0},
		TransitionProbability{226, 31, 31, 1},
		TransitionProbability{227, 31, 32, 0},
		TransitionProbability{228, 32, 28, 0},
		TransitionProbability{229, 32, 29, 0},
		TransitionProbability{230, 32, 30, 0},
		TransitionProbability{231, 32, 31, 0},
		TransitionProbability{232, 32, 32, 1},
		TransitionProbability{233, 33, 33, 0},
		TransitionProbability{234, 33, 34, 0.484388},
		TransitionProbability{235, 33, 35, 0.515612},
		TransitionProbability{236, 33, 36, 0},
		TransitionProbability{237, 34, 33, 0},
		TransitionProbability{238, 34, 34, 1},
		TransitionProbability{239, 34, 35, 0},
		TransitionProbability{240, 34, 36, 0},
		TransitionProbability{241, 35, 33, 0},
		TransitionProbability{242, 35, 34, 0},
		TransitionProbability{243, 35, 35, 1},
		TransitionProbability{244, 35, 36, 0},
		TransitionProbability{245, 36, 33, 0},
		TransitionProbability{246, 36, 34, 0},
		TransitionProbability{247, 36, 35, 0},
		TransitionProbability{248, 36, 36, 1},
		TransitionProbability{249, 37, 37, 0},
		TransitionProbability{250, 37, 38, 1},
		TransitionProbability{251, 37, 39, 0},
		TransitionProbability{252, 37, 40, 0},
		TransitionProbability{253, 37, 41, 0},
		TransitionProbability{254, 38, 37, 0},
		TransitionProbability{255, 38, 38, 0},
		TransitionProbability{256, 38, 39, 0.3},
		TransitionProbability{257, 38, 40, 0.7},
		TransitionProbability{258, 38, 41, 0},
		TransitionProbability{259, 39, 37, 0},
		TransitionProbability{260, 39, 38, 0},
		TransitionProbability{261, 39, 39, 0.9},
		TransitionProbability{262, 39, 40, 0.1},
		TransitionProbability{263, 39, 41, 0},
		TransitionProbability{264, 40, 37, 0},
		TransitionProbability{265, 40, 38, 0},
		TransitionProbability{266, 40, 39, 0.1},
		TransitionProbability{267, 40, 40, 0.9},
		TransitionProbability{268, 40, 41, 0},
		TransitionProbability{269, 41, 37, 0},
		TransitionProbability{270, 41, 38, 0},
		TransitionProbability{271, 41, 39, 0},
		TransitionProbability{272, 41, 40, 0},
		TransitionProbability{273, 41, 41, 1},
		TransitionProbability{274, 42, 42, 0},
		TransitionProbability{275, 42, 43, 0.70004767},
		TransitionProbability{276, 42, 44, 0.29995233},
		TransitionProbability{277, 42, 45, 0},
		TransitionProbability{278, 43, 42, 0},
		TransitionProbability{279, 43, 43, 0.95},
		TransitionProbability{280, 43, 44, 0.05},
		TransitionProbability{281, 43, 45, 0},
		TransitionProbability{282, 44, 42, 0},
		TransitionProbability{283, 44, 43, 0.05},
		TransitionProbability{284, 44, 44, 0.95},
		TransitionProbability{285, 44, 45, 0},
		TransitionProbability{286, 45, 42, 0},
		TransitionProbability{287, 45, 43, 0},
		TransitionProbability{288, 45, 44, 0},
		TransitionProbability{289, 45, 45, 1},
		TransitionProbability{290, 46, 46, 0},
		TransitionProbability{291, 46, 47, 0.023},
		TransitionProbability{292, 46, 48, 0.977},
		TransitionProbability{293, 46, 49, 0},
		TransitionProbability{294, 47, 46, 0},
		TransitionProbability{295, 47, 47, 1},
		TransitionProbability{296, 47, 48, 0},
		TransitionProbability{297, 47, 49, 0},
		TransitionProbability{298, 48, 46, 0},
		TransitionProbability{299, 48, 47, 0},
		TransitionProbability{300, 48, 48, 1},
		TransitionProbability{301, 48, 49, 0},
		TransitionProbability{302, 49, 46, 0},
		TransitionProbability{303, 49, 47, 0},
		TransitionProbability{304, 49, 48, 0},
		TransitionProbability{305, 49, 49, 1},
		TransitionProbability{306, 50, 50, 0},
		TransitionProbability{307, 50, 51, 0.6306},
		TransitionProbability{308, 50, 52, 0.3694},
		TransitionProbability{309, 50, 53, 0},
		TransitionProbability{310, 51, 50, 0},
		TransitionProbability{311, 51, 51, 1},
		TransitionProbability{312, 51, 52, 0},
		TransitionProbability{313, 51, 53, 0},
		TransitionProbability{314, 52, 50, 0},
		TransitionProbability{315, 52, 51, 0},
		TransitionProbability{316, 52, 52, 1},
		TransitionProbability{317, 52, 53, 0},
		TransitionProbability{318, 53, 50, 0},
		TransitionProbability{319, 53, 51, 0},
		TransitionProbability{320, 53, 52, 0},
		TransitionProbability{321, 53, 53, 1},
		TransitionProbability{322, 54, 55, 0.019194468},
		TransitionProbability{323, 54, 56, 0.019194468},
		TransitionProbability{324, 54, 57, 0.019194468},
		TransitionProbability{325, 54, 58, 0.019194468},
		TransitionProbability{326, 54, 59, 0.019194468},
		TransitionProbability{327, 54, 60, 0.018700907},
		TransitionProbability{328, 54, 61, 0.018700907},
		TransitionProbability{329, 54, 62, 0.018700907},
		TransitionProbability{330, 54, 63, 0.018700907},
		TransitionProbability{331, 54, 64, 0.018700907},
		TransitionProbability{332, 54, 65, 0.0177493},
		TransitionProbability{333, 54, 66, 0.0177493},
		TransitionProbability{334, 54, 67, 0.0177493},
		TransitionProbability{335, 54, 68, 0.0177493},
		TransitionProbability{336, 54, 69, 0.0177493},
		TransitionProbability{337, 54, 70, 0.017756917},
		TransitionProbability{338, 54, 71, 0.017756917},
		TransitionProbability{339, 54, 72, 0.017756917},
		TransitionProbability{340, 54, 73, 0.017756917},
		TransitionProbability{341, 54, 74, 0.017756917},
		TransitionProbability{342, 54, 75, 0.018486967},
		TransitionProbability{343, 54, 76, 0.018486967},
		TransitionProbability{344, 54, 77, 0.018486967},
		TransitionProbability{345, 54, 78, 0.018486967},
		TransitionProbability{346, 54, 79, 0.018486967},
		TransitionProbability{347, 54, 80, 0.020017816},
		TransitionProbability{348, 54, 81, 0.020017816},
		TransitionProbability{349, 54, 82, 0.020017816},
		TransitionProbability{350, 54, 83, 0.020017816},
		TransitionProbability{351, 54, 84, 0.020017816},
		TransitionProbability{352, 54, 85, 0.019766711},
		TransitionProbability{353, 54, 86, 0.019766711},
		TransitionProbability{354, 54, 87, 0.019766711},
		TransitionProbability{355, 54, 88, 0.019766711},
		TransitionProbability{356, 54, 89, 0.019766711},
		TransitionProbability{357, 54, 90, 0.017505182},
		TransitionProbability{358, 54, 91, 0.017505182},
		TransitionProbability{359, 54, 92, 0.017505182},
		TransitionProbability{360, 54, 93, 0.017505182},
		TransitionProbability{361, 54, 94, 0.017505182},
		TransitionProbability{362, 54, 95, 0.015024017},
		TransitionProbability{363, 54, 96, 0.015024017},
		TransitionProbability{364, 54, 97, 0.015024017},
		TransitionProbability{365, 54, 98, 0.015024017},
		TransitionProbability{366, 54, 99, 0.015024017},
		TransitionProbability{367, 54, 100, 0.011072776},
		TransitionProbability{368, 54, 101, 0.011072776},
		TransitionProbability{369, 54, 102, 0.011072776},
		TransitionProbability{370, 54, 103, 0.011072776},
		TransitionProbability{371, 54, 104, 0.011072776},
		TransitionProbability{372, 54, 105, 0.008256161},
		TransitionProbability{373, 54, 106, 0.008256161},
		TransitionProbability{374, 54, 107, 0.008256161},
		TransitionProbability{375, 54, 108, 0.008256161},
		TransitionProbability{376, 54, 109, 0.008256161},
		TransitionProbability{377, 54, 110, 0.006472549},
		TransitionProbability{378, 54, 111, 0.006472549},
		TransitionProbability{379, 54, 112, 0.006472549},
		TransitionProbability{380, 54, 113, 0.006472549},
		TransitionProbability{381, 54, 114, 0.006472549},
		TransitionProbability{382, 54, 115, 0.005092902},
		TransitionProbability{383, 54, 116, 0.005092902},
		TransitionProbability{384, 54, 117, 0.005092902},
		TransitionProbability{385, 54, 118, 0.005092902},
		TransitionProbability{386, 54, 119, 0.005092902},
		TransitionProbability{387, 54, 120, 0.024516635},
		TransitionProbability{388, 55, 56, 1},
		TransitionProbability{389, 56, 57, 1},
		TransitionProbability{390, 57, 58, 1},
		TransitionProbability{391, 58, 59, 1},
		TransitionProbability{392, 59, 60, 1},
		TransitionProbability{393, 60, 61, 1},
		TransitionProbability{394, 61, 62, 1},
		TransitionProbability{395, 62, 63, 1},
		TransitionProbability{396, 63, 64, 1},
		TransitionProbability{397, 64, 65, 1},
		TransitionProbability{398, 65, 66, 1},
		TransitionProbability{399, 66, 67, 1},
		TransitionProbability{400, 67, 68, 1},
		TransitionProbability{401, 68, 69, 1},
		TransitionProbability{402, 69, 70, 1},
		TransitionProbability{403, 70, 71, 1},
		TransitionProbability{404, 71, 72, 1},
		TransitionProbability{405, 72, 73, 1},
		TransitionProbability{406, 73, 74, 1},
		TransitionProbability{407, 74, 75, 1},
		TransitionProbability{408, 75, 76, 1},
		TransitionProbability{409, 76, 77, 1},
		TransitionProbability{410, 77, 78, 1},
		TransitionProbability{411, 78, 79, 1},
		TransitionProbability{412, 79, 80, 1},
		TransitionProbability{413, 80, 81, 1},
		TransitionProbability{414, 81, 82, 1},
		TransitionProbability{415, 82, 83, 1},
		TransitionProbability{416, 83, 84, 1},
		TransitionProbability{417, 84, 85, 1},
		TransitionProbability{418, 85, 86, 1},
		TransitionProbability{419, 86, 87, 1},
		TransitionProbability{420, 87, 88, 1},
		TransitionProbability{421, 88, 89, 1},
		TransitionProbability{422, 89, 90, 1},
		TransitionProbability{423, 90, 91, 1},
		TransitionProbability{424, 91, 92, 1},
		TransitionProbability{425, 92, 93, 1},
		TransitionProbability{426, 93, 94, 1},
		TransitionProbability{427, 94, 95, 1},
		TransitionProbability{428, 95, 96, 1},
		TransitionProbability{429, 96, 97, 1},
		TransitionProbability{430, 97, 98, 1},
		TransitionProbability{431, 98, 99, 1},
		TransitionProbability{432, 99, 100, 1},
		TransitionProbability{433, 100, 101, 1},
		TransitionProbability{434, 101, 102, 1},
		TransitionProbability{435, 102, 103, 1},
		TransitionProbability{436, 103, 104, 1},
		TransitionProbability{437, 104, 105, 1},
		TransitionProbability{438, 105, 106, 1},
		TransitionProbability{439, 106, 107, 1},
		TransitionProbability{440, 107, 108, 1},
		TransitionProbability{441, 108, 109, 1},
		TransitionProbability{442, 109, 110, 1},
		TransitionProbability{443, 110, 111, 1},
		TransitionProbability{444, 111, 112, 1},
		TransitionProbability{445, 112, 113, 1},
		TransitionProbability{446, 113, 114, 1},
		TransitionProbability{447, 114, 115, 1},
		TransitionProbability{448, 115, 116, 1},
		TransitionProbability{449, 116, 117, 1},
		TransitionProbability{450, 117, 118, 1},
		TransitionProbability{451, 118, 119, 1},
		TransitionProbability{452, 119, 120, 1},
		TransitionProbability{453, 120, 121, 1},
		TransitionProbability{454, 121, 122, 1},
		TransitionProbability{455, 122, 123, 1},
		TransitionProbability{456, 123, 124, 1},
		TransitionProbability{457, 124, 125, 1},
		TransitionProbability{458, 125, 126, 1},
		TransitionProbability{459, 126, 127, 1},
		TransitionProbability{460, 127, 128, 1},
		TransitionProbability{461, 128, 129, 1},
		TransitionProbability{462, 129, 130, 1},
		TransitionProbability{463, 130, 131, 1},
		TransitionProbability{464, 131, 132, 1},
		TransitionProbability{465, 132, 133, 1},
		TransitionProbability{466, 133, 134, 1},
		TransitionProbability{467, 134, 135, 1},
		TransitionProbability{468, 135, 136, 1},
		TransitionProbability{469, 136, 137, 1},
		TransitionProbability{470, 137, 138, 1},
		TransitionProbability{471, 138, 139, 1},
		TransitionProbability{472, 139, 140, 1},
		TransitionProbability{473, 140, 141, 1},
		TransitionProbability{474, 141, 142, 1},
		TransitionProbability{475, 142, 143, 1},
		TransitionProbability{476, 143, 144, 1},
		TransitionProbability{477, 144, 145, 1},
		TransitionProbability{478, 145, 145, 1}}

	Inputs.Interactions = []Interaction{
		Interaction{0, 25, 2, 3, 1.8, 0},
		Interaction{1, 26, 2, 3, 2.5, 0},
		Interaction{2, 30, 2, 3, 0.93, 0},
		Interaction{3, 31, 2, 3, 1.67, 0}}

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
