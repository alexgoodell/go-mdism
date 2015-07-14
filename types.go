package main

type State struct {
	Id                        int
	Model_id                  int
	Name                      string
	Is_uninitialized_state    bool
	Is_uninitialized_2_state  bool
	Is_disease_specific_death bool
	Is_other_death            bool
	Is_natural_causes_death   bool
}

type Model struct {
	Id   int
	Name string
}

type LifeExpectancy struct {
	Id              int
	Age_state_id    int
	Sex_state_id    int
	Life_expectancy float64
}

type Intervention struct {
	Id                 int
	Name               string
	Fructose_reduction float64
}

type MasterRecord struct {
	Cycle_id               int
	Person_id              int
	State_id               int
	Model_id               int
	YLDs                   float64
	YLLs                   float64
	Costs                  float64
	Has_entered_simulation bool
	State_name             string
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
	PSA_id            int
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

type DisabilityWeight struct {
	Id                int
	State_id          int
	Disability_weight float64
	PSA_id            int
}

type InteractionKey struct {
	In_state_id   int
	From_state_id int
}

type RASkey struct {
	Race_state_id int
	Age_state_id  int
	Sex_state_id  int
	Model_id      int
}

type Query_t struct {
	State_id_by_cycle_and_person_and_model         [][][]int
	States_ids_by_cycle_and_person                 [][]int
	Tps_id_by_from_state                           [][]int
	interaction_id_by_in_state_and_from_state      map[InteractionKey][]int
	State_populations_by_cycle                     [][]int
	Model_id_by_state                              []int
	Other_death_state_by_model                     []int
	Cost_by_state_id                               []float64
	Disability_weight_by_state_id                  []float64
	Master_record_id_by_cycle_and_person_and_model [][][]int
	Life_expectancy_by_sex_and_age                 map[SexAge]float64
	TP_by_RAS                                      map[RASkey][]TPByRAS
	Unintialized_state_by_model                    []int
	Outputs_id_by_cycle_and_state                  [][]int

	// Unexported and used by the "getters"
	model_id_by_name map[string]int
	state_id_by_name map[string]int
}

type SexAge struct {
	Sex, Age int
}

type Input struct {
	//	CurrentCycle            int
	Models                  []Model
	People                  []Person
	States                  []State
	TransitionProbabilities []TransitionProbability
	Interactions            []Interaction
	Cycles                  []Cycle
	MasterRecords           []MasterRecord
	Costs                   []Cost
	DisabilityWeights       []DisabilityWeight
	LifeExpectancies        []LifeExpectancy
	TPByRASs                []TPByRAS
	Interventions           []Intervention
}

type TPByRAS struct {
	Id            int
	Model_id      int
	Model_name    string
	To_state_id   int
	To_state_name string
	Sex_state_id  int
	Race_state_id int
	Age_state_id  int
	Probability   float64
	PSA_id        int
}

// ##################### Output structs ################ //

//this struct will replicate the data found
/*type StatePopulation struct {
	Id         int
	State_name string
	State_id   int
	Cycle_id   int
	Population int
	Model_id   int
}
*/

type OutputByCycleState struct {
	Id         int
	YLLs       float64
	YLDs       float64
	DALYs      float64
	Costs      float64
	Cycle_id   int
	State_id   int
	Population int
	State_name string
}

type Output struct {
	OutputsByCycleStateFull []OutputByCycleState
	OutputsByCycleStatePsa  []OutputByCycleState
	OutputsByCycle          []OutputByCycle
}

type OutputByCycle struct {
	Cycle_id             int
	T2DM_diagnosis_event int
	T2DM_death_event     int
	CHD_diagnosis_event  int
	CHD_death_event      int
	HCC_diagnosis_event  int
	HCC_death_event      int
}
