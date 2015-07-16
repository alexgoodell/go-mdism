package main

import (
	"encoding/csv"
	"fmt"
	"github.com/leesper/go_rng" //imported as rng
	"os"
	"reflect"
	"time"
)

//Initialize the structure of the psa input file
type PsaInputs struct {
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

//Initialize the types with a slice where all 10,000 numbers of a specific variable can be stored.
type AllBetaRandsPerVariable struct {
	BetaRandPerVariable []float64
}

type AllGammaRandsPerVariable struct {
	GammaRandPerVariable []float64
}

type AllNormalRandsPerVariable struct {
	NormalRandPerVariable []float64
}

//Initialize slice of these slices, so that we can store the 10,000 randomly generated number for each variable, not just 1.
// I am not completely sure if this actually works. Will GO automatically know that I want a specific slice of float64's
// for each of these variables? Or will it write over the previously generated 10,000 numbers when I move to the next variable?
// Do I need to work with structures somehow?
var AllBetaRandVariables []AllBetaRandsPerVariable
var AllGammaRandVariables []AllGammaRandsPerVariable
var AllNormalRandVariables []AllNormalRandsPerVariable

//This function is the same as with the other inputs. I want to have access to all the data I just loaded from the input file
// Do I need to initialize anything else to access these inputs after running this function?
func InitializePSA(inputsPath string) {
	// ####################### PSA inputs #######################

	filename = "inputs/" + inputsPath + "/psa.csv"
	numberOfRecords = getNumberOfRecords(filename)

	Inputs.Psa = make([]psaInputs, numberOfRecords, numberOfRecords)
	var psaPtr []interface{}
	for i := 0; i < numberOfRecords; i++ {
		psaPtr = append(psaPtr, new(psaInputs))
	}
	ptrs = fromCsv(filename, Inputs.Psa[0], psaPtr)
	for i, ptr := range psaPtr {
		Inputs.Psa[i] = *ptr.(*psaInputs)
	}
} // When I type Inputs.Psa it still doesn't find anything. What did I do wrong?

func main() {

	//I guess we implement this code after line 76 (Querysetup()) in index.go.
	//That way, the inputs are already loaded so you can freely change parameters, but the initial people are not yet created,
	// so there will be new people every iterations.
	InitializePSA("example") //See above function

	for _, mean := range Inputs.Psa { //for each row in the psa input file
		q := 0
		if Inputs.Psa.Distribution == "beta" { // if the specified type is beta
			betaGen := rng.NewBetaGenerator(time.Now().UnixNano()) // seed the generator
			for i := 0; i < 10000; i++ {                           // do the next thing ten thousand times
				AllBetaRandVariables[q].BetaRandPerVariable[i] = betaGen.Beta(Inputs.Psa.Alpha, Inputs.Psa.beta)
				q++
				// generate ten thousand
				// random numbers following the beta distribution according to the alpha and beta that are specified for this
				// specific psaInput row (this variable), and put the outcome of this generator into a slice that
				// is within a slice. Then I update q, so when we move to the next psaInput and that is
				// also a variable with a beta distribution, the 10,000 random numbers get
				// put in a different slice specific for this second variable.
				// I do need to think about how to keep track of which input goes into which number of the slice.
				// Maybe it is better to split the input file into 3 files, one for each distribution sort?
				// Or: I can also store the PSA_Id in the struc?
				// Does it work how I coded it right now? With a slice inside a slice? Or is there a better way?
				// Does it work if i put:
				// AllBetaRandVariables[q][i] = betaGen.Beta(Inputs.psa.Alpha, Inputs.psa.beta)
			}

		} else if Inputs.Psa.Distribution == "gamma" { // Exactly the same for gamma inputs.
			gammaGen := rng.NewGammaGenerator(time.Now().UnixNano())
			for i := 0; i < 10000; i++ {
				AllGammaRandVariables[q].GammaRandPerVariable[i] = gammaGen.Gamma(Inputs.Psa.Alpha, Inputs.Psa.beta)
				q++
			}

		} else if Inputs.Psa.Distribution == "normal" { //The random function with normal works a little different,
			// but how we store the values works the same.
			rand.Seed(time.Now().UnixNano())
			for i := 0; i < 10000; i++ {
				AllNormalRandVariables[q].NormalRandPerVariable[i] = Inputs.Psa.Mean + Inputs.Psa.SD*rand.Normfloat64()
				q++
			}

		} else if Inputs.Psa.Distribution == "none" { //There is one input that says "none" for all the variables
			// (Tps/Interactions et cetera), that do not need to be varied in the PSA. So we do not need a function for this one.

		} else { //If it's not one of these 4, it should give an error and give me at which row it fails and why.
			fmt.Println("Cannot generate randoms from distributions ", Inputs.Psa.Id, Inputs.Psa.Distribution)
			os.Exit(1)
		}
	}

	//So if everything above has worked, we need to now start to implement the values that we generated to the values that
	// were in the input files. And when we change a TP, we also need to make sure they still add to 1.

	//I will work on this, but maybe it's best if the work before is checked before I proceed.

	for i := 0; i < 10000; i++ { //For 10,000 iterations
		for _, costs := range Inputs.Costs {
			if Inputs.Costs[costs].PSA_id > 0 {
				Inputs.Costs[costs].Costs = AllGammaRandVariables[costs][i]
			}
		}

		_ = mean //Prevent warning

	}

	//OR, a different approach:
	for _, psaId := range Inputs.Psa {
		if Inputs.PSA.Id > 0 && Inputs.PSA.Id < 9 {
			Inputs.Costs[psaId].Costs = AllGammaRandVariables[psaId][i]
		}

	}
}
