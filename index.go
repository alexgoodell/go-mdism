// see readme for todos

// style guide:
// lowercase singular refers to local single
// uppercase singular is type
// uppercase plural is global object

package main

import (
	// "encoding/json"
	"flag"
	"fmt"
	//"github.com/alexgoodell/go-mdism/modules/sugar"
	//"io"
	// 	"net/http"
	"encoding/csv"
	"github.com/davecheney/profile"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	// "runtime/pprof"
	"strconv"
	"time"
)

var beginTime = time.Now() //TODO: Test this [Issue: https://github.com/alexgoodell/go-mdism/issues/32]

type State struct {
	Id                        int
	Model_id                  int
	Name                      string
	Is_uninitialized_state    bool
	Is_uninitialized_2_state  bool
	Is_disease_specific_death bool
	Is_other_death            bool
	Is_natural_causes_death   bool
}

type Model struct {
	Id   int
	Name string
}

type LifeExpectancy struct {
	Id              int
	Age_state_id    int
	Sex_state_id    int
	Life_expectancy float64
}

type MasterRecord struct {
	Cycle_id               int
	Person_id              int
	State_id               int
	Model_id               int
	YLDs                   float64
	YLLs                   float64
	Costs                  float64
	Has_entered_simulation bool
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
	PSA_id            int
}

type TransitionProbability struct {
	Id      int
	From_id int
	To_id   int
	Tp_base float64
	PSA_id  int
}

type Cost struct {
	Id       int
	State_id int
	Costs    float64
	PSA_id   int
}

type DisabilityWeight struct {
	Id                int
	State_id          int
	Disability_weight float64
	PSA_id            int
}

type InteractionKey struct {
	In_state_id   int
	From_state_id int
}

type RASkey struct {
	Race_state_id int
	Age_state_id  int
	Sex_state_id  int
	Model_id      int
}

type Query_t struct {
	State_id_by_cycle_and_person_and_model         [][][]int
	States_ids_by_cycle_and_person                 [][]int
	Tps_id_by_from_state                           [][]int
	interaction_id_by_in_state_and_from_state      map[InteractionKey]int
	State_populations_by_cycle                     [][]int
	Model_id_by_state                              []int
	Other_death_state_by_model                     []int
	Cost_by_state_id                               []float64
	Disability_weight_by_state_id                  []float64
	Master_record_id_by_cycle_and_person_and_model [][][]int
	Life_expectancy_by_sex_and_age                 map[SexAge]float64
	TP_by_RAS                                      map[RASkey][]TPByRAS
	Unintialized_state_by_model                    []int
	Outputs_id_by_cycle_and_state                  [][]int

	// Unexported and used by the "getters"
	model_id_by_name map[string]int
	state_id_by_name map[string]int
}

type SexAge struct {
	Sex, Age int
}

type Input struct {
	//	CurrentCycle            int
	Models                  []Model
	People                  []Person
	States                  []State
	TransitionProbabilities []TransitionProbability
	Interactions            []Interaction
	Cycles                  []Cycle
	MasterRecords           []MasterRecord
	Costs                   []Cost
	DisabilityWeights       []DisabilityWeight
	LifeExpectancies        []LifeExpectancy
	TPByRASs                []TPByRAS
}

type TPByRAS struct {
	Id            int
	Model_id      int
	Model_name    string
	To_state_id   int
	To_state_name string
	Sex_state_id  int
	Race_state_id int
	Age_state_id  int
	Probability   float64
}

// ##################### Output structs ################ //

//this struct will replicate the data found
type StatePopulation struct {
	Id         int
	State_name string
	State_id   int
	Cycle_id   int
	Population int
	Model_id   int
}

type OutputByCycleState struct {
	Id         int
	YLLs       float64
	YLDs       float64
	DALYs      float64
	Costs      float64
	Cycle_id   int
	State_id   int
	Population int
	State_name string
}

type Output struct {
	OutputsByCycleStateFull []OutputByCycleState
	OutputsByCycleStatePsa  []OutputByCycleState
	OutputsByCycle          []OutputByCycle
}

type OutputByCycle struct {
	Cycle_id             int
	T2DM_diagnosis_event int
	T2DM_death_event     int
	CHD_diagnosis_event  int
	CHD_death_event      int
	HCC_diagnosis_event  int
	HCC_death_event      int
}

// these are all global variables, which is why they are Capitalized
// current refers to the current cycle, which is used to calculate the next cycle

var Query Query_t

var GlobalStatePopulations = []StatePopulation{}

var output_dir = "tmp"

// TODO: Capitalize global variables [Issue: https://github.com/alexgoodell/go-mdism/issues/46]
var numberOfPeople int
var numberOfPeopleStarting int
var numberOfIterations int
var numberOfPeopleEnteringPerYear int
var numberOfPeopleEntering int

var inputsPath string
var isProfile string
var reportingMode string

var Inputs Input
var Outputs Output

