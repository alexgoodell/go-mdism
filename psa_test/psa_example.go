package main

import (
	"time"

	"github.com/leesper/go_rng"
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
}

type TransitionProbability struct {
	Id      int
	From_id int
	To_id   int
	Tp_base float64
	PSA_id  int
}

type Input struct {
	PsaInputs                []PsaInput
	TransaitionProbabilities []TransitionProbability
}

var Inputs Input

func main() {
	Inputs.TransaitionProbabilities = make([]TransitionProbability, 10, 10)
}

func generateNewValue(psaInput PsaInput) float64 {
	var valueToReturn float64
	switch psaInput.Distribution {

	case "beta":
		betaGen := rng.NewBetaGenerator(time.Now().UnixNano()) // seed the generator
		valueToReturn = betaGen.Beta(psaInput.Alpha, psaInput.Beta)

	case "gamma":
		// betaGen := rng.NewBetaGenerator(time.Now().UnixNano()) // seed the generator
		// valueToReturn = betaGen.Beta(Inputs.Psa.Alpha, Inputs.Psa.beta)

	case "normal":
		// betaGen := rng.NewBetaGenerator(time.Now().UnixNano()) // seed the generator
		// valueToReturn = betaGen.Beta(Inputs.Psa.Alpha, Inputs.Psa.beta)

	}
	return valueToReturn
}

func runPsa(Inputs Input) {
	for i := 0; i < len(Inputs.PsaInputs); i++ {
		psaInput := Inputs.PsaInputs[i]
		inputFile := psaInput.Input_file

		switch inputFile {

		case "transition_probabilities":

			for p := 0; p < len(Inputs.TransaitionProbabilities); p++ {
				transitionProbability := Inputs.TransaitionProbabilities[p]
				if transitionProbability.PSA_id == psaInput.Id {
					newValue := generateNewValue(psaInput)
					transitionProbability.Tp_base = newValue
				}
			}

		case "states":

			// for transitionProbability := range Inputs.TransaitionProbabilities {
			// 	if transitionProbability.PSA_id == psaInput.Id {
			// 		newValue := generateNewValue(psaInput)
			// 		transitionProbability.Tp_base = newValue
			// 	}
			// }

		} // end switch

	}
}
