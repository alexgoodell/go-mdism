package main

import ()

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
	Id            int
	Model_id      int
	Model_name    string
	From_state_id int
	To_state_id   int
	To_state_name string
	Sex_state_id  int
	Race_state_id int
	Age_state_id  int
	Probability   float64
	PSA_id        int
}

type InterventionValue struct {
	Id                int
	Intervention_id   int
	Name              string
	To_state_id       int
	Sex_state_id      int
	Race_state_id     int
	Age_state_id      int
	Adjustment_factor float64
}

type Intervention struct {
	Id   int
	Name string
}

type Input struct {
	TransitionProbabilities []TransitionProbability
	Interactions            []Interaction
	Costs                   []Cost
	DisabilityWeights       []DisabilityWeight
	TPByRASs                []TPByRAS
	PsaInputs               []PsaInput
	RegressionRates         []RegressionRate
	Interventions           []Intervention
	InterventionValues      []InterventionValue
}

var Inputs Input

func main() {
	Inputs.TransitionProbabilities = make([]TransitionProbability, 10, 10)
	Inputs.Costs = make([]Cost, 10, 10)
	Inputs.DisabilityWeights = make([]DisabilityWeight, 10, 10)
	Inputs.Interactions = make([]Interaction, 10, 10)
	Inputs.TPByRASs = make([]TPByRAS, 10, 10)
	Inputs.RegressionRates = make([]RegressionRate, 10, 10)
	Inputs.Interventions = make([]Intervention, 10, 10)
	Inputs.InterventionValues = make([]InterventionValue, 10, 10)

	for _, eachIntervention := range Inputs.Interventions {
		interventionInitiate(Inputs, eachIntervention)

		//func runModel et cetera...

	}
}

//Run the complete model for each intervention:
//We receive all inputs and we get which specific intervention is being modeled.
func interventionInitiate(Inputs Input, Intervention Intervention) {

	//Find the name of the current intervention.
	currentInterventionName := Intervention.Name

	//Proceed depending on which name is found.
	switch currentInterventionName {

	case "control":
		//= without intervention, no action.

	case "reduction by 0.2":
		//For the full range of the InterventionValues (this is a seperate input file with all relevant adjustment factors per intervention)
		for _, eachInterventionValue := range Inputs.InterventionValues {
			//If the interventionId matches the name for this case
			if eachInterventionValue.Intervention_id == 1 {
				//for the full range of TPbyRASs inputs
				for _, eachTPByRas := range Inputs.TPByRASs {
					//If the values for to_state, age, sex, ethnicity in the RAS file match the ones in the InterventionValue file
					if eachInterventionValue.To_state_id == eachTPByRas.To_state_id && eachInterventionValue.Age_state_id == eachTPByRas.Age_state_id && eachInterventionValue.Sex_state_id == eachTPByRas.Sex_state_id && eachInterventionValue.Race_state_id == eachTPByRas.Race_state_id {
						//Multiply the RAS value by the factor of reduction
						//The InterventionValue file only has the to_state Id for high risk due to sugar consumption, so it will adjust each of these probabilities down.
						eachTPByRas.Probability = eachTPByRas.Probability * eachInterventionValue.Adjustment_factor
						//Since we want the total TP of high risk & low risk to equal 1, we search for the RASinput that matches the one we just changed, but with the other to_state
						//Go through the whole ras file again
						for _, eachRasLowRiskTP := range Inputs.TPByRASs {
							//We need to search according to ID's
							eachRasLowRiskTPId := eachRasLowRiskTP.Id
							//If we find the specific RASinput that matches the model, age, sex and ethnicity of the probability we changed, but does not match the To_state,
							// we know we have found the corresponding low_risk probability
							if Inputs.TPByRASs[eachRasLowRiskTPId].To_state_id != eachTPByRas.To_state_id && eachTPByRas.Age_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Age_state_id && eachTPByRas.Sex_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Sex_state_id && eachTPByRas.Race_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Race_state_id && eachTPByRas.Model_id == Inputs.TPByRASs[eachRasLowRiskTPId].Model_id {
								//So now we change this probability by 1.00 - the value we set for the high risk probability
								Inputs.TPByRASs[eachRasLowRiskTPId].Probability = 1.00 - eachTPByRas.Probability
							}
						}
					}
				}
			}
		}

	case "reduction by 0.5":
		for _, eachInterventionValue := range Inputs.InterventionValues { //This is not correct, I want only the ones with InterventionID = 1
			if eachInterventionValue.Intervention_id == 2 {
				for _, eachTPByRas := range Inputs.TPByRASs {
					if eachInterventionValue.To_state_id == eachTPByRas.To_state_id && eachInterventionValue.Age_state_id == eachTPByRas.Age_state_id && eachInterventionValue.Sex_state_id == eachTPByRas.Sex_state_id && eachInterventionValue.Race_state_id == eachTPByRas.Race_state_id {
						eachTPByRas.Probability = eachTPByRas.Probability * eachInterventionValue.Adjustment_factor
						//rasTPHighRiskId := eachTPByRas.Id
						//Inputs.TPByRASs[rasTPHighRiskId+546].Probability = 1.00 - eachTPByRas.Probability //Get the corresponding low-risk state. HARDCODE = Wrong
						for _, eachRasLowRiskTP := range Inputs.TPByRASs {
							eachRasLowRiskTPId := eachRasLowRiskTP.Id
							if Inputs.TPByRASs[eachRasLowRiskTPId].To_state_id != eachTPByRas.To_state_id && eachTPByRas.Age_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Age_state_id && eachTPByRas.Sex_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Sex_state_id && eachTPByRas.Race_state_id == Inputs.TPByRASs[eachRasLowRiskTPId].Race_state_id && eachTPByRas.Model_id == Inputs.TPByRASs[eachRasLowRiskTPId].Model_id {
								Inputs.TPByRASs[eachRasLowRiskTPId].Probability = 1.00 - eachTPByRas.Probability
							}
						}
					}
				}
			}
		}

	} // End switch

}