func main() {

	flag.IntVar(&numberOfPeopleStarting, "people", 22400, "number of people to run")
	flag.IntVar(&numberOfIterations, "iterations", 1, "number times to run")
	// TODO: index error if number of people entering is <15000 [Issue: https://github.com/alexgoodell/go-mdism/issues/33]
	flag.IntVar(&numberOfPeopleEnteringPerYear, "entering", 416, "number of people that will enter the run(s)")
	flag.StringVar(&inputsPath, "inputs", "example", "folder that stores input csvs")
	flag.StringVar(&isProfile, "profile", "false", "cpu, mem, or false")
	flag.StringVar(&reportingMode, "reporting_mode", "individual", "either individual or psa")
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

	initializeInputs(inputsPath)

	//set up Query
	Query.setUp()

	// create people will generate individuals and add their data to the master
	// records
	Inputs = createInitialPeople(Inputs)

	Inputs = initializeGlobalStatePopulations(Inputs)

	interventionIsOn := false

	// TODO fix this hack
	//Interaction 250 = unin to high fructose (gets lowered from 0.7 to 0.56 (=80%))

	interventionFactor := 0.80 // This is the factor (or %) that you want to lower the high fructose TP by.
	adjustedTpBase := interventionFactor * Inputs.TransitionProbabilities[250].Tp_base
	if interventionIsOn {
		Inputs.TransitionProbabilities[250].Tp_base = adjustedTpBase
		Inputs.TransitionProbabilities[251].Tp_base = 1.00 - adjustedTpBase
		// TODO: Re-implement intervention not using adjust_transitions because adjust_transitions now requires a person [Issue: https://github.com/alexgoodell/go-mdism/issues/24]
		// unitFructoseState := get_state_by_id(&Inputs, 37)
		// tPs := unitFructoseState.get_destination_probabilites(&Inputs)
		// newTps = adjust_transitions(&Inputs, tPs, interventionAsInteraction, cycle, person, false)
	}

	// for _, newTp := range newTps {
	// 	Inputs.TransitionProbabilities[newTp.Id] = newTp
	// }

	// table tests here

	concurrencyBy := "person-within-cycle"

	runModel(concurrencyBy)

}

func runModel(concurrencyBy string) {

	fmt.Println("Intialization complete, time elapsed:", fmt.Sprint(time.Since(beginTime)))
	beginTime = time.Now()

	//create pointer to a new local set of inputs for each independent thread
	// TODO: Not sure if we need to deepCopy local inputs. Two threads can work on [Issue: https://github.com/alexgoodell/go-mdism/issues/39]
	// one global object seperately as long as the aren't changing the same values
	// see http://play.golang.org/p/edEbU10Lq0

	generalChan := make(chan string)

	switch concurrencyBy {

	case "person":

		// TODO have people enter

		for _, person := range Inputs.People { // foreach cycle
			go runFullModelForOnePerson(person, generalChan)
		} // end foreach cycle

		for _, person := range Inputs.People {
			chanString := <-generalChan
			_ = person     // to avoid unused warning
			_ = chanString // to avoid unused warning
		}

	case "person-within-cycle":

		for _, cycle := range Inputs.Cycles { // foreach cycle

			// need to create new people before calculating the year
			// of they're unit states will be written over
			if cycle.Id > 0 {
				createNewPeople(cycle, numberOfPeopleEnteringPerYear) //=The number of created people per cycle
			}

			for _, person := range Inputs.People { // 	foreach person
				go runOneCycleForOnePerson(cycle, person, generalChan)
			}

			for _, person := range Inputs.People { // 	foreach person
				chanString := <-generalChan
				_ = person     // to avoid unused warning
				_ = cycle      // to avoid unused warning
				_ = chanString // to avoid unused warning
			}
		}
	} // end case

	removeUnborns()

	fmt.Println("Time elapsed, excluding data import and export:", fmt.Sprint(time.Since(beginTime)))

	formatOutputs()

	if reportingMode == "individual" {
		//toCsv(output_dir+"/master.csv", Inputs.MasterRecords[0], Inputs.MasterRecords)
		toCsv("output"+"/state_populations.csv", GlobalStatePopulations[0], GlobalStatePopulations)
		toCsv(output_dir+"/output_by_cycle_and_state_full.csv", Outputs.OutputsByCycleStateFull[0], Outputs.OutputsByCycleStateFull)
		toCsv(output_dir+"/output_by_cycle_and_state_psa.csv", Outputs.OutputsByCycleStatePsa[0], Outputs.OutputsByCycleStatePsa)
		toCsv(output_dir+"/output_by_cycle.csv", Outputs.OutputsByCycle[0], Outputs.OutputsByCycle)
	}

	//toCsv(output_dir+"/states.csv", Inputs.States[0], Inputs.States)

	fmt.Println("Time elapsed, including data export:", fmt.Sprint(time.Since(beginTime)))

}

