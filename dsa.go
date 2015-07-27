package main

import (
	"fmt"
	"os"
)

func runNewDsaValue(variableCount int, withinVariableCount int) {
	//I want to keep track of how many iterations he has done, and how many with the same variable as DSA.

	for i := 0; i < len(Inputs.DsaInputs); i++ {
		Inputs.DsaInputs[i].Value = 0.0
	}

	fmt.Println("variableCount: ", variableCount)
	fmt.Println("withinVariableCount: ", withinVariableCount)

	switch withinVariableCount {
	case 1:
		//fmt.Print("Generating DSA values...")
		Inputs.DsaInputs[variableCount].Value = Inputs.DsaInputs[variableCount].Run1

	case 2:
		Inputs.DsaInputs[variableCount].Value = Inputs.DsaInputs[variableCount].Run2

	case 3:
		Inputs.DsaInputs[variableCount].Value = Inputs.DsaInputs[variableCount].Run3

	case 4:
		Inputs.DsaInputs[variableCount].Value = Inputs.DsaInputs[variableCount].Run4

	case 5:
		Inputs.DsaInputs[variableCount].Value = Inputs.DsaInputs[variableCount].Run5

	}
	//fmt.Print("complete.")

}

func runDsa() {

	//fmt.Println("here one")
	//fmt.Println()
	//fmt.Print("Setting PSA into inputs...")
	for i := 0; i < len(Inputs.DsaInputs); i++ {
		dsaInput := Inputs.DsaInputs[i]
		inputFile := dsaInput.Input_file

		if dsaInput.Value != 0 { //I only want to replace the one I just changed.
			// These values must all be reset to 0 after each iteration, is that the case at the moment?

			switch inputFile {

			case "transition-probabilities":

				for p := 0; p < len(Inputs.TransitionProbabilities); p++ {
					transitionProbability := &Inputs.TransitionProbabilities[p]
					if transitionProbability.PSA_id == dsaInput.Id && transitionProbability.PSA_id != 0 {
						//Need to make sure that each variable that has the same PSA_id
						// gets the same newValue (not generate a new one for each).
						//Make this a 'for' statement instead? For each time they equal each other? ALEX
						transitionProbability.Tp_base = dsaInput.Value
					}
				}

				//Making sure everything adds to one
				// occurs outside of the PsaInputs loop

			case "disability-weights":

				for p := 0; p < len(Inputs.DisabilityWeights); p++ {
					disabilityWeight := &Inputs.DisabilityWeights[p]
					if disabilityWeight.PSA_id == dsaInput.Id && disabilityWeight.PSA_id != 0 {
						disabilityWeight.Disability_weight = dsaInput.Value
					}
				}

			case "costs":

				for p := 0; p < len(Inputs.Costs); p++ {
					cost := &Inputs.Costs[p]
					if cost.PSA_id == dsaInput.Id && cost.PSA_id != 0 {
						cost.Costs = dsaInput.Value
					}
				}

			case "interactions":

				for p := 0; p < len(Inputs.Interactions); p++ {
					interactions := &Inputs.Interactions[p]
					if interactions.PSA_id == dsaInput.Id && interactions.PSA_id != 0 {
						interactions.Adjustment = dsaInput.Value
					}
				}

			case "regression-rates":

				for p := 0; p < len(Inputs.RegressionRates); p++ {
					regressionrate := &Inputs.RegressionRates[p]
					if regressionrate.Psa_id == dsaInput.Id && regressionrate.Psa_id != 0 {
						regressionrate.Regression_rate = dsaInput.Value
					}
				}

			case "ras":

				for p := 0; p < len(Inputs.TPByRASs); p++ {
					tpByRas := &Inputs.TPByRASs[p]
					if tpByRas.PSA_id == dsaInput.Id && tpByRas.PSA_id != 0 {
						tpByRas.Probability = dsaInput.Value
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

			relevantRASs := Query.getTpsByRAS(raceState, ageState, sexState, model)

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
			//fmt.Println(sumThisFromState)
			//fmt.Println("== State ", fromState, " ====")
			var sumThisFromState float64 // Need to make this len(Inputs.States) as well.
			// use tps := Query.Tp_ids_by_from_state[fromState]
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

					//fmt.Println("Old recursive tp was: ", Inputs.TransitionProbabilities[eachTP.Id].Tp_base)
					Inputs.TransitionProbabilities[eachTP.Id].Tp_base = 1.00 - sumThisFromState

					//fmt.Println("New recursive tp is: ", Inputs.TransitionProbabilities[eachTP.Id].Tp_base)
					// correct the TP_base by the sum you found from the other TPs.
				}
			}

			//Some checks here? Check_sum?
		}
		//fmt.Print("complete.")
		//fmt.Println()
	}
}
