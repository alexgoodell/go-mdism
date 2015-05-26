package main

import (
	// "encoding/json"
	// "flag"
	"fmt"
	// 	"github.com/alexgoodell/ghdmodel/models"
	// 	"io/ioutil"
	// 	"net/http"
	// 	"strconv"
	"math/rand"
	"time"
)

type State struct {
	Id                 int
	Model_id           int
	Name               string
	Initial_proportion float32
}

type Model struct {
	Id   int
	Name string
}

type Master struct {
	Cycle_id  int
	Person_id int
	State_id  int
}

type Cycle struct {
	Id   int
	Name string
}

type Person struct {
	Id int
}

type Interaction struct {
	Id            int
	In_state_id   int
	From_state_id int
	To_state_id   int
	Adjustment    float32
}

type Transition_probability struct {
	Id      int
	From_id int
	To_id   int
	Tp_base float32
}

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	models := []Model{
		Model{1, "HIV"},
		Model{2, "TB"}}

	people := []Person{
		Person{1},
		Person{2},
		Person{3},
		Person{4},
		Person{5},
		Person{6},
		Person{7},
		Person{8},
		Person{9},
		Person{10}}

	states := []State{
		State{1, 1, "HIV-", 0.9},
		State{2, 1, "HIV+", 0.1},
		State{3, 2, "TB-", 0.9},
		State{4, 2, "TB+", 0.1}}

	transition_probabilities := []Transition_probability{
		Transition_probability{1, 1, 1, 0.9},
		Transition_probability{2, 1, 2, 0.1},
		Transition_probability{3, 2, 1, 0},
		Transition_probability{4, 2, 2, 1},
		Transition_probability{5, 3, 3, 0.8},
		Transition_probability{6, 3, 4, 0.2},
		Transition_probability{7, 4, 3, 0},
		Transition_probability{8, 4, 4, 1}}

	interactions := []Interaction{Interaction{1, 2, 3, 4, 2}}

	cycles := []Cycle{
		Cycle{1, "2015"},
		Cycle{2, "2016"},
		Cycle{3, "2017"},
		Cycle{4, "2018"},
		Cycle{5, "2019"}}

	var master Master

	// table tests here

	// population initial cycle

	for _, model := range models {
		states := model.get_states()
		for _, person := range people {
			state := pickState(states)
			// state.add_person(person)
			person.add_state(state)
		}
	}

	for _, cycle := range cycles {
		for _, person := range people {
			shuffled := shuffle(models)
			for _, model := range shuffled {
				//if this is a model effected by an active interaction
				interactions := model.get_effected_interactions_filtered_by(person.states)
				if len(interactions) > 0 {
					for _, interaction := range interactions {
						adjustment = interaction.get_adjustment()
						transition_probability = interaction.get_transition_probability()
						adjust_transitions(adjustment, transition_probability)
						table_tests()
					}
				}

				current_state_in_this_model := person.get_state_by_model(model)
				transition_probabilities := person.get_state_by_model(model).get_transition_probabilites()

			}

		}
	}

	fmt.Println(models, people, states, transition_probabilities, interactions,
		cycles, master)

}

// func pickState(states []State) State {

// }

func pick(probabilities []float32) int {
	// iterates over array of potential states and uses a random value to find
	// where new state is. returns new state id.
	random := rand.Float32()
	sum := float32(0.0)
	for i, prob := range probabilities { //for i := 0; i < len(probabilities); i++ {
		sum += prob
		if random <= sum {
			return i
		}
	}
	// TODO(alex): figure this out - needed error of something
	return 0
}