func formatOutputs() {

	for _, masterRecord := range Inputs.MasterRecords {
		//TODO: Remove state populations set up, output by cycle state is replacing [Issue: https://github.com/alexgoodell/go-mdism/issues/47]
		Query.State_populations_by_cycle[masterRecord.Cycle_id][masterRecord.State_id] += 1

		var oldStateId int
		if masterRecord.Cycle_id > 0 {
			oldStateId = Query.State_id_by_cycle_and_person_and_model[masterRecord.Cycle_id-1][masterRecord.Person_id][masterRecord.Model_id]
		}

		currentStateId := masterRecord.State_id

		outputCSId := Query.Outputs_id_by_cycle_and_state[masterRecord.Cycle_id][masterRecord.State_id]
		outputCS := &Outputs.OutputsByCycleStateFull[outputCSId]
		if outputCS.Cycle_id != masterRecord.Cycle_id || outputCS.State_id != masterRecord.State_id {
			fmt.Println("problem formating ouput by state cycle")
			os.Exit(1)
		}
		outputCS.Costs += masterRecord.Costs
		outputCS.YLDs += masterRecord.YLDs
		outputCS.YLLs += masterRecord.YLLs
		outputCS.DALYs += masterRecord.YLDs + masterRecord.YLLs
		outputCS.Population += 1

		/// per cycle outputs

		outputByCycle := &Outputs.OutputsByCycle[masterRecord.Cycle_id]
		outputByCycle.Cycle_id = masterRecord.Cycle_id
		evMapper := make(map[string]*int)
		evMapper["T2DM"] = &outputByCycle.T2DM_diagnosis_event
		evMapper["T2DM death"] = &outputByCycle.T2DM_death_event
		evMapper["CHD"] = &outputByCycle.CHD_diagnosis_event
		evMapper["CHD death"] = &outputByCycle.CHD_death_event
		evMapper["HCC"] = &outputByCycle.HCC_diagnosis_event
		evMapper["HCC death"] = &outputByCycle.HCC_death_event

		eventStateNames := []string{"T2DM", "T2DM death", "CHD", "CHD death", "HCC"}
		for _, eventStateName := range eventStateNames {
			eventStateId := Query.getStateByName(eventStateName).Id

			oldState := Inputs.States[oldStateId]

			// Find if they just transfered to the state of interest
			if currentStateId == eventStateId && oldStateId != eventStateId && !oldState.Is_uninitialized_2_state {

				*evMapper[eventStateName] += 1
			}
			//HCC is handled differently, because there is no "HCC death" state. So,
			// we are looking for people who's last state was HCC and are now in the
			// "liver death" group
			hccStateId := Query.getStateByName("HCC").Id
			liverDeathStateId := Query.getStateByName("Liver death").Id
			if currentStateId == liverDeathStateId && oldStateId == hccStateId {
				*evMapper["HCC death"] += 1
			}
		}

	}

	for _, outputCS := range Outputs.OutputsByCycleStateFull {
		stateNamesForPSA := []string{
			"Steatosis",
			"NASH", "Cirrhosis",
			"HCC", "Liver death",
			"Natural death", "CHD",
			"CHD death", "T2DM", "T2DM death",
			"Overweight", "Obese"}

		for _, stateName := range stateNamesForPSA {
			stateId := Query.getStateByName(stateName).Id
			if stateId == outputCS.State_id {
				Outputs.OutputsByCycleStatePsa = append(Outputs.OutputsByCycleStatePsa, outputCS)
			}
		}
	}

	for s, statePopulation := range GlobalStatePopulations {
		GlobalStatePopulations[s].Population = Query.State_populations_by_cycle[statePopulation.Cycle_id][statePopulation.State_id]
	}

	// for PSA reporting

	// Steatosis_prev       int
	// NASH_prev            int
	// Cirrhosis_prev       int
	// HCC_prev             int
	// Liver_death_prev     int
	// Natural_death_prev   int
	// CHD_prev             int
	// CHD_death_prev       int
	// T2DM_prev            int
	// T2DM_death_prev      int
	// Overweight_prev      int
	// Obese_prev           int

}

