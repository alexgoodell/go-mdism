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
	"sync"
	//"github.com/alexgoodell/go-mdism/modules/sugar"
	//"io"
	// 	"net/http"
	"encoding/csv"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"runtime"

	"github.com/cheggaaa/pb"
	"github.com/davecheney/profile"
	"github.com/mgutz/ansi"
	// "runtime/pprof"
	"hash/fnv"
	"strconv"
	"time"
)

var interventionId int
var randomController RandomController_t

func main() {

	show_greeting()

	flag.IntVar(&numberOfPeopleStarting, "people", 22400, "number of people to run")
	flag.IntVar(&numberOfIterations, "iterations", 1, "number times to run")
	// TODO: index error if number of people entering is <15000 [Issue: https://github.com/alexgoodell/go-mdism/issues/33]
	flag.IntVar(&numberOfPeopleEnteringPerYear, "entering", 416, "number of people that will enter the run(s)")
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

	// TODO: remove hardcoded cycles [Issue: https://github.com/alexgoodell/go-mdism/issues/40]
	numberOfPeopleEntering = numberOfPeopleEnteringPerYear * (26 + 1)
	numberOfPeople = numberOfPeopleEntering + numberOfPeopleStarting

	fmt.Println("and ", numberOfPeopleStarting, "initial individuals")
	fmt.Println("and ", numberOfPeopleEntering, "individuals entering")
	fmt.Println("and ", numberOfIterations, "iterations")
	fmt.Println("and ", inputsPath, " as inputs")

	initializeInputs(inputsPath)

	// create people will generate individuals and add their data to the master
	// records

	fmt.Println("Intialization complete, time elapsed:", fmt.Sprint(time.Since(beginTime)))
	concurrencyBy := "person-within-cycle"

	isRunIntervention := true
	reportingMode = "individual"
	randId := 0

	switch isRunIntervention {

	case true:

		for _, eachIntervention := range Inputs.Interventions {

			//set up Query
			Query.setUp()

			interventionId = eachIntervention.Id
			interventionInitiate(eachIntervention)

			//clear results from last run
			Inputs.MasterRecords = []MasterRecord{}
			initializeMasterRecords()

			//build people
			Inputs.People = []Person{}
			Inputs = createInitialPeople(Inputs)

			if eachIntervention.Id == 0 {
				randomController.initialize() // needs to be made after people are created
			}
			randomController.resetCounters()

			fmt.Println("Using this many people: ", len(Inputs.People))

			count = 0

			runModel(concurrencyBy, eachIntervention.Name, randId)

			fmt.Println("count is", count)

		}

	case false:

		//runModel(concurrencyBy, "Base case")

	}

	// table tests here

}

var count int

func runModel(concurrencyBy string, interventionName string, randId int) {

	var mutex = &sync.Mutex{}

	msg := "Running " + interventionName + " simulation..."
	msg = ansi.Color(msg, "red+bh")

	fmt.Println("")
	fmt.Println(msg)
	fmt.Println("")
	count := len(Inputs.Models) * (len(Inputs.Cycles))
	bar = pb.StartNew(count)
	bar.ShowCounters = false
	bar.ShowTimeLeft = false

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

			// Alex: please check this; I have added the regression of the baseline TP's of CHD incidence and mortality.
			// I did this by adjusting the initial baseline TP by the set factor for each concomitant cycle.
			//Moved them here, to be calculated per cycle, because in CyclePersonModel, they would get discounted multiple
			//times if there was more than 1 interaction.

			for _, person := range Inputs.People { // 	foreach person
				go runOneCycleForOnePerson(cycle, person, generalChan, mutex)
			}

			for _, person := range Inputs.People { // 	foreach person
				chanString := <-generalChan
				_ = person     // to avoid unused warning
				_ = cycle      // to avoid unused warning
				_ = chanString // to avoid unused warning
			}

			randId = randId + numberOfPeople*len(Inputs.Cycles)
		}
	} // end case

	fmt.Println("Used shuffle random this many times: ", randomController.getShuffleCounter(mutex))
	fmt.Println("Used CPM random this many times: ", randomController.getCPMCounter(mutex))

	removeUnborns()

	fmt.Println("")
	fmt.Println("Time elapsed, excluding data import and export:", fmt.Sprint(time.Since(beginTime)))

	formatOutputs()

	if reportingMode == "individual" {
		// toCsv(output_dir+"/master.csv", Inputs.MasterRecords[0], Inputs.MasterRecords)
		//toCsv("output"+"/state_populations.csv", GlobalStatePopulations[0], GlobalStatePopulations)

		filename := "/output_by_cycle_and_state_full_interv_" + strconv.Itoa(interventionId) + ".csv"
		toCsv(output_dir+filename, Outputs.OutputsByCycleStateFull[0], Outputs.OutputsByCycleStateFull)
	}

	if reportingMode == "psa" {
		toCsv(output_dir+"/output_by_cycle_and_state_psa.csv", Outputs.OutputsByCycleStatePsa[0], Outputs.OutputsByCycleStatePsa)

	}

	//toCsv(output_dir+"/output_by_cycle.csv", Outputs.OutputsByCycle[0], Outputs.OutputsByCycle)

	fmt.Println("Time elapsed, including data export:", fmt.Sprint(time.Since(beginTime)))

}

