// see readme for todos

package main

import (
	// "encoding/json"
	"flag"
	"fmt"
	//"github.com/alexgoodell/go-mdism/modules/sugar"
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
	Id                        int
	Model_id                  int
	Name                      string
	Is_uninitialized_state    bool
	Disability_weight         float64
	Cost_per_cycle            float64
	Is_disease_specific_death bool
	Is_other_death            bool
	Is_natural_causes_death   bool
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

//this struct will replicate the data found
type StatePopulation struct {
	Id         int
	State_id   int
	Cycle_id   int
	Population int
	Model_id   int
}

type Query struct {
	State_id_by_cycle_and_person_and_model [][][]int
	States_ids_by_cycle_and_person         [][]int
	Tps_id_by_from_state                   [][]int
	Interactions_id_by_in_state_and_model  [][]int
	State_populations_by_cycle             [][]int
	Model_id_by_state                      []int
	Other_death_state_by_model             []int
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

type TPByRAS struct {
	Id            int
	Model_id      int
	Model_name    string
	To_state_id   int
	To_state_name string
	Sex_state_id  int
	Race_state_id int
	Age           int
	Probability   float64
}

var GlobalTPsByRAS []TPByRAS

// these are all global variables, which is why they are Capitalized
// current refers to the current cycle, which is used to calculate the next cycle

var GlobalMasterRecords = []MasterRecord{}
var GlobalStatePopulations = []StatePopulation{}

var GlobalYLDsByState = make([]float64, 150, 150)
var GlobalYLLsByState = make([]float64, 150, 150)
var GlobalCostsByState = make([]float64, 150, 150)
var GlobalDALYs = 0.00

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

	// TODO
	// assume same amount of people will enter over 20 years as are currently
	// in model
	numberOfPeopleEntering := numberOfPeople * 10
	//set up queryData
	Inputs = setUpQueryData(Inputs, numberOfPeople, numberOfPeopleEntering)

	// create people will generate individuals and add their data to the master
	// records
	Inputs = createInitialPeople(Inputs, numberOfPeople)

	Inputs = initializeGlobalStatePopulations(Inputs)

	setUpGlobalMasterRecordsByIPCM(Inputs)

	interventionIsOn := false

	interventionAsInteraction := Interaction{}
	// changes % increase risk from 0.7 to 0.5
	interventionAsInteraction.Adjustment = 0.71
	interventionAsInteraction.From_state_id = 42
	interventionAsInteraction.To_state_id = 43

	var newTps []TransitionProbability
	// TODO fix this hack
	if interventionIsOn {
		unitFructoseState := get_state_by_id(&Inputs, 42)
		tPs := unitFructoseState.get_destination_probabilites(&Inputs)
		newTps = adjust_transitions(&Inputs, tPs, interventionAsInteraction)
	}

	for _, newTp := range newTps {
		Inputs.TransitionProbabilities[newTp.Id] = newTp
	}

	// table tests here

	concurrencyBy := "person-within-cycle"

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

	fmt.Println("Intialization complete, time elaspsed:", fmt.Sprint(time.Since(beginTime)))
	beginTime = time.Now()

	masterRecordsToAdd := make(chan []MasterRecord)

	//create pointer to a new local set of inputs for each independent thread
	var localInputs Input
	localInputs = deepCopy(Inputs)

	switch concurrencyBy {

	case "person":

		for _, person := range Inputs.People { // foreach cycle
			go runFullModelForOnePerson(localInputs, person, masterRecordsToAdd)
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

	case "person-within-cycle":

		for _, cycle := range Inputs.Cycles { // foreach cycle

			// need to create new people before calculating the year
			// of they're unit states will be written over
			createNewPeople(&localInputs, cycle, 10)

			for _, person := range localInputs.People { // 	foreach person
				go runOneCycleForOnePerson(&localInputs, cycle, person, masterRecordsToAdd)
			}

			for _, person := range localInputs.People { // 	foreach person
				mRstoAdd := <-masterRecordsToAdd
				//TODO very slow!
				GlobalMasterRecords = append(GlobalMasterRecords, mRstoAdd...)
				for _, mRtoAdd := range mRstoAdd {
					//GlobalMasterRecordsByIPCM[0][mRtoAdd.Person_id][mRtoAdd.Cycle_id][mRtoAdd.Model_id] = mRtoAdd.State_id
					_ = mRtoAdd
				}
				_ = person // to avoid unused warning
				_ = cycle  // to avoid unused warning
			}

			localInputs.CurrentCycle++
			//fmt.Println("total num people, a or d, in sim", len(localInputs.People))
			//createNewPeople(&Inputs, cycle, 100)
		}
		// for _, cycle := range Inputs.Cycles { // foreach cycle
		// 	for _, person := range Inputs.People { // 	foreach person
		// 		mRstoAdd := <-masterRecordsToAdd
		// 		GlobalMasterRecords = append(GlobalMasterRecords, mRstoAdd...)
		// 		_ = person // to avoid unused warning
		// 		_ = cycle  // to avoid unused warning
		// 	}
		// }
	} // end case

	GlobalDALYs += sumSlices(GlobalYLDsByState, GlobalYLLsByState)
	//after total YLD and YLL is calculated, add everything into total DALYs
	//(cannot specify this per disease, since YLL for all states in NAFLD is split into natural death and liver death)

	fmt.Println("Time elapsed, excluding data import and export:", fmt.Sprint(time.Since(beginTime)))

	for _, masterRecord := range GlobalMasterRecords {

		if masterRecord.Person_id > 100 {
			//fmt.Println("Bingo", masterRecord.Person_id)
		}
		//fmt.Println(masterRecord.Cycle_id, masterRecord.State_id, Inputs.QueryData.State_populations_by_cycle[masterRecord.Cycle_id][masterRecord.State_id])
		Inputs.QueryData.State_populations_by_cycle[masterRecord.Cycle_id][masterRecord.State_id] += 1
	}

	for s, statePopulation := range GlobalStatePopulations {
		//fmt.Println(Inputs.QueryData.State_populations_by_cycle[statePopulation.Cycle_id][statePopulation.State_id])
		GlobalStatePopulations[s].Population = Inputs.QueryData.State_populations_by_cycle[statePopulation.Cycle_id][statePopulation.State_id]

	}

	//outputs
	for i := 0; i < 150; i++ {
		if GlobalYLDsByState[i] != 0 {
			fmt.Println("Global YLDs: ", GlobalYLDsByState[i])
		}
	}
	for i := 0; i < 150; i++ {
		if GlobalYLLsByState[i] != 0 {
			fmt.Println("Global YLLs: ", GlobalYLLsByState[i])
		}
	}
	for i := 0; i < 150; i++ {
		if GlobalCostsByState[i] != 0 {
			fmt.Println("Global Costs: ", GlobalCostsByState[i])
		}
	}
	fmt.Println("GlobalDALYs: ", GlobalDALYs)

	/*fmt.Println("Global YLDs Steatosis: ", GlobalYLDsByState[3])
	fmt.Println("Global YLDs NASH: ", GlobalYLDsByState[4])
	fmt.Println("Global YLDs Cirrhosis: ", GlobalYLDsByState[5])
	fmt.Println("Global YLDs HCC: ", GlobalYLDsByState[6])
	fmt.Println("Global YLDs CHD: ", GlobalYLDsByState[13])
	fmt.Println("Global YLDs T2D: ", GlobalYLDsByState[19])
	fmt.Println("Global YLDs Overweight: ", GlobalYLDsByState[25])
	fmt.Println("Global YLDs Obesity: ", GlobalYLDsByState[26])
	fmt.Println("Global YLLs Liver Death: ", GlobalYLLsByState[7])
	fmt.Println("Global YLLs Natural Death: ", GlobalYLLsByState[8])
	fmt.Println("Global YLLs CHD: ", GlobalYLLsByState[14])
	fmt.Println("Global YLLs T2D: ", GlobalYLLsByState[20])
	fmt.Println("GlobalDALYs: ", GlobalDALYs)
	fmt.Println("Global costs Steatosis: ", GlobalCostsByState[3])
	fmt.Println("Global costs NASH: ", GlobalCostsByState[4])
	fmt.Println("Global costs Cirrhosis: ", GlobalCostsByState[5])
	fmt.Println("Global costs HCC: ", GlobalCostsByState[6])
	fmt.Println("Global costs CHD: ", GlobalCostsByState[13])
	fmt.Println("Global costs T2D: ", GlobalCostsByState[19])
	fmt.Println("Global costs Overweight: ", GlobalCostsByState[25])
	fmt.Println("Global costs Obesity: ", GlobalCostsByState[26])
	*/

	toCsv(output_dir+"/master.csv", GlobalMasterRecords[0], GlobalMasterRecords)
	toCsv("output"+"/state_populations.csv", GlobalStatePopulations[0], GlobalStatePopulations)

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

func runCyclePersonModel(localInputsPointer *Input, cycle Cycle, model Model, person Person, theseMasterRecordsToAddPtr *[]MasterRecord, mrIndex int) {

	// get the current state of the person in this model (should be
	// the uninitialized state for cycle 0)
	currentStateInThisModel := person.get_state_by_model(localInputsPointer, model)

	// get the transition probabilities from the given state
	transitionProbabilities := currentStateInThisModel.get_destination_probabilites(localInputsPointer)

	// get all states this person is in in current cycle
	states := person.get_states(localInputsPointer)

	// if current state is "unitialized 2", this means that the transition
	// probabilities rely on information about the person's sex, race, and
	// age. So a different set of transition probabilties must be used

	// TODO add in CHD
	if currentStateInThisModel.Id == 11 || currentStateInThisModel.Id == 17 || currentStateInThisModel.Id == 23 || currentStateInThisModel.Id == 38 {
		transitionProbabilities = getTransitionProbByRAS(localInputsPointer, currentStateInThisModel, states, person)
	}
	check_sum(transitionProbabilities) // will throw error if sum isn't 1

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

	// health metrics

	//Cost calculations
	discountValue := math.Pow(0.97, float64(cycle.Id)) //OR: LocalInputsPointer.CurrentCycle ?

	stateCosts := make([]float64, 147, 147)
	stateCosts[3] = 150.00
	stateCosts[4] = 262.00
	stateCosts[5] = 5330.00
	stateCosts[6] = 37951.00
	stateCosts[13] = 8000.00
	stateCosts[19] = 7888.00
	stateCosts[25] = 350.00
	stateCosts[26] = 852.00

	if cycle.Id > 1 {
		GlobalCostsByState[new_state.Id] += stateCosts[new_state.Id] * discountValue
	}

	// years of life lost from disability
	stateSpecificYLDs := make([]float64, 150, 150)

	if cycle.Id > 1 {
		stateSpecificYLDs[new_state.Id] = new_state.Disability_weight / (1 - discountValue) * (1 - math.Exp(-(1 - discountValue)))

		if math.IsNaN(stateSpecificYLDs[new_state.Id]) {
			fmt.Println("problem w discount. discount, disyld, dw:")
			fmt.Println(discountValue, stateSpecificYLDs[new_state.Id], new_state.Disability_weight)
			os.Exit(1)
		}

		GlobalYLDsByState[new_state.Id] += stateSpecificYLDs[new_state.Id]
	}

	//fmt.Println("model Id", model.Id)
	justDiedOfDiseaseSpecific := new_state.Is_disease_specific_death && !currentStateInThisModel.Is_disease_specific_death

	justDiedOfNaturalCauses := new_state.Is_natural_causes_death && !currentStateInThisModel.Is_natural_causes_death

	stateSpecificYLLs := make([]float64, 150, 150)

	if justDiedOfDiseaseSpecific || justDiedOfNaturalCauses {

		stateSpecificYLLs[new_state.Id] = 1 / (1 - discountValue) * (1 - math.Exp(-(1-discountValue)*getYLLFromDeath(localInputsPointer, person)))
		GlobalYLLsByState[new_state.Id] += stateSpecificYLLs[new_state.Id]

		//fmt.Println("death sync in model ", model.Id)
		// Sync deaths. Put person in "other death"
		for _, sub_model := range localInputsPointer.Models {
			//skip current model because should show disease-specific death
			if sub_model.Id != model.Id {

				otherDeathState := getOtherDeathStateByModel(localInputsPointer, sub_model)
				//fmt.Println("updated model ", sub_model.Id, " with otherdeathstate ", otherDeathState)
				// add new records for all the deaths for this cycle and next
				// TODO add toQueryData adds to the next cycle not the currrent cycle
				// make this more clear
				prev_cycle := Cycle{}
				prev_cycle.Id = cycle.Id - 1
				addToQueryDataMasterRecord(localInputsPointer, prev_cycle, person, otherDeathState)
				addToQueryDataMasterRecord(localInputsPointer, cycle, person, otherDeathState)
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

	if localInputsPointer.CurrentCycle != cycle.Id {
		fmt.Println("cycle mismatch!", localInputsPointer.CurrentCycle, cycle.Id)
		os.Exit(1)
	}

	// store new state in master object
	err := addToQueryDataMasterRecord(localInputsPointer, cycle, person, new_state)

	check_new_state_id := localInputsPointer.QueryData.State_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id]

	if check_new_state_id != new_state.Id {
		fmt.Println("Was not correctly assigned... bug")
		os.Exit(1)
	}

	if err != false {
		fmt.Println("problem adding master record")
		os.Exit(1)
	}
}

func sumSlices(x []float64, y []float64) float64 {

	totalx := 0.0
	for _, valuex := range x {
		totalx += valuex
	}

	totaly := 0.0
	for _, valuey := range y {
		totaly += valuey
	}

	return totalx + totaly
}

func getYLLFromDeath(localInputsPointer *Input, person Person) float64 {

	//TODO sloppy need to make imported table

	agesModel := localInputsPointer.Models[10]
	stateInAge := person.get_state_by_model(localInputsPointer, agesModel)
	//TODO fix this age hack - not sustainable, what happens is the state IDS change?
	age := stateInAge.Id - 35
	getLifeexpectancy := make([]float64, 111, 111)

	getLifeexpectancy[20] = 49.627
	getLifeexpectancy[21] = 49.627
	getLifeexpectancy[22] = 49.627
	getLifeexpectancy[23] = 49.627
	getLifeexpectancy[24] = 49.627
	getLifeexpectancy[25] = 45.331
	getLifeexpectancy[26] = 45.331
	getLifeexpectancy[27] = 45.331
	getLifeexpectancy[28] = 45.331
	getLifeexpectancy[29] = 45.331
	getLifeexpectancy[30] = 41.078
	getLifeexpectancy[31] = 41.078
	getLifeexpectancy[32] = 41.078
	getLifeexpectancy[33] = 41.078
	getLifeexpectancy[34] = 41.078
	getLifeexpectancy[35] = 36.848
	getLifeexpectancy[36] = 36.848
	getLifeexpectancy[37] = 36.848
	getLifeexpectancy[38] = 36.848
	getLifeexpectancy[39] = 36.848
	getLifeexpectancy[40] = 32.682
	getLifeexpectancy[41] = 32.682
	getLifeexpectancy[42] = 32.682
	getLifeexpectancy[43] = 32.682
	getLifeexpectancy[44] = 32.682
	getLifeexpectancy[45] = 28.638
	getLifeexpectancy[46] = 28.638
	getLifeexpectancy[47] = 28.638
	getLifeexpectancy[48] = 28.638
	getLifeexpectancy[49] = 28.638
	getLifeexpectancy[50] = 24.749
	getLifeexpectancy[51] = 24.749
	getLifeexpectancy[52] = 24.749
	getLifeexpectancy[53] = 24.749
	getLifeexpectancy[54] = 24.749
	getLifeexpectancy[55] = 21.034
	getLifeexpectancy[56] = 21.034
	getLifeexpectancy[57] = 21.034
	getLifeexpectancy[58] = 21.034
	getLifeexpectancy[59] = 21.034
	getLifeexpectancy[60] = 17.498
	getLifeexpectancy[61] = 17.498
	getLifeexpectancy[62] = 17.498
	getLifeexpectancy[63] = 17.498
	getLifeexpectancy[64] = 17.498
	getLifeexpectancy[65] = 14.217
	getLifeexpectancy[66] = 14.217
	getLifeexpectancy[67] = 14.217
	getLifeexpectancy[68] = 14.217
	getLifeexpectancy[69] = 14.217
	getLifeexpectancy[70] = 11.217
	getLifeexpectancy[71] = 11.217
	getLifeexpectancy[72] = 11.217
	getLifeexpectancy[73] = 11.217
	getLifeexpectancy[74] = 11.217
	getLifeexpectancy[75] = 8.537
	getLifeexpectancy[76] = 8.537
	getLifeexpectancy[77] = 8.537
	getLifeexpectancy[78] = 8.537
	getLifeexpectancy[79] = 8.537
	getLifeexpectancy[80] = 6.202
	getLifeexpectancy[81] = 6.202
	getLifeexpectancy[82] = 6.202
	getLifeexpectancy[83] = 6.202
	getLifeexpectancy[84] = 6.202
	getLifeexpectancy[85] = 6.202
	getLifeexpectancy[86] = 6.202
	getLifeexpectancy[87] = 6.202
	getLifeexpectancy[88] = 6.202
	getLifeexpectancy[89] = 6.202
	getLifeexpectancy[90] = 6.202
	getLifeexpectancy[91] = 6.202
	getLifeexpectancy[92] = 6.202
	getLifeexpectancy[93] = 6.202
	getLifeexpectancy[94] = 6.202
	getLifeexpectancy[95] = 6.202
	getLifeexpectancy[96] = 6.202
	getLifeexpectancy[97] = 6.202
	getLifeexpectancy[98] = 6.202
	getLifeexpectancy[99] = 6.202
	getLifeexpectancy[100] = 6.202
	getLifeexpectancy[101] = 6.202
	getLifeexpectancy[102] = 6.202
	getLifeexpectancy[103] = 6.202
	getLifeexpectancy[104] = 6.202
	getLifeexpectancy[105] = 6.202
	getLifeexpectancy[106] = 6.202
	getLifeexpectancy[107] = 6.202
	getLifeexpectancy[108] = 6.202
	getLifeexpectancy[109] = 6.202
	getLifeexpectancy[110] = 6.202

	lifeExpectancy := getLifeexpectancy[age]

	return lifeExpectancy
}

func getOtherDeathStateByModel(localInputsPointer *Input, model Model) State {
	otherDeathStateId := localInputsPointer.QueryData.Other_death_state_by_model[model.Id]
	otherDeathState := get_state_by_id(localInputsPointer, otherDeathStateId)
	return otherDeathState
}

// This represents running the full model for one person
func runFullModelForOnePerson(localInputs Input, person Person, masterRecordsToAdd chan []MasterRecord) {

	// --------- FIX WITH OTHER DEATHS ==============

	// localInputsPointer := &localInputs

	// mrSize := len(localInputsPointer.Cycles) * len(localInputsPointer.Models)
	// theseMasterRecordsToAdd := make([]MasterRecord, mrSize, mrSize)
	// mrIndex := 0
	// //fmt.Println("Person:", person.Id)
	// for _, cycle := range localInputsPointer.Cycles { // foreach cycle
	// 	//fmt.Println("Cycle: ", cycle.Name)
	// 	//shuffled := shuffle(localInputsPointer.Models) // randomize the order of the models //TODO place back in not sure why broken.
	// 	for _, model := range localInputsPointer.Models { // foreach model
	// 		//fmt.Println(model.Name)
	// 		runCyclePersonModel(localInputsPointer, cycle, model, person, &theseMasterRecordsToAdd, mrIndex)
	// 		mrIndex++
	// 	} // end foreach model
	// 	localInputsPointer.CurrentCycle++

	// } //end foreach cycle

	// //Timer := nitro.Initialize()

	// masterRecordsToAdd <- theseMasterRecordsToAdd
}

func runOneCycleForOnePerson(localInputs *Input, cycle Cycle, person Person, masterRecordsToAdd chan []MasterRecord) {

	localInputsPointer := localInputs
	//small MR size bc just for one person
	mrSize := len(localInputsPointer.Models)
	theseMasterRecordsToAdd := make([]MasterRecord, mrSize, mrSize)
	mrIndex := 0
	for _, model := range localInputsPointer.Models { // foreach model
		runCyclePersonModel(localInputsPointer, cycle, model, person, &theseMasterRecordsToAdd, mrIndex)
	}
	// Below iteration finds the new states. This needs to be done here
	// in case someone died - even if someone dies in the "last" model,
	// that deaths forces a death in all other models
	for _, model := range localInputsPointer.Models { // foreach model
		var newMasterRecord MasterRecord
		newMasterRecord.Cycle_id = cycle.Id + 1
		newMasterRecord.Person_id = person.Id
		newMasterRecord.State_id = localInputsPointer.QueryData.State_id_by_cycle_and_person_and_model[cycle.Id+1][person.Id][model.Id]
		newMasterRecord.Model_id = model.Id
		theseMasterRecordsToAdd[mrIndex] = newMasterRecord
		mrIndex++
	}

	masterRecordsToAdd <- theseMasterRecordsToAdd

}

// func runModelWithConcurrentPeopleWithinCycle(person Person, cycle Cycle) {
// 	fmt.Println("Person:", person.Id)
// 	fmt.Println("Cycle: ", cycle.Name)
// 	shuffled := shuffle(Models)      // randomize the order of the models
// 	for _, model := range shuffled { // foreach model
// 		runPersonCycleModel(person, cycle, model)
// 	} // end foreach model
// }

func setUpQueryData(Inputs Input, numberOfPeople int, numberOfPeopleEntering int) Input {
	// Need to have lengths to be able to access them
	//Cycles

	numberOfPeople = numberOfPeople + numberOfPeopleEntering
	fmt.Println("Total num", numberOfPeople)

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
			Inputs.QueryData.Interactions_id_by_in_state_and_model[i][r] = 99999999 // TODO placeholder value to represent no interaction
		}
	}

	for _, interaction := range Inputs.Interactions {
		// if person is in a state with an interaction that effects current model
		Inputs.QueryData.Interactions_id_by_in_state_and_model[interaction.In_state_id][interaction.Effected_model_id] = interaction.Id
	}

	Inputs.QueryData.Model_id_by_state = make([]int, len(Inputs.States), len(Inputs.States))

	for _, state := range Inputs.States {
		Inputs.QueryData.Model_id_by_state[state.Id] = state.Model_id
	}

	/* TODO Fix this hack. We actually end up storing len(Cycles)+1 cycles,
	because we start on 0 and calculate the cycle ahead of us, so if we have
	up to cycle 19 in the inputs, we will calculate 0-19, as well as cycle 20 */

	numberOfCalculatedCycles := len(Inputs.Cycles) + 1

	Inputs.QueryData.State_populations_by_cycle = make([][]int, numberOfCalculatedCycles, numberOfCalculatedCycles)
	for c := 0; c < numberOfCalculatedCycles; c++ {
		Inputs.QueryData.State_populations_by_cycle[c] = make([]int, len(Inputs.States), len(Inputs.States))
	}

	Inputs.QueryData.Other_death_state_by_model = make([]int, len(Inputs.Models), len(Inputs.Models))
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

		Inputs.QueryData.Other_death_state_by_model[model.Id] = otherDeathState.Id

	}

	Timer.Step("set up query data")
	return Inputs
}

func initializeGlobalStatePopulations(Inputs Input) Input {
	/* TODO Fix this hack. We actually end up storing len(Cycles)+1 cycles,
	because we start on 0 and calculate the cycle ahead of us, so if we have
	up to cycle 19 in the inputs, we will calculate 0-19, as well as cycle 20 */
	numberOfCalculatedCycles := len(Inputs.Cycles) + 1
	GlobalStatePopulations = make([]StatePopulation, numberOfCalculatedCycles*len(Inputs.States))
	q := 0
	for c := 0; c < numberOfCalculatedCycles; c++ {
		for s, _ := range Inputs.States {
			GlobalStatePopulations[q].Cycle_id = c
			GlobalStatePopulations[q].Id = q
			GlobalStatePopulations[q].Population = 0
			GlobalStatePopulations[q].State_id = s
			GlobalStatePopulations[q].Model_id = Inputs.QueryData.Model_id_by_state[s]
			q++
		}
	}
	return Inputs
}
func setUpGlobalMasterRecordsByIPCM(Inputs Input) {

	GlobalMasterRecordsByIPCM = make([][][][]int, numberOfIterations, numberOfIterations)
	for i := 0; i < numberOfIterations; i++ {
		GlobalMasterRecordsByIPCM[i] = make([][][]int, numberOfPeople, numberOfPeople)
		for p := 0; p < numberOfPeople; p++ {
			GlobalMasterRecordsByIPCM[i][p] = make([][]int, len(Inputs.Cycles)+1, len(Inputs.Cycles)+1) // TODO cycles hack
			for q := 0; q < len(Inputs.Cycles)+1; q++ {
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

// Since we are using an open cohort, we need to add people to the
// simulation every year - these represent people that are being
// born into the simulation
func createNewPeople(Inputs *Input, cycle Cycle, number int) {
	idForFirstPerson := len(Inputs.People) + 1
	//newPeople = make([]Person, number, number)
	for i := 0; i < number; i++ {
		newPerson := Person{idForFirstPerson + i}
		//newPeople[i] = newPerson
		Inputs.People = append(Inputs.People, newPerson)
		//fmt.Println("new person", newPerson.Id)
		for _, model := range Inputs.Models {
			// TODO fix hack here - this should be more systematic
			// Place person into correct age category
			uninitializedState := State{}
			uninitializedState = model.get_uninitialized_state(Inputs)
			if model.Name == "Age" {
				// Start them at age 20
				// TODO they will enter the model at age 21?
				uninitializedState = get_state_by_id(Inputs, 55)
			}
			//fmt.Println("unit state", uninitializedState)
			var mr MasterRecord
			mr.Cycle_id = cycle.Id
			mr.State_id = uninitializedState.Id
			mr.Model_id = model.Id
			mr.Person_id = newPerson.Id

			qd := Inputs.QueryData.State_id_by_cycle_and_person_and_model

			qd[mr.Cycle_id][mr.Person_id][mr.Model_id] = mr.State_id

			Inputs.MasterRecords = append(Inputs.MasterRecords, mr)
		}
	}
}

// create people will generate individuals and add their data to the master
// records
func createInitialPeople(Inputs Input, number int) Input {
	for i := 0; i < number; i++ {
		Inputs.People = append(Inputs.People, Person{i})
	}

	for _, person := range Inputs.People {
		for _, model := range Inputs.Models {
			uninitializedState := model.get_uninitialized_state(&Inputs)
			var mr MasterRecord
			mr.Cycle_id = 0
			mr.State_id = uninitializedState.Id
			mr.Model_id = model.Id
			mr.Person_id = person.Id
			// generate a hash key for a map, allows easy access to states
			// by hashing cycle, person and model.
			qd := Inputs.QueryData.State_id_by_cycle_and_person_and_model

			qd[mr.Cycle_id][mr.Person_id][mr.Model_id] = mr.State_id

			// this inputs will go into the threads of the model
			Inputs.MasterRecords = append(Inputs.MasterRecords, mr)

			// this inputs is the master inputs and is used to display data
			// at the end of the cycle
			GlobalMasterRecords = append(Inputs.MasterRecords, mr)

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
		//fmt.Println("unint state after cycle 0!")
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
func (model *Model) get_uninitialized_state(Inputs *Input) State {
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
func addToQueryDataMasterRecord(localInputs *Input, cycle Cycle, person Person, newState State) bool {
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

	GlobalTPsByRAS = make([]TPByRAS, numberOfRecords, numberOfRecords)
	var tpbrsPtr []interface{}
	for i := 0; i < numberOfRecords; i++ {
		tpbrsPtr = append(tpbrsPtr, new(TPByRAS))
	}
	ptrs = fromCsv(filename, GlobalTPsByRAS[0], tpbrsPtr)
	for i, ptr := range tpbrsPtr {
		GlobalTPsByRAS[i] = *ptr.(*TPByRAS)
	}

	return Inputs
}

func getTransitionProbByRAS(localInputsPointer *Input, currentStateInThisModel State, states []State, person Person) []TransitionProbability {

	var tpsToReturn []TransitionProbability

	modelId := currentStateInThisModel.Model_id

	raceModel := localInputsPointer.Models[4]
	raceStateId := person.get_state_by_model(localInputsPointer, raceModel).Id

	sexModel := localInputsPointer.Models[5]
	sexStateId := person.get_state_by_model(localInputsPointer, sexModel).Id

	ageModel := localInputsPointer.Models[10]
	ageModelId := person.get_state_by_model(localInputsPointer, ageModel).Id
	actualAge := ageModelId - 35

	for _, tpByRAS := range GlobalTPsByRAS {
		if tpByRAS.Model_id == modelId && tpByRAS.Race_state_id == raceStateId && tpByRAS.Age == actualAge && tpByRAS.Sex_state_id == sexStateId {
			//fmt.Println(tpByRAS.Model_id, modelId, tpByRAS.Race_state_id, raceStateId, tpByRAS.Age, actualAge, tpByRAS.Sex_state_id, sexStateId)
			var newTp TransitionProbability
			newTp.To_id = tpByRAS.To_state_id
			newTp.Tp_base = tpByRAS.Probability
			tpsToReturn = append(tpsToReturn, newTp)
		}
	}

	if len(tpsToReturn) < 1 {
		fmt.Println("No TPs found with getTransitionProbByRAS for m r a s", modelId, raceStateId, actualAge, sexStateId)
		os.Exit(1)
	}

	return tpsToReturn
}