func runCyclePersonModel(cycle Cycle, model Model, person Person) {

	// get the current state of the person in this model (should be
	// the uninitialized state for cycle 0)
	currentStateInThisModel := person.get_state_by_model(model, cycle)

	otherDeathState := getOtherDeathStateByModel(model)
	if Query.Master_record_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id] == otherDeathState.Id {
		return
	}

	// if currentStateInThisModel == otherDeathState {
	// 	fmt.Println(person.Id, " has died in ", model.Name)
	// }

	// get the transition probabilities from the given state
	transitionProbabilities := currentStateInThisModel.get_destination_probabilites()

	// get all states this person is in in current cycle
	states := person.get_states(cycle)

	// if current state is "unitialized 2", this means that the transition
	// probabilities rely on information about the person's sex, race, and
	// age. So a different set of transition probabilties must be used

	isCHDuninit := currentStateInThisModel.Is_uninitialized_2_state && model.Id == Query.getModelByName("CHD").Id
	isT2DMuninit := currentStateInThisModel.Is_uninitialized_2_state && model.Id == Query.getModelByName("T2DM").Id
	isBMIuninit := currentStateInThisModel.Is_uninitialized_2_state && model.Id == Query.getModelByName("BMI").Id

	if isCHDuninit || isT2DMuninit || isBMIuninit {
		transitionProbabilities = getTransitionProbByRAS(currentStateInThisModel, states, person, cycle)

	}

	check_sum(transitionProbabilities) // will throw error if sum isn't 1

	// get any interactions that will effect the transtion from
	// the persons current states based on all states that they are
	// in - it is a method of their current state in this model,
	// and accepts an array of all currents states they occupy
	interactions := currentStateInThisModel.get_relevant_interactions(states)

	if len(interactions) > 0 { // if there are interactions
		for _, interaction := range interactions { // foreach interaction
			// apply the interactions to the transition probabilities
			newTransitionProbabilities := adjust_transitions(transitionProbabilities, interaction, cycle, person)
			transitionProbabilities = newTransitionProbabilities
		} // end foreach interaction
	} // end if there are interactions

	check_sum(transitionProbabilities) // will throw error if sum isn't 1

	// using  final transition probabilities, assign new state to person
	new_state := pickState(transitionProbabilities)

	// ------ health metrics ---------

	//Cost calculations
	discountValue := math.Pow((1 / 1.03), float64(cycle.Id)) //OR: LocalInputsPointer.CurrentCycle ?

	if cycle.Id > 0 {

		costs := Query.Cost_by_state_id[new_state.Id] * discountValue
		mrId := Query.Master_record_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id]
		mr := &Inputs.MasterRecords[mrId]
		mr.Costs += costs

		// years of life lost from disability
		stateSpecificYLDs := Query.Disability_weight_by_state_id[new_state.Id] * discountValue
		if math.IsNaN(stateSpecificYLDs) {
			fmt.Println("problem w discount. discount, disyld, dw:")
			fmt.Println(discountValue) //stateSpecificYLDs, new_state.Disability_weight)
			os.Exit(1)
		}
		mr.YLDs += stateSpecificYLDs

		// mortality
		justDiedOfDiseaseSpecific := new_state.Is_disease_specific_death && !currentStateInThisModel.Is_disease_specific_death
		justDiedOfNaturalCauses := new_state.Is_natural_causes_death && !currentStateInThisModel.Is_natural_causes_death
		if justDiedOfDiseaseSpecific {
			//fmt.Println("Just died of ", model.Name)
			stateSpecificYLLs := getYLLFromDeath(person, cycle) * discountValue
			//fmt.Println("incurring ", stateSpecificYLLs, " YLLs ")
			mr.YLLs += stateSpecificYLLs
		}

		// Sync deaths with other models
		if justDiedOfDiseaseSpecific || justDiedOfNaturalCauses {

			//fmt.Println("death in ", person.Id, " at cycle ", cycle.Id, " bc ", model.Name)
			// Sync deaths. Put person in "other death"
			for _, sub_model := range Inputs.Models {

				//skip current model because should show disease-specific death
				if sub_model.Id != model.Id {

					otherDeathState := getOtherDeathStateByModel(sub_model)
					// fmt.Println("moving ", person.Id, " to state, ", otherDeathState)
					// add new records for all the deaths for this cycle and next
					// TODO add toQuery adds to the next cycle not the currrent cycle
					// make this more clear

					// Set that they have died "other death" in models that are not this one
					// For the current cycle
					// mrId := Query.Master_record_id_by_cycle_and_person_and_model[cycle.Id][person.Id][sub_model.Id]
					// mr := &Inputs.MasterRecords[mrId]
					// mr.State_id = otherDeathState.Id

					// Query.State_id_by_cycle_and_person_and_model[cycle.Id][person.Id][sub_model.Id] = otherDeathState.Id

					// For the next cycle - in case this model has already
					// passed and they were assigned a new state
					mrId = Query.Master_record_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][sub_model.Id]
					mr = &Inputs.MasterRecords[mrId]
					mr.State_id = otherDeathState.Id

					Query.State_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][sub_model.Id] = otherDeathState.Id
				}
			}

		}
	}
	// check to make sure they are not mis-assigned
	if new_state.Is_other_death && !currentStateInThisModel.Is_other_death {
		fmt.Println("Should not be assigned other death here")
		os.Exit(1)
	}

	if new_state.Id < 1 {
		fmt.Println("No new state!")
		os.Exit(1)
	}

	//Store in two places, the master record and ...
	mrId := Query.Master_record_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id]
	mr := &Inputs.MasterRecords[mrId]
	mr.State_id = new_state.Id
	mr.Has_entered_simulation = true

	/// ... the state holder
	Query.State_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id] = new_state.Id

	check_new_state_id := Query.State_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id]

	if check_new_state_id != new_state.Id {
		fmt.Println("Was not correctly assigned... bug")
		os.Exit(1)
	}

}

func getYLLFromDeath(person Person, cycle Cycle) float64 {
	agesModel := Query.getModelByName("Age")
	ageState := person.get_state_by_model(agesModel, cycle)
	sexModel := Query.getModelByName("Sex")
	sexState := person.get_state_by_model(sexModel, cycle)
	return Query.getLifeExpectancyBySexAge(sexState, ageState)
}

func getOtherDeathStateByModel(model Model) State {
	otherDeathStateId := Query.Other_death_state_by_model[model.Id]
	otherDeathState := get_state_by_id(otherDeathStateId)
	return otherDeathState
}

// This represents running the full model for one person
func runFullModelForOnePerson(person Person, generalChan chan string) {
	for _, cycle := range Inputs.Cycles {
		shuffled := shuffle(Inputs.Models)
		for _, model := range shuffled {
			runCyclePersonModel(cycle, model, person)
		}
	}
	generalChan <- "Done"
}

func runOneCycleForOnePerson(cycle Cycle, person Person, generalChan chan string) {
	shuffled := shuffle(Inputs.Models)
	for _, model := range shuffled { // foreach model
		// cannot be made concurrent, because if they die in one model
		runCyclePersonModel(cycle, model, person)
	}
	generalChan <- "Done"
}

func (Query *Query_t) getModelByName(name string) Model {
	modelId := Query.model_id_by_name[name]
	model := Inputs.Models[modelId]
	if model.Name != name {
		fmt.Println("problem getting model by name: ", name, " does not exist")
		os.Exit(1)
	}
	return model
}

func (Query *Query_t) getStateByName(name string) State {
	stateId := Query.state_id_by_name[name]
	state := Inputs.States[stateId]
	if state.Name != name {
		fmt.Println("problem getting state by name: ", name, " does not exist")
		os.Exit(1)
	}
	return state
}

func (Query *Query_t) getLifeExpectancyBySexAge(sex State, age State) float64 {
	//Use struct as map key
	key := SexAge{sex.Id, age.Id}
	le := Query.Life_expectancy_by_sex_and_age[key]
	return le
}

