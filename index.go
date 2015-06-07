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
	// "github.com/davecheney/profile"
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

	numberOfPeople := 6800

	fmt.Println("and ", numberOfPeople, "individuals")

	//set up queryData
	Inputs = setUpQueryData(Inputs, numberOfPeople)

	// create people will generate individuals and add their data to the master
	// records
	Inputs = createPeople(Inputs, numberOfPeople)

	// table tests here

	concurrencyBy := "person"

	iterationChan := make(chan string)

	for i := 0; i < 100; i++ {
		go runModel(Inputs, concurrencyBy, iterationChan)
	}

	for i := 0; i < 100; i++ {
		toPrint := <-iterationChan
		fmt.Println(toPrint)
	}

	fmt.Println("Time elapsed:", fmt.Sprint(time.Since(beginTime)))

}

func runModel(Inputs Input, concurrencyBy string, iterationChan chan string) {

	var localInputs Input
	localInputs = deepCopy(Inputs)

	switch concurrencyBy {

	case "person":

		masterRecordsToAdd := make(chan []MasterRecord)

		//create pointer to a new local set of inputs for each independent thread

		for _, person := range Inputs.People { // foreach cycle
			go runModelWithConcurrentPeople(localInputs, person, masterRecordsToAdd)
		} // end foreach cycle

		for _, person := range Inputs.People {
			mRtoAdd := <-masterRecordsToAdd
			localInputs.MasterRecords = append(localInputs.MasterRecords, mRtoAdd...)
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

	//outputs

	randomNumber := fmt.Sprintf("%v", rand.Float64()*100000.0)

	toCsv(output_dir+"/master"+randomNumber+".csv", localInputs.MasterRecords[0], localInputs.MasterRecords)
	toCsv(output_dir+"/states"+randomNumber+".csv", Inputs.States[0], Inputs.States)

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
	//Inputs.QueryData.Tps_id_by_from_state = make([]int, 1000000, 1000000)

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
		Model{0, "TB disease"},
		Model{1, "TB treatment"},
		Model{2, "TB resistance"},
		Model{3, "HIV disease"},
		Model{4, "HIV treatment"},
		Model{5, "HIV risk groups"},
		Model{6, "Setting"},
		Model{7, "Diabetes disease and treatment"}}

	Inputs.States = []State{

		State{0, 0, "Uninitialized	", true},
		State{1, 0, "Uninfected	", false},
		State{2, 0, "Fast latent	", false},
		State{3, 0, "Slow latent	", false},
		State{4, 0, "Non-infectious active	", false},
		State{5, 0, "Infectious active	", false},
		State{6, 0, "Death	", false},
		State{7, 1, "Uninitialized	", true},
		State{8, 1, "Uninfected	", false},
		State{9, 1, "Untreated	", false},
		State{10, 1, "Treated	", false},
		State{11, 1, "Death	", false},
		State{12, 2, "Uninitialized	", true},
		State{13, 2, "Uninfected	", false},
		State{14, 2, "Sensitive	", false},
		State{15, 2, "Isoniazid	", false},
		State{16, 2, "INH resistant	", false},
		State{17, 2, "RIF resistant	", false},
		State{18, 2, "MDR	", false},
		State{19, 2, "XDR	", false},
		State{20, 2, "Death	", false},
		State{21, 3, "Uninitialized	", true},
		State{22, 3, "Uninfected	", false},
		State{23, 3, "Acute	", false},
		State{24, 3, "Early	", false},
		State{25, 3, "Medium	", false},
		State{26, 3, "Late	", false},
		State{27, 3, "Advanced	", false},
		State{28, 3, "AIDS	", false},
		State{29, 3, "Death	", false},
		State{30, 4, "Uninitialized	", true},
		State{31, 4, "Uninfected	", false},
		State{32, 4, "Untreated	", false},
		State{33, 4, "Treated	", false},
		State{34, 4, "Death	", false},
		State{35, 5, "Uninitialized	", true},
		State{36, 5, "GenPop Male	", false},
		State{37, 5, "GenPop Female	", false},
		State{38, 5, "FSW	", false},
		State{39, 5, "IDU male	", false},
		State{40, 5, "IDU female	", false},
		State{41, 5, "MSM	", false},
		State{42, 5, "Death	", false},
		State{43, 6, "Uninitialized	", true},
		State{44, 6, "Urban	", false},
		State{45, 6, "Rural	", false},
		State{46, 6, "Death	", false},
		State{47, 7, "Uninitialized	", true},
		State{48, 7, "Pre-diabetes	", false},
		State{49, 7, "Uncomplicated diabetes, untreated	", false},
		State{50, 7, "Uncomplicated diabetes, treated	", false},
		State{51, 7, "Complicated diabetes, untreated (non-CVD)	", false},
		State{52, 7, "Complicated diabetes, untreated (CVD)	", false},
		State{53, 7, "Complicated diabetes, treated (non-CVD)	", false},
		State{54, 7, "Complicated diabetes, treated (CVD)	", false},
		State{55, 7, "Death	", false}}

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
		TransitionProbability{7, 1, 0, 0},
		TransitionProbability{8, 1, 1, 0},
		TransitionProbability{9, 1, 2, 0},
		TransitionProbability{10, 1, 3, 1},
		TransitionProbability{11, 1, 4, 0},
		TransitionProbability{12, 1, 5, 0},
		TransitionProbability{13, 1, 6, 0},
		TransitionProbability{14, 2, 0, 0},
		TransitionProbability{15, 2, 1, 0},
		TransitionProbability{16, 2, 2, 0},
		TransitionProbability{17, 2, 3, 0},
		TransitionProbability{18, 2, 4, 0},
		TransitionProbability{19, 2, 5, 0},
		TransitionProbability{20, 2, 6, 1},
		TransitionProbability{21, 3, 0, 0},
		TransitionProbability{22, 3, 1, 0},
		TransitionProbability{23, 3, 2, 0},
		TransitionProbability{24, 3, 3, 0},
		TransitionProbability{25, 3, 4, 0},
		TransitionProbability{26, 3, 5, 0},
		TransitionProbability{27, 3, 6, 1},
		TransitionProbability{28, 4, 0, 0},
		TransitionProbability{29, 4, 1, 0},
		TransitionProbability{30, 4, 2, 0},
		TransitionProbability{31, 4, 3, 1},
		TransitionProbability{32, 4, 4, 0},
		TransitionProbability{33, 4, 5, 0},
		TransitionProbability{34, 4, 6, 0},
		TransitionProbability{35, 5, 0, 0},
		TransitionProbability{36, 5, 1, 0},
		TransitionProbability{37, 5, 2, 0},
		TransitionProbability{38, 5, 3, 0},
		TransitionProbability{39, 5, 4, 1},
		TransitionProbability{40, 5, 5, 0},
		TransitionProbability{41, 5, 6, 0},
		TransitionProbability{42, 6, 0, 0},
		TransitionProbability{43, 6, 1, 0},
		TransitionProbability{44, 6, 2, 0},
		TransitionProbability{45, 6, 3, 1},
		TransitionProbability{46, 6, 4, 0},
		TransitionProbability{47, 6, 5, 0},
		TransitionProbability{48, 6, 6, 0},
		TransitionProbability{49, 7, 7, 0},
		TransitionProbability{50, 7, 8, 1},
		TransitionProbability{51, 7, 9, 0},
		TransitionProbability{52, 7, 10, 0},
		TransitionProbability{53, 7, 11, 0},
		TransitionProbability{54, 8, 7, 0},
		TransitionProbability{55, 8, 8, 0},
		TransitionProbability{56, 8, 9, 0},
		TransitionProbability{57, 8, 10, 0},
		TransitionProbability{58, 8, 11, 1},
		TransitionProbability{59, 9, 7, 0},
		TransitionProbability{60, 9, 8, 0},
		TransitionProbability{61, 9, 9, 0},
		TransitionProbability{62, 9, 10, 1},
		TransitionProbability{63, 9, 11, 0},
		TransitionProbability{64, 10, 7, 0},
		TransitionProbability{65, 10, 8, 0},
		TransitionProbability{66, 10, 9, 0},
		TransitionProbability{67, 10, 10, 0},
		TransitionProbability{68, 10, 11, 1},
		TransitionProbability{69, 11, 7, 0},
		TransitionProbability{70, 11, 8, 0},
		TransitionProbability{71, 11, 9, 0},
		TransitionProbability{72, 11, 10, 1},
		TransitionProbability{73, 11, 11, 0},
		TransitionProbability{74, 12, 12, 0},
		TransitionProbability{75, 12, 13, 1},
		TransitionProbability{76, 12, 14, 0},
		TransitionProbability{77, 12, 15, 0},
		TransitionProbability{78, 12, 16, 0},
		TransitionProbability{79, 12, 17, 0},
		TransitionProbability{80, 12, 18, 0},
		TransitionProbability{81, 12, 19, 0},
		TransitionProbability{82, 12, 20, 0},
		TransitionProbability{83, 13, 12, 0},
		TransitionProbability{84, 13, 13, 0},
		TransitionProbability{85, 13, 14, 0},
		TransitionProbability{86, 13, 15, 0},
		TransitionProbability{87, 13, 16, 0},
		TransitionProbability{88, 13, 17, 1},
		TransitionProbability{89, 13, 18, 0},
		TransitionProbability{90, 13, 19, 0},
		TransitionProbability{91, 13, 20, 0},
		TransitionProbability{92, 14, 12, 0},
		TransitionProbability{93, 14, 13, 0},
		TransitionProbability{94, 14, 14, 0},
		TransitionProbability{95, 14, 15, 0},
		TransitionProbability{96, 14, 16, 0},
		TransitionProbability{97, 14, 17, 1},
		TransitionProbability{98, 14, 18, 0},
		TransitionProbability{99, 14, 19, 0},
		TransitionProbability{100, 14, 20, 0},
		TransitionProbability{101, 15, 12, 0},
		TransitionProbability{102, 15, 13, 0},
		TransitionProbability{103, 15, 14, 0},
		TransitionProbability{104, 15, 15, 0},
		TransitionProbability{105, 15, 16, 0},
		TransitionProbability{106, 15, 17, 0},
		TransitionProbability{107, 15, 18, 0},
		TransitionProbability{108, 15, 19, 0},
		TransitionProbability{109, 15, 20, 1},
		TransitionProbability{110, 16, 12, 0},
		TransitionProbability{111, 16, 13, 0},
		TransitionProbability{112, 16, 14, 0},
		TransitionProbability{113, 16, 15, 1},
		TransitionProbability{114, 16, 16, 0},
		TransitionProbability{115, 16, 17, 0},
		TransitionProbability{116, 16, 18, 0},
		TransitionProbability{117, 16, 19, 0},
		TransitionProbability{118, 16, 20, 0},
		TransitionProbability{119, 17, 12, 0},
		TransitionProbability{120, 17, 13, 0},
		TransitionProbability{121, 17, 14, 0},
		TransitionProbability{122, 17, 15, 0},
		TransitionProbability{123, 17, 16, 0},
		TransitionProbability{124, 17, 17, 1},
		TransitionProbability{125, 17, 18, 0},
		TransitionProbability{126, 17, 19, 0},
		TransitionProbability{127, 17, 20, 0},
		TransitionProbability{128, 18, 12, 0},
		TransitionProbability{129, 18, 13, 0},
		TransitionProbability{130, 18, 14, 1},
		TransitionProbability{131, 18, 15, 0},
		TransitionProbability{132, 18, 16, 0},
		TransitionProbability{133, 18, 17, 0},
		TransitionProbability{134, 18, 18, 0},
		TransitionProbability{135, 18, 19, 0},
		TransitionProbability{136, 18, 20, 0},
		TransitionProbability{137, 19, 12, 0},
		TransitionProbability{138, 19, 13, 0},
		TransitionProbability{139, 19, 14, 0},
		TransitionProbability{140, 19, 15, 0},
		TransitionProbability{141, 19, 16, 0},
		TransitionProbability{142, 19, 17, 1},
		TransitionProbability{143, 19, 18, 0},
		TransitionProbability{144, 19, 19, 0},
		TransitionProbability{145, 19, 20, 0},
		TransitionProbability{146, 20, 12, 0},
		TransitionProbability{147, 20, 13, 0},
		TransitionProbability{148, 20, 14, 0},
		TransitionProbability{149, 20, 15, 1},
		TransitionProbability{150, 20, 16, 0},
		TransitionProbability{151, 20, 17, 0},
		TransitionProbability{152, 20, 18, 0},
		TransitionProbability{153, 20, 19, 0},
		TransitionProbability{154, 20, 20, 0},
		TransitionProbability{155, 21, 21, 0},
		TransitionProbability{156, 21, 22, 0},
		TransitionProbability{157, 21, 23, 0},
		TransitionProbability{158, 21, 24, 0},
		TransitionProbability{159, 21, 25, 0},
		TransitionProbability{160, 21, 26, 1},
		TransitionProbability{161, 21, 27, 0},
		TransitionProbability{162, 21, 28, 0},
		TransitionProbability{163, 21, 29, 0},
		TransitionProbability{164, 22, 21, 0},
		TransitionProbability{165, 22, 22, 0},
		TransitionProbability{166, 22, 23, 0},
		TransitionProbability{167, 22, 24, 0},
		TransitionProbability{168, 22, 25, 0},
		TransitionProbability{169, 22, 26, 0},
		TransitionProbability{170, 22, 27, 1},
		TransitionProbability{171, 22, 28, 0},
		TransitionProbability{172, 22, 29, 0},
		TransitionProbability{173, 23, 21, 0},
		TransitionProbability{174, 23, 22, 0},
		TransitionProbability{175, 23, 23, 0},
		TransitionProbability{176, 23, 24, 0},
		TransitionProbability{177, 23, 25, 1},
		TransitionProbability{178, 23, 26, 0},
		TransitionProbability{179, 23, 27, 0},
		TransitionProbability{180, 23, 28, 0},
		TransitionProbability{181, 23, 29, 0},
		TransitionProbability{182, 24, 21, 0},
		TransitionProbability{183, 24, 22, 1},
		TransitionProbability{184, 24, 23, 0},
		TransitionProbability{185, 24, 24, 0},
		TransitionProbability{186, 24, 25, 0},
		TransitionProbability{187, 24, 26, 0},
		TransitionProbability{188, 24, 27, 0},
		TransitionProbability{189, 24, 28, 0},
		TransitionProbability{190, 24, 29, 0},
		TransitionProbability{191, 25, 21, 0},
		TransitionProbability{192, 25, 22, 0},
		TransitionProbability{193, 25, 23, 0},
		TransitionProbability{194, 25, 24, 0},
		TransitionProbability{195, 25, 25, 0},
		TransitionProbability{196, 25, 26, 1},
		TransitionProbability{197, 25, 27, 0},
		TransitionProbability{198, 25, 28, 0},
		TransitionProbability{199, 25, 29, 0},
		TransitionProbability{200, 26, 21, 1},
		TransitionProbability{201, 26, 22, 0},
		TransitionProbability{202, 26, 23, 0},
		TransitionProbability{203, 26, 24, 0},
		TransitionProbability{204, 26, 25, 0},
		TransitionProbability{205, 26, 26, 0},
		TransitionProbability{206, 26, 27, 0},
		TransitionProbability{207, 26, 28, 0},
		TransitionProbability{208, 26, 29, 0},
		TransitionProbability{209, 27, 21, 0},
		TransitionProbability{210, 27, 22, 0},
		TransitionProbability{211, 27, 23, 0},
		TransitionProbability{212, 27, 24, 0},
		TransitionProbability{213, 27, 25, 1},
		TransitionProbability{214, 27, 26, 0},
		TransitionProbability{215, 27, 27, 0},
		TransitionProbability{216, 27, 28, 0},
		TransitionProbability{217, 27, 29, 0},
		TransitionProbability{218, 28, 21, 0},
		TransitionProbability{219, 28, 22, 0},
		TransitionProbability{220, 28, 23, 0},
		TransitionProbability{221, 28, 24, 0},
		TransitionProbability{222, 28, 25, 0},
		TransitionProbability{223, 28, 26, 1},
		TransitionProbability{224, 28, 27, 0},
		TransitionProbability{225, 28, 28, 0},
		TransitionProbability{226, 28, 29, 0},
		TransitionProbability{227, 29, 21, 0},
		TransitionProbability{228, 29, 22, 0},
		TransitionProbability{229, 29, 23, 1},
		TransitionProbability{230, 29, 24, 0},
		TransitionProbability{231, 29, 25, 0},
		TransitionProbability{232, 29, 26, 0},
		TransitionProbability{233, 29, 27, 0},
		TransitionProbability{234, 29, 28, 0},
		TransitionProbability{235, 29, 29, 0},
		TransitionProbability{236, 30, 30, 1},
		TransitionProbability{237, 30, 31, 0},
		TransitionProbability{238, 30, 32, 0},
		TransitionProbability{239, 30, 33, 0},
		TransitionProbability{240, 30, 34, 0},
		TransitionProbability{241, 31, 30, 0},
		TransitionProbability{242, 31, 31, 1},
		TransitionProbability{243, 31, 32, 0},
		TransitionProbability{244, 31, 33, 0},
		TransitionProbability{245, 31, 34, 0},
		TransitionProbability{246, 32, 30, 0},
		TransitionProbability{247, 32, 31, 0},
		TransitionProbability{248, 32, 32, 0},
		TransitionProbability{249, 32, 33, 1},
		TransitionProbability{250, 32, 34, 0},
		TransitionProbability{251, 33, 30, 0},
		TransitionProbability{252, 33, 31, 1},
		TransitionProbability{253, 33, 32, 0},
		TransitionProbability{254, 33, 33, 0},
		TransitionProbability{255, 33, 34, 0},
		TransitionProbability{256, 34, 30, 1},
		TransitionProbability{257, 34, 31, 0},
		TransitionProbability{258, 34, 32, 0},
		TransitionProbability{259, 34, 33, 0},
		TransitionProbability{260, 34, 34, 0},
		TransitionProbability{261, 35, 35, 0},
		TransitionProbability{262, 35, 36, 0},
		TransitionProbability{263, 35, 37, 0},
		TransitionProbability{264, 35, 38, 0},
		TransitionProbability{265, 35, 39, 0},
		TransitionProbability{266, 35, 40, 1},
		TransitionProbability{267, 35, 41, 0},
		TransitionProbability{268, 35, 42, 0},
		TransitionProbability{269, 36, 35, 0},
		TransitionProbability{270, 36, 36, 0},
		TransitionProbability{271, 36, 37, 1},
		TransitionProbability{272, 36, 38, 0},
		TransitionProbability{273, 36, 39, 0},
		TransitionProbability{274, 36, 40, 0},
		TransitionProbability{275, 36, 41, 0},
		TransitionProbability{276, 36, 42, 0},
		TransitionProbability{277, 37, 35, 0},
		TransitionProbability{278, 37, 36, 0},
		TransitionProbability{279, 37, 37, 0},
		TransitionProbability{280, 37, 38, 1},
		TransitionProbability{281, 37, 39, 0},
		TransitionProbability{282, 37, 40, 0},
		TransitionProbability{283, 37, 41, 0},
		TransitionProbability{284, 37, 42, 0},
		TransitionProbability{285, 38, 35, 0},
		TransitionProbability{286, 38, 36, 0},
		TransitionProbability{287, 38, 37, 0},
		TransitionProbability{288, 38, 38, 0},
		TransitionProbability{289, 38, 39, 1},
		TransitionProbability{290, 38, 40, 0},
		TransitionProbability{291, 38, 41, 0},
		TransitionProbability{292, 38, 42, 0},
		TransitionProbability{293, 39, 35, 0},
		TransitionProbability{294, 39, 36, 0},
		TransitionProbability{295, 39, 37, 0},
		TransitionProbability{296, 39, 38, 0},
		TransitionProbability{297, 39, 39, 0},
		TransitionProbability{298, 39, 40, 0},
		TransitionProbability{299, 39, 41, 0},
		TransitionProbability{300, 39, 42, 1},
		TransitionProbability{301, 40, 35, 0},
		TransitionProbability{302, 40, 36, 0},
		TransitionProbability{303, 40, 37, 1},
		TransitionProbability{304, 40, 38, 0},
		TransitionProbability{305, 40, 39, 0},
		TransitionProbability{306, 40, 40, 0},
		TransitionProbability{307, 40, 41, 0},
		TransitionProbability{308, 40, 42, 0},
		TransitionProbability{309, 41, 35, 0},
		TransitionProbability{310, 41, 36, 0},
		TransitionProbability{311, 41, 37, 0},
		TransitionProbability{312, 41, 38, 0},
		TransitionProbability{313, 41, 39, 0},
		TransitionProbability{314, 41, 40, 0},
		TransitionProbability{315, 41, 41, 0},
		TransitionProbability{316, 41, 42, 1},
		TransitionProbability{317, 42, 35, 0},
		TransitionProbability{318, 42, 36, 0},
		TransitionProbability{319, 42, 37, 0},
		TransitionProbability{320, 42, 38, 0},
		TransitionProbability{321, 42, 39, 0},
		TransitionProbability{322, 42, 40, 0},
		TransitionProbability{323, 42, 41, 1},
		TransitionProbability{324, 42, 42, 0},
		TransitionProbability{325, 43, 43, 0},
		TransitionProbability{326, 43, 44, 1},
		TransitionProbability{327, 43, 45, 0},
		TransitionProbability{328, 43, 46, 0},
		TransitionProbability{329, 44, 43, 0},
		TransitionProbability{330, 44, 44, 0},
		TransitionProbability{331, 44, 45, 0},
		TransitionProbability{332, 44, 46, 1},
		TransitionProbability{333, 45, 43, 0},
		TransitionProbability{334, 45, 44, 0},
		TransitionProbability{335, 45, 45, 1},
		TransitionProbability{336, 45, 46, 0},
		TransitionProbability{337, 46, 43, 0},
		TransitionProbability{338, 46, 44, 0},
		TransitionProbability{339, 46, 45, 1},
		TransitionProbability{340, 46, 46, 0},
		TransitionProbability{341, 47, 47, 0},
		TransitionProbability{342, 47, 48, 0},
		TransitionProbability{343, 47, 49, 0},
		TransitionProbability{344, 47, 50, 0},
		TransitionProbability{345, 47, 51, 1},
		TransitionProbability{346, 47, 52, 0},
		TransitionProbability{347, 47, 53, 0},
		TransitionProbability{348, 47, 54, 0},
		TransitionProbability{349, 47, 55, 0},
		TransitionProbability{350, 48, 47, 0},
		TransitionProbability{351, 48, 48, 0},
		TransitionProbability{352, 48, 49, 0},
		TransitionProbability{353, 48, 50, 0},
		TransitionProbability{354, 48, 51, 0},
		TransitionProbability{355, 48, 52, 0},
		TransitionProbability{356, 48, 53, 0},
		TransitionProbability{357, 48, 54, 0},
		TransitionProbability{358, 48, 55, 1},
		TransitionProbability{359, 49, 47, 1},
		TransitionProbability{360, 49, 48, 0},
		TransitionProbability{361, 49, 49, 0},
		TransitionProbability{362, 49, 50, 0},
		TransitionProbability{363, 49, 51, 0},
		TransitionProbability{364, 49, 52, 0},
		TransitionProbability{365, 49, 53, 0},
		TransitionProbability{366, 49, 54, 0},
		TransitionProbability{367, 49, 55, 0},
		TransitionProbability{368, 50, 47, 0},
		TransitionProbability{369, 50, 48, 0},
		TransitionProbability{370, 50, 49, 1},
		TransitionProbability{371, 50, 50, 0},
		TransitionProbability{372, 50, 51, 0},
		TransitionProbability{373, 50, 52, 0},
		TransitionProbability{374, 50, 53, 0},
		TransitionProbability{375, 50, 54, 0},
		TransitionProbability{376, 50, 55, 0},
		TransitionProbability{377, 51, 47, 0},
		TransitionProbability{378, 51, 48, 0},
		TransitionProbability{379, 51, 49, 0},
		TransitionProbability{380, 51, 50, 0},
		TransitionProbability{381, 51, 51, 1},
		TransitionProbability{382, 51, 52, 0},
		TransitionProbability{383, 51, 53, 0},
		TransitionProbability{384, 51, 54, 0},
		TransitionProbability{385, 51, 55, 0},
		TransitionProbability{386, 52, 47, 0},
		TransitionProbability{387, 52, 48, 0},
		TransitionProbability{388, 52, 49, 0},
		TransitionProbability{389, 52, 50, 0},
		TransitionProbability{390, 52, 51, 0},
		TransitionProbability{391, 52, 52, 0},
		TransitionProbability{392, 52, 53, 1},
		TransitionProbability{393, 52, 54, 0},
		TransitionProbability{394, 52, 55, 0},
		TransitionProbability{395, 53, 47, 0},
		TransitionProbability{396, 53, 48, 0},
		TransitionProbability{397, 53, 49, 0},
		TransitionProbability{398, 53, 50, 0},
		TransitionProbability{399, 53, 51, 0},
		TransitionProbability{400, 53, 52, 1},
		TransitionProbability{401, 53, 53, 0},
		TransitionProbability{402, 53, 54, 0},
		TransitionProbability{403, 53, 55, 0},
		TransitionProbability{404, 54, 47, 0},
		TransitionProbability{405, 54, 48, 0},
		TransitionProbability{406, 54, 49, 0},
		TransitionProbability{407, 54, 50, 0},
		TransitionProbability{408, 54, 51, 0},
		TransitionProbability{409, 54, 52, 0},
		TransitionProbability{410, 54, 53, 0},
		TransitionProbability{411, 54, 54, 1},
		TransitionProbability{412, 54, 55, 0},
		TransitionProbability{413, 55, 47, 0},
		TransitionProbability{414, 55, 48, 0},
		TransitionProbability{415, 55, 49, 0},
		TransitionProbability{416, 55, 50, 0},
		TransitionProbability{417, 55, 51, 0},
		TransitionProbability{418, 55, 52, 1},
		TransitionProbability{419, 55, 53, 0},
		TransitionProbability{420, 55, 54, 0},
		TransitionProbability{421, 55, 55, 0}}

	Inputs.Interactions = []Interaction{
		Interaction{0, 1, 29, 57, 2, 1},
		Interaction{1, 2, 30, 58, 2, 2},
		Interaction{2, 3, 31, 59, 2, 3},
		Interaction{3, 4, 32, 60, 2, 4},
		Interaction{4, 5, 33, 61, 2, 1},
		Interaction{5, 6, 34, 62, 2, 2},
		Interaction{6, 7, 35, 63, 2, 3},
		Interaction{7, 8, 36, 64, 2, 4},
		Interaction{8, 9, 37, 65, 2, 1},
		Interaction{9, 10, 38, 66, 2, 2},
		Interaction{10, 11, 39, 67, 2, 3},
		Interaction{11, 12, 40, 68, 2, 4},
		Interaction{12, 13, 41, 69, 2, 1},
		Interaction{13, 14, 42, 70, 2, 2},
		Interaction{14, 15, 43, 71, 2, 3},
		Interaction{15, 16, 44, 72, 2, 4},
		Interaction{16, 17, 45, 73, 2, 1},
		Interaction{17, 18, 46, 74, 2, 2},
		Interaction{18, 19, 47, 75, 2, 3},
		Interaction{19, 20, 48, 76, 2, 4},
		Interaction{20, 21, 49, 77, 2, 1},
		Interaction{21, 22, 50, 78, 2, 2},
		Interaction{22, 23, 51, 79, 2, 3},
		Interaction{23, 24, 52, 80, 2, 4},
		Interaction{24, 25, 53, 81, 2, 1},
		Interaction{25, 26, 54, 82, 2, 2},
		Interaction{26, 27, 55, 83, 2, 3},
		Interaction{27, 28, 56, 84, 2, 4}}

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
