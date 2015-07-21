package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/leesper/go_rng" //imported as rng
)

// Add this code to the 'InitializeInputs' function:
// ####################### PSA inputs #######################

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

func generateAllPsaValues() {
	fmt.Print("Generating PSA values...")
	for i := 0; i < len(Inputs.PsaInputs); i++ {
		psaInputPtr := &Inputs.PsaInputs[i]
		psaInput := Inputs.PsaInputs[i]
		psaInputPtr.Value = generateNewValue(psaInput)
	}
	fmt.Print("complete.")

}

func runPsa() {

	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Println("here one")
	fmt.Println()
	fmt.Print("Setting PSA into inputs...")
	for i := 0; i < len(Inputs.PsaInputs); i++ {
		psaInput := Inputs.PsaInputs[i]
		inputFile := psaInput.Input_file

		switch inputFile {

		case "transition-probabilities":

			for p := 0; p < len(Inputs.TransitionProbabilities); p++ {
				transitionProbability := &Inputs.TransitionProbabilities[p]
				if transitionProbability.PSA_id == psaInput.Id && transitionProbability.PSA_id != 0 {
					//Need to make sure that each variable that has the same PSA_id
					// gets the same newValue (not generate a new one for each).
					//Make this a 'for' statement instead? For each time they equal each other? ALEX
					transitionProbability.Tp_base = psaInput.Value
				}
			}

			//Making sure everything adds to one
			// occurs outside of the PsaInputs loop

		case "disability-weights":

			for p := 0; p < len(Inputs.DisabilityWeights); p++ {
				disabilityWeight := &Inputs.DisabilityWeights[p]
				if disabilityWeight.PSA_id == psaInput.Id && disabilityWeight.PSA_id != 0 {
					newValue := generateNewValue(psaInput)
					disabilityWeight.Disability_weight = newValue
				}
			}

		case "costs":

			for p := 0; p < len(Inputs.Costs); p++ {
				cost := &Inputs.Costs[p]
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

		// case "regression-rates":

		// 	for p := 0; p < len(Inputs.RegressionRates); p++ {
		// 		regressionrate := Inputs.RegressionRates[p]
		// 		if regressionrate.Psa_id == psaInput.Id && regressionrate.Psa_id != 0 {
		// 			newValue := generateNewValue(psaInput)
		// 			regressionrate.Regression_rate = newValue
		// 		}
		// 	}

		case "ras":

			for p := 0; p < len(Inputs.TPByRASs); p++ {
				tpByRas := &Inputs.TPByRASs[p]
				if tpByRas.PSA_id == psaInput.Id && tpByRas.PSA_id != 0 {
					newValue := generateNewValue(psaInput)
					tpByRas.Probability = newValue
				}
			}

			// // I realise these things are hard coded now, but how would we set these ranges otherwise?
			// for noDiseaseState := 0; noDiseaseState < 45; noDiseaseState++ { //In the ras file, No_disease_state_id == id for the non-disease state.
			// 	//Maybe we can set the from state in the ras file as the stay state? It is technically incorrect, but would probably work.
			// 	for rasSex := 33; rasSex < 36; rasSex++ { //For each possible combination of ras Sex
			// 		for rasEthnicity := 28; rasEthnicity < 33; rasEthnicity++ { //For each combo of ras ethnicity
			// 			for rasAge := 42; rasAge < 135; rasAge++ { // For each combo of ras age
			// 				//We need to get the sum of all of the transitions from a specific state (within the corresponding age, sex and ethnicity)
			// 				sumThisModel := make([]float64, 50, 50)  //len(Inputs.Models) ??
			// 				for _, eachTP := range Inputs.TPByRASs { //For each of the TPByRASs
			// 					// If the from ID equals the from state we are assessing right now, and it matches the specific r a and s, and the TP is not for staying in the same state
			// 					if eachTP.No_disease_state_id == noDiseaseState && eachTP.No_disease_state_id != eachTP.To_state_id && eachTP.Sex_state_id == rasSex && eachTP.Race_state_id == rasEthnicity && eachTP.Age_state_id == rasAge {
			// 						// Add the probability of this specific TP to the sum
			// 						sumThisModel[noDiseaseState] += eachTP.Probability

			// 					}
			// 				}
			// 				for _, eachTP := range Inputs.TPByRASs { //For each of the TPByRASs
			// 					// If we come to the TP of this specific fromstate, and this TP is for staying in that state
			// 					if eachTP.No_disease_state_id == noDiseaseState && eachTP.No_disease_state_id == eachTP.To_state_id && eachTP.Sex_state_id == rasSex && eachTP.Race_state_id == rasEthnicity && eachTP.Age_state_id == rasAge {
			// 						// correct the TP_base by the sum you found from the other TPs.
			// 						eachTP.Probability = 1.00 - sumThisModel[noDiseaseState]

			// 					}
			// 				}
			// 				//Some checks here? Check_sum?

			// 			}
			// 		}
			// 	}
			// }

		} // end switch

	} // end psaInput iteration

	//since we've changed the inputs, we need to re-initialize the query parameter
	Query.setUp()

	for _, ras := range Inputs.TPByRASs {
		raceState := Inputs.States[ras.Race_state_id]
		ageState := Inputs.States[ras.Age_state_id]
		sexState := Inputs.States[ras.Sex_state_id]
		model := Inputs.Models[ras.Model_id]

		sumThisModel := 0.0

		relevantRASs := Query.getTpByRAS(raceState, ageState, sexState, model)

		for _, relevantRAS := range relevantRASs {
			if !relevantRAS.No_disease_state {
				sumThisModel += relevantRAS.Probability
			}
		}

		if equalFloat(sumThisModel, 1, 0.000000001) {
			sumThisModel = 1
		}

		for _, relevantRAS := range relevantRASs {
			if relevantRAS.No_disease_state {
				//find actual RAS input to change
				relRASPtr := &Inputs.TPByRASs[relevantRAS.Id]
				if relRASPtr.Id != relevantRAS.Id {
					fmt.Println("problem finding ras")
					os.Exit(1)
				}
				relRASPtr.Probability = 1.0 - sumThisModel

				if sumThisModel+relRASPtr.Probability != 1 {
					fmt.Println("doesn't equal one")
					os.Exit(1)
				}

			}
		}

	}

	for fromState := 0; fromState < len(Inputs.States); fromState++ { //Do this for all relevant from states.
		// I have set that at 42, but might be nicer to use len()? But then I should take len(Inputs.States) ?
		// It is not really necessary, because we don't want him to change anything to the age model, so nothing above 42.

		//fmt.Println("== State ", fromState, " ====")
		var sumThisFromState float64 // Need to make this len(Inputs.States) as well.
		// use tps := Query.Tps_id_by_from_state[fromState]
		for _, eachTP := range Inputs.TransitionProbabilities { //For each of the TPs
			if eachTP.From_id == fromState && eachTP.To_id != fromState {
				// If the from ID equals the from state we are assessing right now and the TP is not for staying in the same state
				//fmt.Println("tp id: ", eachTP.Id, " tp base: ", eachTP.Tp_base)
				sumThisFromState += eachTP.Tp_base
				// Add the TPbase of this specific TP to the sum (of all of the TPs from this state)
			}
		} //Now that we have all the sums per each state in a slice, we are going to subtract this from the remaining.

		if equalFloat(sumThisFromState, 1, 0.000000001) {
			sumThisFromState = 1
		}

		for _, eachTP := range Inputs.TransitionProbabilities {
			if eachTP.From_id == fromState && eachTP.To_id == fromState {
				// If we come to the TP of this specific fromstate, and this TP is for staying in that state
				//fmt.Println("Old recursive tp was: ", Inputs.TransitionProbabilities[eachTP.Id].Tp_base)
				Inputs.TransitionProbabilities[eachTP.Id].Tp_base = 1.00 - sumThisFromState

				//fmt.Println("New recursive tp is: ", Inputs.TransitionProbabilities[eachTP.Id].Tp_base)
				// correct the TP_base by the sum you found from the other TPs.
			}
		}

		//Some checks here? Check_sum?
	}
	fmt.Print("complete.")
	fmt.Println()
}
