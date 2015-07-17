package main

import (
	"math/rand"
	"time"

	"github.com/leesper/go_rng" //imported as rng
)

type PsaInput struct {
	Id           int
	Variable     string
	Input_file   string
	Distribution string
	Min          float64
	Max          float64
	Mean         float64
	SD           float64
	Alpha        float64
	Beta         float64
	Value        float64
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

type Interaction struct {
	Id                int
	In_state_id       int
	From_state_id     int
	To_state_id       int
	Adjustment        float64
	Effected_model_id int
	PSA_id            int
}

type DisabilityWeight struct {
	Id                int
	State_id          int
	Disability_weight float64
	PSA_id            int
}

type RegressionRate struct {
	Id              int
	From_state      int
	To_state        int
	Age_low         int
	Age_high        int
	Regression_rate float64
	Psa_id          int
}

type TPByRAS struct {
	Id                  int
	Model_id            int
	Model_name          string
	No_disease_state_id int
	To_state_id         int
	To_state_name       string
	Sex_state_id        int
	Race_state_id       int
	Age_state_id        int
	Probability         float64
	PSA_id              int
}

type Input struct {
	TransitionProbabilities []TransitionProbability
	Interactions            []Interaction
	Costs                   []Cost
	DisabilityWeights       []DisabilityWeight
	TPByRASs                []TPByRAS
	PsaInputs               []PsaInput
	RegressionRates         []RegressionRate
}

var Inputs Input

// Add this code to the 'InitializeInputs' function:
// ####################### PSA inputs #######################

/*	filename = "inputs/" + inputsPath + "/psa.csv"
	numberOfRecords = getNumberOfRecords(filename)

	Inputs.PsaInputs = make([]PsaInput, numberOfRecords, numberOfRecords)
	var PsaPtrs []interface{}
	for i := 0; i < numberOfRecords; i++ {
		PsaPtrs = append(PsaPtrs, new(PsaInput))
	}
	Psaptrs = fromCsv(filename, Inputs.PsaInputs[0], Psaptrs)
	for i, ptr := range PsaPtrs {
		Inputs.PsaInputs[i] = *ptr.(*PsaInput)
	}
*/

func main() {
	Inputs.TransitionProbabilities = make([]TransitionProbability, 10, 10)
	Inputs.Costs = make([]Cost, 10, 10)
	Inputs.DisabilityWeights = make([]DisabilityWeight, 10, 10)
	Inputs.Interactions = make([]Interaction, 10, 10)
	Inputs.TPByRASs = make([]TPByRAS, 10, 10)
	Inputs.RegressionRates = make([]RegressionRate, 10, 10)

}

func generateNewValue(psaInput PsaInput) float64 {
	var valueToReturn float64
	switch psaInput.Distribution {

	case "beta":
		betaGen := rng.NewBetaGenerator(time.Now().UnixNano()) // seed the generator
		valueToReturn = betaGen.Beta(psaInput.Alpha, psaInput.Beta)

	case "gamma":
		gammaGen := rng.NewGammaGenerator(time.Now().UnixNano()) // seed the generator
		valueToReturn = gammaGen.Gamma(psaInput.Alpha, psaInput.Beta)

	case "normal":
		rand.Seed(time.Now().UnixNano()) // seed the generator -> Do we need to do this again? ALEX
		valueToReturn = psaInput.Mean + psaInput.SD*rand.NormFloat64()

	}
	return valueToReturn
}

func generateAllPsaValues(psaInputs PsaInput) PsaInput {
	for i := 0; i < len(psaInputs); i++ {
		psaInputPtr = &psaInput[i]
		psaInputPtr.Value = generateNewValue(psaInput)
	}
}

func runPsa(Inputs Input) {
	for i := 0; i < len(Inputs.PsaInputs); i++ {
		psaInput := Inputs.PsaInputs[i]
		inputFile := psaInput.Input_file

		switch inputFile {

		case "transition-probabilities":

			for p := 0; p < len(Inputs.TransitionProbabilities); p++ {
				transitionProbability := Inputs.TransitionProbabilities[p]
				if transitionProbability.PSA_id == psaInput.Id && transitionProbability.PSA_id != 0 {
					//Need to make sure that each variable that has the same PSA_id
					// gets the same newValue (not generate a new one for each).
					//Make this a 'for' statement instead? For each time they equal each other? ALEX
					transitionProbability.Tp_base = psaInput.Value
				}
			}

			// Now make sure everything adds up to 1 again.
			// First get the sum of all transition per 'from' state
			// Then set the TP for people that remain to 1 - this sum.
			// Find a specific way to do this for the ras file -> there is a slice of TPs that all need to be changed by the same value.

			for fromState := 0; fromState < 43; fromState++ { //Do this for all relevant from states.
				// I have set that at 42, but might be nicer to use len()? But then I should take len(Inputs.States) ?
				// It is not really necessary, because we don't want him to change anything to the age model, so nothing above 42.
				sumThisFromState := make([]float64, 150, 150) // Need to make this len(Inputs.States) as well.
				// use tps := Query.Tps_id_by_from_state[fromState]
				for _, eachTP := range Inputs.TransitionProbabilities { //For each of the TPs
					if eachTP.From_id == fromState && eachTP.From_id != eachTP.To_id {
						// If the from ID equals the from state we are assessing right now and the TP is not for staying in the same state
						sumThisFromState[fromState] += eachTP.Tp_base
						// Add the TPbase of this specific TP to the sum (of all of the TPs from this state)
					}
				} //Now that we have all the sums per each state in a slice, we are going to subtract this from the remaining.
				for _, eachTP := range Inputs.TransitionProbabilities {
					if eachTP.From_id == fromState && eachTP.From_id == eachTP.To_id {
						// If we come to the TP of this specific fromstate, and this TP is for staying in that state
						eachTP.Tp_base = 1.00 - sumThisFromState[fromState]
						// correct the TP_base by the sum you found from the other TPs.
					}
				}

				//Some checks here? Check_sum?
			}

		case "disability-weights":

			for p := 0; p < len(Inputs.DisabilityWeights); p++ {
				disabilityWeight := Inputs.DisabilityWeights[p]
				if disabilityWeight.PSA_id == psaInput.Id && disabilityWeight.PSA_id != 0 {
					newValue := generateNewValue(psaInput)
					disabilityWeight.Disability_weight = newValue
				}
			}

		case "costs":

			for p := 0; p < len(Inputs.Costs); p++ {
				cost := Inputs.Costs[p]
				if cost.PSA_id == psaInput.Id && cost.PSA_id != 0 {
					newValue := generateNewValue(psaInput)
					cost.Costs = newValue
				}
			}

		case "interactions":

			for p := 0; p < len(Inputs.Interactions); p++ {
				interactions := &Inputs.Interactions[p]
				if interactions.PSA_id == psaInput.Id && interactions.PSA_id != 0 {
					newValue := generateNewValue(psaInput)
					interactions.Adjustment = newValue
				}
			}

		case "regression-rates":

			for p := 0; p < len(Inputs.RegressionRates); p++ {
				regressionrate := Inputs.RegressionRates[p]
				if regressionrate.Psa_id == psaInput.Id && regressionrate.Psa_id != 0 {
					newValue := generateNewValue(psaInput)
					regressionrate.Regression_rate = newValue
				}
			}

		case "ras":

			for p := 0; p < len(Inputs.TPByRASs); p++ {
				tpByRas := Inputs.TPByRASs[p]
				if tpByRas.PSA_id == psaInput.Id && tpByRas.PSA_id != 0 {
					newValue := generateNewValue(psaInput)
					tpByRas.Probability = newValue
				}
			}

			for ras := range Inputs.TPByRASs {
				ageState := Query.getStateById(ras.Age_state_id)
				//
				//
				//
				sumThisModel := 0
				relevantRASs = Query.getTPsByRas(ageState, etc)
				for relevantRAS := range relevantRASs {
					if relevantRAS.No_disease_state_id != relevantRAS.To_state_id {
						sumThisModel += eachTP.Probability
					}
				}
				for relevantRAS := range relevantRASs {
					if relevantRAS.No_disease_state_id == relevantRAS.To_state_id {
						relevantRAS.Probability = 1.0 - sumThisModel
					}
				}

			}

			// I realise these things are hard coded now, but how would we set these ranges otherwise?
			for noDiseaseState := 0; noDiseaseState < 45; noDiseaseState++ { //In the ras file, No_disease_state_id == id for the non-disease state.
				//Maybe we can set the from state in the ras file as the stay state? It is technically incorrect, but would probably work.
				for rasSex := 33; rasSex < 36; rasSex++ { //For each possible combination of ras Sex
					for rasEthnicity := 28; rasEthnicity < 33; rasEthnicity++ { //For each combo of ras ethnicity
						for rasAge := 42; rasAge < 135; rasAge++ { // For each combo of ras age
							//We need to get the sum of all of the transitions from a specific state (within the corresponding age, sex and ethnicity)
							sumThisModel := make([]float64, 50, 50)  //len(Inputs.Models) ??
							for _, eachTP := range Inputs.TPByRASs { //For each of the TPByRASs
								// If the from ID equals the from state we are assessing right now, and it matches the specific r a and s, and the TP is not for staying in the same state
								if eachTP.No_disease_state_id == noDiseaseState && eachTP.No_disease_state_id != eachTP.To_state_id && eachTP.Sex_state_id == rasSex && eachTP.Race_state_id == rasEthnicity && eachTP.Age_state_id == rasAge {
									// Add the probability of this specific TP to the sum
									sumThisModel[noDiseaseState] += eachTP.Probability

								}
							}
							for _, eachTP := range Inputs.TPByRASs { //For each of the TPByRASs
								// If we come to the TP of this specific fromstate, and this TP is for staying in that state
								if eachTP.No_disease_state_id == noDiseaseState && eachTP.No_disease_state_id == eachTP.To_state_id && eachTP.Sex_state_id == rasSex && eachTP.Race_state_id == rasEthnicity && eachTP.Age_state_id == rasAge {
									// correct the TP_base by the sum you found from the other TPs.
									eachTP.Probability = 1.00 - sumThisModel[noDiseaseState]

								}
							}
							//Some checks here? Check_sum?

						}
					}
				}
			}

		} // end switch

	}
}