func initializeMasterRecords() {
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

	// for s, statePopulation := range GlobalStatePopulations {
	// 	GlobalStatePopulations[s].Population = Query.State_populations_by_cycle[statePopulation.Cycle_id][statePopulation.State_id]
	// }

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

func runCyclePersonModel(cycle Cycle, model Model, person Person, mutex *sync.Mutex) {

	random := randomController.nextCPM(mutex, cycle.Id, person.Id, model.Id)

	if person.Id == 1 {
		bar.Increment()
	}

	otherDeathState := getOtherDeathStateByModel(model)
	if Query.State_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id] == otherDeathState.Id {
		return
	}

	// get the current state of the person in this model (should be
	// the uninitialized state for cycle 0)
	currentStateInThisModel := person.get_state_by_model(model, cycle)

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
	isNAFLDuninit := currentStateInThisModel.Is_uninitialized_2_state && model.Id == Query.getModelByName("NAFLD").Id
	isFrucuninit := currentStateInThisModel.Is_uninitialized_2_state && model.Id == Query.getModelByName("Fructose").Id

	if isCHDuninit || isT2DMuninit || isBMIuninit || isNAFLDuninit || isFrucuninit {
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

	// Alex: please check this; I have added the regression of the baseline TP's of CHD incidence and mortality.
	// I did this by adjusting the initial baseline TP by the set factor for each concomitant cycle.
	//Moved them here, to be calculated per cycle, because in CyclePersonModel, they would get discounted multiple
	//times if there was more than 1 interaction.
	if cycle.Id > 2 {
		y := float64(cycle.Id - 2)

		for q := 0; q < len(transitionProbabilities); q++ {
			switch transitionProbabilities[q].Id {

			case 115:
				transitionProbabilities[q].Tp_base = Inputs.TransitionProbabilities[115].Tp_base * math.Pow(0.985, y)
			case 122:
				transitionProbabilities[q].Tp_base = Inputs.TransitionProbabilities[122].Tp_base * math.Pow(0.979, y)
			case 114:
				transitionProbabilities[q].Tp_base = 1 - Inputs.TransitionProbabilities[115].Tp_base*math.Pow(0.985, y)
			case 121:
				transitionProbabilities[q].Tp_base = 1 - Inputs.TransitionProbabilities[122].Tp_base*math.Pow(0.979, y)

			}
		}

	}

	check_sum(transitionProbabilities) // will throw error if sum isn't 1

	// if cycle.Id > 2 { // TODO: FIX THIS! [Issue: https://github.com/alexgoodell/go-mdism/issues/56]
	// 	Inputs.TransitionProbabilities[115].Tp_base = Inputs.TransitionProbabilities[115].Tp_base * 0.985
	// 	Inputs.TransitionProbabilities[122].Tp_base = Inputs.TransitionProbabilities[122].Tp_base * 0.979
	// 	Inputs.TransitionProbabilities[114].Tp_base = 1 - Inputs.TransitionProbabilities[115].Tp_base
	// 	Inputs.TransitionProbabilities[121].Tp_base = 1 - Inputs.TransitionProbabilities[122].Tp_base
	// }

	// using  final transition probabilities, assign new state to person
	new_state := pickState(transitionProbabilities, random)

	// ------ health metrics ---------

	//Cost calculations
	discountValue := math.Pow((1 / 1.03), float64(cycle.Id-6)) //OR: LocalInputsPointer.CurrentCycle ?

	if cycle.Id > 0 {

		costs := Query.Cost_by_state_id[new_state.Id] * discountValue
		mrId := Query.Master_record_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id]
		mr := &Inputs.MasterRecords[mrId]
		if cycle.Id > 5 {
			mr.Costs += costs
		}

		// years of life lost from disability

		stateSpecificYLDs := Query.Disability_weight_by_state_id[new_state.Id] * discountValue
		if math.IsNaN(stateSpecificYLDs) {
			fmt.Println("problem w discount. discount, disyld, dw:")
			fmt.Println(discountValue) //stateSpecificYLDs, new_state.Disability_weight)
			os.Exit(1)
		}
		if cycle.Id > 5 {
			mr.YLDs += stateSpecificYLDs
		}

		// mortality
		justDiedOfDiseaseSpecific := new_state.Is_disease_specific_death && !currentStateInThisModel.Is_disease_specific_death
		justDiedOfNaturalCauses := new_state.Is_natural_causes_death && !currentStateInThisModel.Is_natural_causes_death
		if justDiedOfDiseaseSpecific {
			//fmt.Println("Just died of ", model.Name)

			stateSpecificYLLs := getYLLFromDeath(person, cycle) * discountValue
			//fmt.Println("incurring ", stateSpecificYLLs, " YLLs ")
			if cycle.Id > 5 {
				mr.YLLs += stateSpecificYLLs
			}
		}

		// Sync deaths with other models
		if justDiedOfDiseaseSpecific || justDiedOfNaturalCauses {

			//fmt.Println("death in ", person.Id, " at cycle ", cycle.Id, " bc ", model.Name)
			// Sync deaths. Put person in "other death"
			for _, sub_model := range Inputs.Models {

				//skip current model because should show disease-specific death
				if sub_model.Id != model.Id {

					otherDeathState := getOtherDeathStateByModel(sub_model)
					//fmt.Println("moving ", person.Id, " to state, ", otherDeathState, " in model ", sub_model.Name)
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
					//once they've been assigned a death state, need to clear
					// their other information (in case another model filled in
					// the data for the state they would have done to. But they've
					// died, so we need to clear the YLLs, YLDs, and Costs
					mr.YLDs = 0
					mr.YLLs = 0
					mr.Costs = 0
					mr.Has_entered_simulation = true
					mr.State_name = "Other death"

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

// This represents running the full model for one person
func runFullModelForOnePerson(person Person, generalChan chan string) {
	// for _, cycle := range Inputs.Cycles {
	// 	shuffled := Inputs.Models //needs shuffle
	// 	for _, model := range shuffled {
	// 		//runCyclePersonModel(cycle, model, person)
	// 		_ = cycle
	// 		_ = person
	// 		_ = model
	// 	}
	// }
	// generalChan <- "Done"
}

func runOneCycleForOnePerson(cycle Cycle, person Person, generalChan chan string, mutex *sync.Mutex) {
	shuffled := shuffle(Inputs.Models, mutex, person.Id)
	for _, model := range shuffled { // foreach model
		// cannot be made concurrent, because if they die in one model
		runCyclePersonModel(cycle, model, person, mutex)
	}
	generalChan <- "Done"
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

	// TODO: Change to TPs by race [Issue: https://github.com/alexgoodell/go-mdism/issues/57]
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

	// TODO: Change name to interaction ids [Issue: https://github.com/alexgoodell/go-mdism/issues/51]
	Query.interaction_id_by_in_state_and_from_state = make(map[InteractionKey][]int)
	for _, interaction := range Inputs.Interactions {
		var interactionKey InteractionKey
		interactionKey.From_state_id = interaction.From_state_id
		interactionKey.In_state_id = interaction.In_state_id
		Query.interaction_id_by_in_state_and_from_state[interactionKey] = append(Query.interaction_id_by_in_state_and_from_state[interactionKey], interaction.Id)
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

/*func initializeGlobalStatePopulations(Inputs Input) Input {
/* See cycle to do above */
/*numberOfCalculatedCycles := len(Inputs.Cycles) + 1
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
*/

func random_int(max int, mutex *sync.Mutex, personId int) int {
	random := randomController.nextShuffle(mutex, personId)
	random = random * float64(max)
	return int(random)
}

func shuffle(models []Model, mutex *sync.Mutex, personId int) []Model {
	modelsCopy := make([]Model, len(models), len(models))
	//Println("og: ", models)
	copy(modelsCopy, models)
	N := len(modelsCopy)
	for i := 0; i < N; i++ {
		// choose index uniformly in [i, N-1]
		r := i + random_int(N-i, mutex, personId)
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
				uninitializedState = Query.getStateByName("Age of 19")
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

// ------------------------------------------- methods

// --------------- transition probabilities

func adjust_transitions(theseTPs []TransitionProbability, interaction Interaction, cycle Cycle, person Person) []TransitionProbability {

	adjustmentFactor := interaction.Adjustment
	/*if person.Id == 20000 && cycle.Id == 10 {
		fmt.Println(adjustmentFactor, interaction.To_state_id, interaction.In_state_id)
	}*/
	// TODO Implement hooks
	// this adjusts a few transition probabilities which have a projected change over time
	if cycle.Id > 2 && interaction.To_state_id == 8 {
		// these prepresent the remaining risk after N cycles. ie remaining risk
		// is equal to original risk * 0.985 ^ number of years from original risk

		ageModel := Query.getModelByName("Age")
		ageModelStateId := person.get_state_by_model(ageModel, cycle).Id
		actualAge := ageModelStateId - 24 //Fix this hack = hardcoded

		timeEffectByToState := make([]float64, 15, 15)
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
		//if person.Id == 20000 && cycle.Id == 10 {
		//	fmt.Println("After Regression", adjustmentFactor, interaction.To_state_id, interaction.In_state_id)
		//}
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

// Using  the final transition probabilities, pickState assigns a new state to
// a person. It is given many states and returns one.
func pickState(tPs []TransitionProbability, random float64) State {
	probs := make([]float64, len(tPs), len(tPs))
	for i, tP := range tPs {
		probs[i] = tP.Tp_base
	}

	chosenIndex := pick(probs, random)
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
func pick(probabilities []float64, random float64) int {
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

	if ageState.Name == "Age of 19" {
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

type RandomController_t struct {
	randomListShuffle    []float64     // this is a slice of random variables
	randomListCPM        [][][]float64 // this is a slice of random variables
	accessCounterCPM     int           //this is a counter for how many times a random number was generated
	accessCounterShuffle int
}

func (randomController *RandomController_t) initialize() {

	rand.Seed(1)

	randomController.randomListCPM = make([][][]float64, len(Inputs.Cycles), len(Inputs.Cycles))
	for c := range Inputs.Cycles {
		randomController.randomListCPM[c] = make([][]float64, numberOfPeople, numberOfPeople)
		for p := 0; p < numberOfPeople; p++ {
			randomController.randomListCPM[c][p] = make([]float64, len(Inputs.Models), len(Inputs.Models))
			for m := range Inputs.Models {
				randomController.randomListCPM[c][p][m] = rand.Float64()
			}
		}
	}

	randomController.randomListShuffle = make([]float64, numberOfPeople, numberOfPeople)
	for p := 0; p < numberOfPeople; p++ {
		randomController.randomListShuffle[p] = rand.Float64()
	}

	randomController.accessCounterCPM = 0
	randomController.accessCounterShuffle = 0
}

func (randomController *RandomController_t) resetCounters() {
	randomController.accessCounterCPM = 0
	randomController.accessCounterShuffle = 0
}

func (randomController *RandomController_t) nextShuffle(mutex *sync.Mutex, person_id int) float64 {

	mutex.Lock()
	randomController.accessCounterShuffle++
	mutex.Unlock()

	return randomController.randomListShuffle[person_id]
}

func (randomController *RandomController_t) nextCPM(mutex *sync.Mutex, cycleId int, personId int, modelId int) float64 {

	mutex.Lock()
	randomController.accessCounterCPM++
	mutex.Unlock()

	//fmt.Println(len(randomController.randomListCPM[0]))
	//fmt.Println(cycleId, personId, modelId, ":")
	//os.Exit(1)
	//fmt.Println(randomController.randomListCPM[cycleId][personId][modelId])

	return randomController.randomListCPM[cycleId][personId][modelId]
}

func (randomController *RandomController_t) getShuffleCounter(mutex *sync.Mutex) int {

	mutex.Lock()
	count := randomController.accessCounterShuffle
	mutex.Unlock()

	return count
}

func (randomController *RandomController_t) getCPMCounter(mutex *sync.Mutex) int {

	mutex.Lock()
	count := randomController.accessCounterCPM
	mutex.Unlock()

	return count
}

func hash(s string) int64 {
	h := fnv.New32a()
	h.Write([]byte(s))

	p := int64(h.Sum32())
	// fmt.Println(s, ":", p)
	// pause()
	return p
}