func (Query *Query_t) getInteractionId(inState State, fromState State) (int, bool) {
	//Use struct as map key
	var key InteractionKey
	isInteraction := false
	key.In_state_id = inState.Id
	key.From_state_id = fromState.Id
	interactionId := Query.interaction_id_by_in_state_and_from_state[key]
	interaction := &Inputs.Interactions[interactionId]
	if interaction.From_state_id == fromState.Id && interaction.In_state_id == inState.Id {
		isInteraction = true
	}
	return interaction.Id, isInteraction
}

func (Query *Query_t) getTpByRAS(raceState State, ageState State, sexState State, model Model) []TPByRAS {
	var key RASkey
	key.Age_state_id = ageState.Id
	key.Race_state_id = raceState.Id
	key.Sex_state_id = sexState.Id
	key.Model_id = model.Id
	RASs := Query.TP_by_RAS[key]

	// if ras.Model_id != model.Id || ras.Age+22 != ageState.Id || ras.Race_state_id != raceState.Id || ras.Sex_state_id != sexState.Id {
	// 	fmt.Println("cannot find by RAS")
	// 	os.Exit(1)
	// }
	return RASs
}

func (Query *Query_t) setUp() {

	numberOfCalculatedCycles := len(Inputs.Cycles) + 1

	Outputs.OutputsByCycle = make([]OutputByCycle, numberOfCalculatedCycles, numberOfCalculatedCycles)

	Query.Outputs_id_by_cycle_and_state = make([][]int, numberOfCalculatedCycles, numberOfCalculatedCycles)

	Outputs.OutputsByCycleStateFull = make([]OutputByCycleState, numberOfCalculatedCycles*len(Inputs.States), numberOfCalculatedCycles*len(Inputs.States))
	i := 0
	for c := 0; c < numberOfCalculatedCycles; c++ {
		Query.Outputs_id_by_cycle_and_state[c] = make([]int, len(Inputs.States), len(Inputs.States))
		for s, state := range Inputs.States {
			var outputCS OutputByCycleState
			outputCS.Id = i
			outputCS.State_id = s
			outputCS.Cycle_id = c
			outputCS.State_name = state.Name
			outputCS.Population = 0
			Outputs.OutputsByCycleStateFull[i] = outputCS

			Query.Outputs_id_by_cycle_and_state[c][s] = i
			i++
		}
	}

	Query.Unintialized_state_by_model = make([]int, len(Inputs.Models), len(Inputs.Models))
	for _, state := range Inputs.States {
		if state.Is_uninitialized_state == true {
			Query.Unintialized_state_by_model[state.Model_id] = state.Id
		}
	}

	Query.TP_by_RAS = make(map[RASkey][]TPByRAS)
	for _, ras := range Inputs.TPByRASs {
		var key RASkey
		key.Age_state_id = ras.Age_state_id
		key.Race_state_id = ras.Race_state_id
		key.Sex_state_id = ras.Sex_state_id
		key.Model_id = ras.Model_id
		Query.TP_by_RAS[key] = append(Query.TP_by_RAS[key], ras)
	}

	Query.Life_expectancy_by_sex_and_age = make(map[SexAge]float64)

	for _, lifeExpectancy := range Inputs.LifeExpectancies {
		key := SexAge{lifeExpectancy.Sex_state_id, lifeExpectancy.Age_state_id}
		Query.Life_expectancy_by_sex_and_age[key] = lifeExpectancy.Life_expectancy
	}

	Query.model_id_by_name = make(map[string]int)
	for _, model := range Inputs.Models {
		Query.model_id_by_name[model.Name] = model.Id
	}

	Query.state_id_by_name = make(map[string]int)
	for _, state := range Inputs.States {
		Query.state_id_by_name[state.Name] = state.Id
	}

	Query.State_id_by_cycle_and_person_and_model = make([][][]int, len(Inputs.Cycles)+1, len(Inputs.Cycles)+1)
	for i, _ := range Query.State_id_by_cycle_and_person_and_model {
		//People
		Query.State_id_by_cycle_and_person_and_model[i] = make([][]int, numberOfPeople, numberOfPeople)
		for p, _ := range Query.State_id_by_cycle_and_person_and_model[i] {
			Query.State_id_by_cycle_and_person_and_model[i][p] = make([]int, len(Inputs.Models), len(Inputs.Models))
		}
	}

	//Cycles
	//Query.States_ids_by_cycle_and_person = make([][]int, 1000000, 1000000)

	Query.Tps_id_by_from_state = make([][]int, len(Inputs.States), len(Inputs.States))
	for i, _ := range Query.Tps_id_by_from_state {
		var tPIdsToReturn []int
		for _, transitionProbability := range Inputs.TransitionProbabilities {
			if transitionProbability.From_id == i {
				tPIdsToReturn = append(tPIdsToReturn, transitionProbability.Id)
			}
		}
		Query.Tps_id_by_from_state[i] = tPIdsToReturn
	}

	Query.interaction_id_by_in_state_and_from_state = make(map[InteractionKey]int)
	for _, interaction := range Inputs.Interactions {
		var interactionKey InteractionKey
		interactionKey.From_state_id = interaction.From_state_id
		interactionKey.In_state_id = interaction.In_state_id
		Query.interaction_id_by_in_state_and_from_state[interactionKey] = interaction.Id
	}

	//fmt.Println(Query.interaction_id_by_in_state_and_from_state)

	Query.Model_id_by_state = make([]int, len(Inputs.States), len(Inputs.States))

	for _, state := range Inputs.States {
		Query.Model_id_by_state[state.Id] = state.Model_id
	}

	/* TODO  Fix the cycle system. We actually end up storing len(Cycles)+1 cycles,
	because we start on 0 and calculate the cycle ahead of us, so if we have
	up to cycle 19 in the inputs, we will calculate 0-19, as well as cycle 20 */

	Query.State_populations_by_cycle = make([][]int, numberOfCalculatedCycles, numberOfCalculatedCycles)
	for c := 0; c < numberOfCalculatedCycles; c++ {
		Query.State_populations_by_cycle[c] = make([]int, len(Inputs.States), len(Inputs.States))
	}

	// ############## Other death state by model id ##################

	Query.Other_death_state_by_model = make([]int, len(Inputs.Models), len(Inputs.Models))
	for _, model := range Inputs.Models {
		// find other death state by iteration
		otherDeathState := State{}
		for _, state := range Inputs.States {
			if state.Is_other_death && state.Model_id == model.Id {
				otherDeathState = state
			}
		}

		if !otherDeathState.Is_other_death {
			fmt.Println("Problem finding other death state for model", model.Id)
			os.Exit(1)
		}

		Query.Other_death_state_by_model[model.Id] = otherDeathState.Id

	}

	// ############## Costs by state id ##################

	// fill in structure of query data with blanks
	Query.Cost_by_state_id = make([]float64, len(Inputs.States), len(Inputs.States))
	for i := 0; i < len(Inputs.States); i++ {
		Query.Cost_by_state_id[i] = 0
	}

	// put input from costs.csv (stored in Inputs.Costs) to fill in query data
	for _, cost := range Inputs.Costs {
		Query.Cost_by_state_id[cost.State_id] = cost.Costs
	}

	// ############## Disability weights by state id ##################

	// fill in structure of query data with blanks
	Query.Disability_weight_by_state_id = make([]float64, len(Inputs.States), len(Inputs.States))
	for i := 0; i < len(Inputs.States); i++ {
		Query.Disability_weight_by_state_id[i] = 0
	}

	// put input from costs.csv (stored in Inputs.Costs) to fill in query data
	for _, dw := range Inputs.DisabilityWeights {
		Query.Disability_weight_by_state_id[dw.State_id] = dw.Disability_weight
	}

}

