//Before main:
type DALYs struct {
	Id         	int
	State_id  	int
	Cycle_id   	int
	YLD 		float64
	YLL 		float64
	DALY		float64
	Model_id   	int
}

var GlobalDALYs = []DALYs{}

//in MAIN:
	Inputs = initializeGlobalDALYs(Inputs)

//In func RunMODEL:

	for _, masterRecord := range GlobalMasterRecords {
		if masterRecord.Person_id > 100 {
		}
		Inputs.QueryData.DALYs_by_cycle[masterRecord.Cycle_id][masterRecord.State_id] += 1
	} // I do not understand what is happening here... Is this function really necessary?
	// are we updating each cycle and state by 1? It would make sense to update the cycle, but state? I do not understand.

	for s, allDALYs := range GlobalDALYs {
		GlobalDALYs[s].YLD = Inputs.QueryData.DALYs_by_cycle[allDALYs.Cycle_id][allDALYs.State_id]
		GlobalDALYs[s].YLL = Inputs.QueryData.DALYs_by_cycle[allDALYs.Cycle_id][allDALYs.State_id]
		GlobalDALYs[s].DALY = Inputs.QueryData.DALYs_by_cycle[allDALYs.Cycle_id][allDALYs.State_id]
	}//This one I also do not fully comprehend. For the full range of GlobalDALYs (states*cycles), the value of our outputs is equal to our inputs?
	//I don't get it.
	
	toCsv("output"+"/DALYs.csv", GlobalDALYs[0], GlobalDALYs)//I think this is fine like this.


//In func SETUPQUERYDATA:

Inputs.QueryData.DALYs_by_cycle = make([][]int, numberOfCalculatedCycles, numberOfCalculatedCycles)
	for c := 0; c < numberOfCalculatedCycles; c++ {
		Inputs.QueryData.DALYs_by_cycle[c] = make([]int, len(Inputs.States)-106, len(Inputs.States)-106)
		//We could do this only for states 1 - 27 (disease models). I just subtracted the amount of states above 27, is that ok?
	}


func initializeGlobalDALYs(Inputs Input) Input {
	/* TODO Fix this hack. We actually end up storing len(Cycles)+1 cycles,
	because we start on 0 and calculate the cycle ahead of us, so if we have
	up to cycle 19 in the inputs, we will calculate 0-19, as well as cycle 20 */
	numberOfCalculatedCycles := len(Inputs.Cycles) + 1
	GlobalDALYs = make([]DALYs, numberOfCalculatedCycles*len(Inputs.States))
	q := 0
	for c := 0; c < numberOfCalculatedCycles; c++ {
		for s, _ := range Inputs.States {
			GlobalDALYs[q].Cycle_id = c
			GlobalDALYs[q].Id = q
			GlobalDALYs[q].YLD = 0.00
			GlobalDALYs[q].YLL = 0.00
			GlobalDALYs[q].DALY = 0.00
			GlobalDALYs[q].State_id = s
			GlobalDALYs[q].Model_id = Inputs.QueryData.Model_id_by_state[s]
			q++
		}
	}
	return Inputs//This function I completely understand.
}