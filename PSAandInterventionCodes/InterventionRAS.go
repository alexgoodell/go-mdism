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
	}
}

//Run the complete model for each intervention:
func interventionInitiate(Inputs Input, Intervention Intervention) {
	currentInterventionName := Intervention.Name

	switch currentInterventionName {

	case "control":
		//= without intervention
	case "reduction by 0.2":
		for _, eachInterventionValue := range Inputs.InterventionValues {
			if eachInterventionValue.Intervention_id == 1 {
				for _, eachTPByRas := range Inputs.TPByRASs {
					if eachInterventionValue.To_state_id == eachTPByRas.To_state_id && eachInterventionValue.Age_state_id == eachTPByRas.Age_state_id && eachInterventionValue.Sex_state_id == eachTPByRas.Sex_state_id && eachInterventionValue.Race_state_id == eachTPByRas.Race_state_id {
						eachTPByRas.Probability = eachTPByRas.Probability * eachInterventionValue.Adjustment_factor
						rasTPHighRiskId := eachTPByRas.Id
						Inputs.TPByRASs[rasTPHighRiskId+546].Probability = 1.00 - eachTPByRas.Probability //Get the corresponding low-risk state. HARDCODE = Wrong
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
						rasTPHighRiskId := eachTPByRas.Id
						Inputs.TPByRASs[rasTPHighRiskId+546].Probability = 1.00 - eachTPByRas.Probability //Get the corresponding low-risk state. HARDCODE = Wrong
					}
				}
			}
		}

	} // End switch

}