func initializeGlobalStatePopulations(Inputs Input) Input {
	/* See cycle to do above */
	numberOfCalculatedCycles := len(Inputs.Cycles) + 1
	GlobalStatePopulations = make([]StatePopulation, numberOfCalculatedCycles*len(Inputs.States))
	q := 0
	for c := 0; c < numberOfCalculatedCycles; c++ {
		for s, state := range Inputs.States {
			GlobalStatePopulations[q].State_name = state.Name
			GlobalStatePopulations[q].Cycle_id = c
			GlobalStatePopulations[q].Id = q
			GlobalStatePopulations[q].Population = 0
			GlobalStatePopulations[q].State_id = s
			GlobalStatePopulations[q].Model_id = Query.Model_id_by_state[s]
			q++
		}
	}
	return Inputs
}

func shuffle(models []Model) []Model {
	modelsCopy := make([]Model, len(models), len(models))
	//Println("og: ", models)
	copy(modelsCopy, models)
	N := len(modelsCopy)
	for i := 0; i < N; i++ {
		// choose index uniformly in [i, N-1]
		r := i + rand.Intn(N-i)
		modelsCopy[r], modelsCopy[i] = modelsCopy[i], modelsCopy[r]
	}
	//fmt.Println("shuffled: ", modelsCopy)
	return modelsCopy
}

// Since we are using an open cohort, we need to add people to the
// simulation every year - these represent people that are being
// born into the simulation
func createNewPeople(cycle Cycle, number int) {
	idForFirstPerson := len(Inputs.People) + 1
	for i := 0; i < number; i++ {
		newPerson := Person{idForFirstPerson + i}
		Inputs.People = append(Inputs.People, newPerson)
		for _, model := range Inputs.Models {
			// TODO Age system should be more systematic
			// Place person into correct age category
			uninitializedState := State{}
			uninitializedState = model.get_uninitialized_state()
			if model.Name == "Age" {
				// Start them at age 20
				// TODO Do entering individuals actually enter at 21yo?
				uninitializedState = Query.getStateByName("Unitialized2")
			}
			var mr MasterRecord
			mr.Cycle_id = cycle.Id
			mr.State_id = uninitializedState.Id
			mr.Model_id = model.Id
			mr.Person_id = newPerson.Id
			mr.Has_entered_simulation = true

			Query.State_id_by_cycle_and_person_and_model[mr.Cycle_id][mr.Person_id][mr.Model_id] = mr.State_id
			mrId := Query.Master_record_id_by_cycle_and_person_and_model[mr.Cycle_id][mr.Person_id][mr.Model_id]
			Inputs.MasterRecords[mrId] = mr
		}
	}
}

// create people will generate individuals and add their data to the master
// records
func createInitialPeople(Inputs Input) Input {
	for i := 0; i < numberOfPeopleStarting; i++ {
		Inputs.People = append(Inputs.People, Person{i})
	}

	for _, person := range Inputs.People {
		for _, model := range Inputs.Models {
			uninitializedState := model.get_uninitialized_state()
			var mr MasterRecord
			mr.Cycle_id = 0
			mr.State_id = uninitializedState.Id
			mr.Model_id = model.Id
			mr.Person_id = person.Id
			mr.Has_entered_simulation = true

			Query.State_id_by_cycle_and_person_and_model[mr.Cycle_id][mr.Person_id][mr.Model_id] = mr.State_id

			mrId := Query.Master_record_id_by_cycle_and_person_and_model[mr.Cycle_id][mr.Person_id][mr.Model_id]
			Inputs.MasterRecords[mrId] = mr
		}
	}

	return Inputs

}

