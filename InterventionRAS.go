package main

import (
	"fmt"
	"os"
)

//Run the complete model for each intervention:
//We receive all inputs and we get which specific intervention is being modeled.
func interventionInitiate(Intervention Intervention) {

	//Find the name of the current intervention.
	currentInterventionName := Intervention.Name

	//Proceed depending on which name is found.
	switch currentInterventionName {

	case "control":
		//= without intervention, no action.

		// TODO: Give Alex the new InterventionValue file and tell him to change the two lines that adjust the low risk TP here.
		// Also, tel Alex to change the first if statement to the names instead of the numbers 1 and 2
		// Also, make a check that exits when nothing gets changed.

	case "reduction by 0.2":
		//For the full range of the InterventionValues (this is a seperate input file with all relevant adjustment factors per intervention)
		for _, eachInterventionValue := range Inputs.InterventionValues {
			if len(Inputs.InterventionValues) < 1 {
				fmt.Println("No interventionValues found ", eachInterventionValue.Id)
				os.Exit(1)
			}
			//fmt.Println(eachInterventionValue.Adjustment_factor)
			//If the interventionId matches the name for this case
			if eachInterventionValue.Name == currentInterventionName {
				//for the full range of TPbyRASs inputs
				for _, eachTPByRas := range Inputs.TPByRASs {
					//If the values for to_state, age, sex, ethnicity in the RAS file match the ones in the InterventionValue file
					if eachInterventionValue.To_state_id == eachTPByRas.To_state_id && eachInterventionValue.Age_state_id == eachTPByRas.Age_state_id && eachInterventionValue.Sex_state_id == eachTPByRas.Sex_state_id && eachInterventionValue.Race_state_id == eachTPByRas.Race_state_id {
						//Multiply the RAS value by the factor of reduction
						//fmt.Println("After match, before adjustment: ", Inputs.TPByRASs[eachTPByRas.Id].Probability)
						oldProbability := Inputs.TPByRASs[eachTPByRas.Id].Probability
						//The InterventionValue file only has the to_state Id for high risk due to sugar consumption, so it will adjust each of these probabilities down.
						Inputs.TPByRASs[eachTPByRas.Id].Probability = eachTPByRas.Probability * eachInterventionValue.Adjustment_factor
						//Since we want the total TP of high risk & low risk to equal 1, we search for the RASinput that matches the one we just changed, but with the other to_state
						//fmt.Println("After match, after adjustment: ", Inputs.TPByRASs[eachTPByRas.Id].Probability)
						newProbability := Inputs.TPByRASs[eachTPByRas.Id].Probability
						//Go through the whole ras file again
						if oldProbability == newProbability {
							fmt.Println("Was not able to change ras values of: ", eachTPByRas.Id, "from: ", Inputs.TPByRASs[eachTPByRas.Id].Probability)
							os.Exit(1)
						}

						for _, eachRasLowRiskTP := range Inputs.TPByRASs {
							//We need to search according to ID's
							eachRasLowRiskTPId := eachRasLowRiskTP.Id
							//If we find the specific RASinput that matches the model, age, sex and ethnicity of the probability we changed, but does not match the To_state,
							// we know we have found the corresponding low_risk probability

							if Inputs.TPByRASs[eachRasLowRiskTPId].To_state_id != eachTPByRas.To_state_id && eachTPByRas.Age_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Age_state_id && eachTPByRas.Sex_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Sex_state_id && eachTPByRas.Race_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Race_state_id && eachTPByRas.Model_id == Inputs.TPByRASs[eachRasLowRiskTPId].Model_id {
								//So now we change this probability by 1.00 - the value we set for the high risk probability
								//fmt.Println("After match2, before adjustment low risk: ", Inputs.TPByRASs[eachRasLowRiskTPId].Probability)
								Inputs.TPByRASs[eachRasLowRiskTPId].Probability = 1.00 - Inputs.TPByRASs[eachTPByRas.Id].Probability
								fmt.Println("After match2, after adjustment low risk: ", Inputs.TPByRASs[eachRasLowRiskTPId].Probability)
							}
						}
					}
				}
			}
		}

	case "reduction by 0.5":

		for _, eachInterventionValue := range Inputs.InterventionValues { //This is not correct, I want only the ones with InterventionID = 1
			if eachInterventionValue.Name == currentInterventionName {
				//fmt.Println(eachInterventionValue.Adjustment_factor)

				for _, eachTPByRas := range Inputs.TPByRASs {
					if eachInterventionValue.To_state_id == eachTPByRas.To_state_id && eachInterventionValue.Age_state_id == eachTPByRas.Age_state_id && eachInterventionValue.Sex_state_id == eachTPByRas.Sex_state_id && eachInterventionValue.Race_state_id == eachTPByRas.Race_state_id {
						Inputs.TPByRASs[eachTPByRas.Id].Probability = eachTPByRas.Probability * eachInterventionValue.Adjustment_factor
						//rasTPHighRiskId := eachTPByRas.Id
						//Inputs.TPByRASs[rasTPHighRiskId+546].Probability = 1.00 - eachTPByRas.Probability //Get the corresponding low-risk state. HARDCODE = Wrong
						for _, eachRasLowRiskTP := range Inputs.TPByRASs {
							eachRasLowRiskTPId := eachRasLowRiskTP.Id
							if Inputs.TPByRASs[eachRasLowRiskTPId].To_state_id != eachTPByRas.To_state_id && eachTPByRas.Age_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Age_state_id && eachTPByRas.Sex_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Sex_state_id && eachTPByRas.Race_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Race_state_id && eachTPByRas.Model_id == Inputs.TPByRASs[eachRasLowRiskTPId].Model_id {
								Inputs.TPByRASs[eachRasLowRiskTPId].Probability = 1.00 - Inputs.TPByRASs[eachTPByRas.Id].Probability
								fmt.Println("After match2, after adjustment low risk: ", Inputs.TPByRASs[eachRasLowRiskTPId].Probability)
							}
						}
					}
				}
			}
		}

	default:
		fmt.Println("Cannot find the case for intervention ", currentInterventionName)
		os.Exit(1)

	} // End switch

}