func get_state_by_id(stateId int) State {

	theState := Inputs.States[stateId]

	if theState.Id == stateId {
		return theState
	}

	fmt.Println("Cannot find state by id ", stateId)
	os.Exit(1)
	return theState

}

// ------------------------------------------- methods

// --------------- transition probabilities

func adjust_transitions(theseTPs []TransitionProbability, interaction Interaction, cycle Cycle, person Person) []TransitionProbability {

	adjustmentFactor := interaction.Adjustment

	// TODO Implement hooks
	// this adjusts a few transition probabilities which have a projected change over time
	hasTimeEffect := interaction.To_state_id == 13 || interaction.To_state_id == 14 || interaction.To_state_id == 8
	if cycle.Id > 1 && hasTimeEffect {
		// these prepresent the remaining risk after N cycles. ie remaining risk
		// is equal to original risk * 0.985 ^ number of years from original risk

		ageModel := Query.getModelByName("Age")
		ageModelStateId := person.get_state_by_model(ageModel, cycle).Id
		actualAge := ageModelStateId - 22 //Fix this hack = hardcoded

		timeEffectByToState := make([]float64, 15, 15)
		timeEffectByToState[13] = 0.985 //CHD incidence
		timeEffectByToState[14] = 0.979 //CHD mortality
		if actualAge >= 20 && actualAge <= 30 {
			timeEffectByToState[8] = 1.000 //natural deaths
		} else if actualAge > 30 && actualAge <= 55 {
			timeEffectByToState[8] = 0.980
		} else if actualAge > 55 {
			timeEffectByToState[8] = 0.970
		} else {
			fmt.Println("Cannot determine regression rate of natural mortality ", timeEffectByToState[8])
			os.Exit(1)
		}
		adjustmentFactor = adjustmentFactor * math.Pow(timeEffectByToState[interaction.To_state_id], float64(cycle.Id-2))
	}

	for i, _ := range theseTPs {
		// & represents the address, so now tp is a pointer - needed because you want to change the
		// underlying value of the elements of theseTPs, not just a copy of them
		tp := &theseTPs[i]
		originalTpBase := tp.Tp_base
		if tp.From_id == interaction.From_state_id && tp.To_id == interaction.To_state_id {
			tp.Tp_base = tp.Tp_base * adjustmentFactor
			if tp.Tp_base == originalTpBase && adjustmentFactor != 1 && originalTpBase != 0 {
			}
		}
	}

	// now, we need to make sure everything adds to one. to do so, we find what
	// it currently sums to, and make a new adjustment factor. We can then
	// adjust every transition probability by that amount.
	sum := get_sum(theseTPs)
	remain := sum - 1.0

	var recursiveTp float64
	for i, _ := range theseTPs {
		tp := &theseTPs[i] // need pointer to get underlying value, see above
		//find "recursive" tp, ie chance of staying in same state
		if tp.From_id == tp.To_id {
			tp.Tp_base -= remain
			recursiveTp = tp.Tp_base
			if tp.Tp_base < 0 {
				fmt.Println("Error: Tp under 0. Interaction: ", interaction.Id)
				os.Exit(1)
			}
		}
	}

	//check to make sure that people don't stay unitialized
	// TODO: Check for uninitialized 2 [Issue: https://github.com/alexgoodell/go-mdism/issues/43]
	model := Inputs.Models[interaction.Effected_model_id]
	unitState := model.get_uninitialized_state()
	if theseTPs[0].From_id == unitState.Id && recursiveTp != 0 {
		fmt.Println("recursiveTp is not zero for initialization!")
		os.Exit(1)
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

// get the current state of the person in this model (should be the uninitialized state for cycle 0)
func (thisPerson *Person) get_state_by_model(thisModel Model, cycle Cycle) State {
	thisModelId := thisModel.Id
	var stateToReturn State
	var stateToReturnId int
	stateToReturnId = Query.State_id_by_cycle_and_person_and_model[cycle.Id][thisPerson.Id][thisModelId]
	stateToReturn = Inputs.States[stateToReturnId]
	if stateToReturn.Id == stateToReturnId {
		return stateToReturn
	}
	fmt.Println("Cannot find state via get_state_by_model, error 2")
	os.Exit(1)
	return stateToReturn
}

// get all states this person is in at the current cycle
func (thisPerson *Person) get_states(cycle Cycle) []State {
	thisPersonId := thisPerson.Id
	statesToReturnIds := Query.State_id_by_cycle_and_person_and_model[cycle.Id][thisPersonId]
	statesToReturn := make([]State, len(statesToReturnIds), len(statesToReturnIds))
	for i, statesToReturnId := range statesToReturnIds {
		if Inputs.States[statesToReturnId].Id == statesToReturnId {
			statesToReturn[i] = Inputs.States[statesToReturnId]
		} else {
			fmt.Println("cannot find states via get_states, cycle & person id =", cycle.Id, thisPersonId)
			fmt.Println("looking for id", statesToReturnId, "but found", Inputs.States[statesToReturnId].Id)
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
	stateId := Query.Unintialized_state_by_model[model.Id]
	state := Inputs.States[stateId]
	return state
}

//  --------------- state

// get the transition probabilities *from* the given state. It's called
// destination because we're finding the chances of moving to each destination
func (state *State) get_destination_probabilites() []TransitionProbability {
	var tPIdsToReturn []int
	tPIdsToReturn = Query.Tps_id_by_from_state[state.Id]
	tPsToReturn := make([]TransitionProbability, len(tPIdsToReturn), len(tPIdsToReturn))
	for i, id := range tPIdsToReturn {
		tPsToReturn[i] = Inputs.TransitionProbabilities[id]
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
func (fromState State) get_relevant_interactions(allStates []State) []Interaction {

	i := 0
	relevantInteractions := make([]Interaction, len(Inputs.Models), len(Inputs.Models))
	for _, inState := range allStates {
		interactionId, isInteraction := Query.getInteractionId(inState, fromState)
		if isInteraction {
			//fmt.Println(interactionId)
			interaction := Inputs.Interactions[interactionId]
			relevantInteractions[i] = interaction
			i++
		}
	}
	// :i is faster than append()
	return relevantInteractions[:i]

}

// Using  the final transition probabilities, pickState assigns a new state to
// a person. It is given many states and returns one.
func pickState(tPs []TransitionProbability) State {
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
	// TODO: Add error report here [Issue: https://github.com/alexgoodell/go-mdism/issues/4]
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

func initializeInputs(inputsPath string) {

	// ####################### Life Expectancy #######################

	// initialize inputs, needed for fromCsv function
	filename := "inputs/" + inputsPath + "/life-expectancies.csv"
	numberOfRecords := getNumberOfRecords(filename)
	Inputs.LifeExpectancies = make([]LifeExpectancy, numberOfRecords, numberOfRecords)
	var LEptrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		LEptrs = append(LEptrs, new(LifeExpectancy))
	}
	LEptrs = fromCsv(filename, Inputs.LifeExpectancies[0], LEptrs)
	for i, ptr := range LEptrs {
		Inputs.LifeExpectancies[i] = *ptr.(*LifeExpectancy)
	}
	fmt.Println("complete")

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
	fmt.Println("complete")

	// ####################### States #######################

	// initialize inputs, needed for fromCsv function
	filename = "inputs/" + inputsPath + "/states.csv"
	fmt.Println(filename)
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

	fmt.Println(Inputs.DisabilityWeights)

	// ####################### Master Records & Accessor

	length := numberOfPeople * (len(Inputs.Cycles) + 1) * len(Inputs.Models)
	Inputs.MasterRecords = make([]MasterRecord, length, length)

	i := 0
	Query.Master_record_id_by_cycle_and_person_and_model = make([][][]int, len(Inputs.Cycles)+1, len(Inputs.Cycles)+1)
	for c, _ := range Query.Master_record_id_by_cycle_and_person_and_model {
		//People
		Query.Master_record_id_by_cycle_and_person_and_model[c] = make([][]int, numberOfPeople, numberOfPeople)
		for p, _ := range Query.Master_record_id_by_cycle_and_person_and_model[c] {
			Query.Master_record_id_by_cycle_and_person_and_model[c][p] = make([]int, len(Inputs.Models), len(Inputs.Models))
			for m, _ := range Query.Master_record_id_by_cycle_and_person_and_model[c][p] {
				var masterRecord MasterRecord
				masterRecord.Cycle_id = c
				masterRecord.Person_id = p
				masterRecord.Model_id = m
				masterRecord.Has_entered_simulation = false
				Inputs.MasterRecords[i] = masterRecord
				Query.Master_record_id_by_cycle_and_person_and_model[c][p][m] = i
				i++
			}
		}
	}
}

func removeUnborns() {
	i := 0
	masterRecordsToReturn := make([]MasterRecord, len(Inputs.MasterRecords), len(Inputs.MasterRecords))
	for p, _ := range Inputs.MasterRecords {
		if Inputs.MasterRecords[p].Has_entered_simulation == true {
			masterRecordsToReturn[i] = Inputs.MasterRecords[p]
			i++
		}
	}
	Inputs.MasterRecords = masterRecordsToReturn[:i]
}

func getTransitionProbByRAS(currentStateInThisModel State, states []State, person Person, cycle Cycle) []TransitionProbability {

	var tpsToReturn []TransitionProbability

	model := Inputs.Models[currentStateInThisModel.Model_id]

	raceModel := Query.getModelByName("Ethnicity")
	raceState := person.get_state_by_model(raceModel, cycle)

	sexModel := Query.getModelByName("Sex")
	sexState := person.get_state_by_model(sexModel, cycle)

	ageModel := Query.getModelByName("Age")
	ageState := person.get_state_by_model(ageModel, cycle)

	if ageState.Name == "Unitialized2" {
		//Use the 20yo transition probabilities for the 19yos (ie uninit 2)
		ageState = Query.getStateByName("Age of 20")
	}

	RASs := Query.getTpByRAS(raceState, ageState, sexState, model)

	for _, ras := range RASs {
		var newTp TransitionProbability
		newTp.To_id = ras.To_state_id
		newTp.Tp_base = ras.Probability
		tpsToReturn = append(tpsToReturn, newTp)
	}

	if len(tpsToReturn) < 1 {
		fmt.Println("No TPs found with getTransitionProbByRAS for m r a s")
		fmt.Println(raceState, ageState, sexState, model)
		os.Exit(1)
	}

	return tpsToReturn
}
