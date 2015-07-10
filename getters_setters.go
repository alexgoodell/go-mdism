package main

import (
	"fmt"
	"os"
)

//  --------------- model

// gets the uninitialized state for a model (the state individuals start in)
func (model *Model) get_uninitialized_state() State {
	stateId := Query.Unintialized_state_by_model[model.Id]
	state := Inputs.States[stateId]
	return state
}

//  --------------- state

// get the transition probabilities *from* the given state. It's called
// destination because we're finding the chances of moving to each destination
func (state *State) get_destination_probabilites() []TransitionProbability {
	var tPIdsToReturn []int
	tPIdsToReturn = Query.Tps_id_by_from_state[state.Id]
	tPsToReturn := make([]TransitionProbability, len(tPIdsToReturn), len(tPIdsToReturn))
	for i, id := range tPIdsToReturn {
		tPsToReturn[i] = Inputs.TransitionProbabilities[id]
	}
	if len(tPsToReturn) > 0 {
		return tPsToReturn
	} else {
		fmt.Println("cannot find destination probabilities via get_destination_probabilites")
		os.Exit(1)
		return tPsToReturn
	}
}

// get any interactions that will effect the transtion from
// the persons current states based on all states that they are
// in - it is a method of their current state in this model,
// and accepts an array of all currents states they occupy
func (fromState State) get_relevant_interactions(allStates []State) []Interaction {

	var relevantInteractions []Interaction
	var interactionIds []int
	for _, inState := range allStates {
		interactionIds = append(interactionIds, Query.getInteractionIds(inState, fromState)...)
	}
	for _, interactionId := range interactionIds {
		relevantInteractions = append(relevantInteractions, Inputs.Interactions[interactionId])
	}

	var toState []int
	if len(relevantInteractions) > 0 {
		for relevantInteractionId := range relevantInteractions {
			toState = append(toState, Inputs.Interactions[relevantInteractionId].To_state_id)
		}
		fmt.Println(toState)
		pause()
	}
	_ = toState

	// :i is faster than append()
	return relevantInteractions

}

func get_state_by_id(stateId int) State {

	theState := Inputs.States[stateId]

	if theState.Id == stateId {
		return theState
	}

	fmt.Println("Cannot find state by id ", stateId)
	os.Exit(1)
	return theState

}

func getYLLFromDeath(person Person, cycle Cycle) float64 {
	agesModel := Query.getModelByName("Age")
	ageState := person.get_state_by_model(agesModel, cycle)
	sexModel := Query.getModelByName("Sex")
	sexState := person.get_state_by_model(sexModel, cycle)
	return Query.getLifeExpectancyBySexAge(sexState, ageState)
}

func getOtherDeathStateByModel(model Model) State {
	otherDeathStateId := Query.Other_death_state_by_model[model.Id]
	otherDeathState := get_state_by_id(otherDeathStateId)
	return otherDeathState
}

func (Query *Query_t) getModelByName(name string) Model {
	modelId := Query.model_id_by_name[name]
	model := Inputs.Models[modelId]
	if model.Name != name {
		fmt.Println("problem getting model by name: ", name, " does not exist")
		os.Exit(1)
	}
	return model
}

func (Query *Query_t) getStateByName(name string) State {
	stateId := Query.state_id_by_name[name]
	state := Inputs.States[stateId]
	if state.Name != name {
		fmt.Println("problem getting state by name: ", name, " does not exist")
		os.Exit(1)
	}
	return state
}

func (Query *Query_t) getLifeExpectancyBySexAge(sex State, age State) float64 {
	//Use struct as map key
	key := SexAge{sex.Id, age.Id}
	le := Query.Life_expectancy_by_sex_and_age[key]
	return le
}

func (Query *Query_t) getInteractionIds(inState State, fromState State) []int {
	//Use struct as map key
	var key InteractionKey
	var interactionIdsToReturn []int
	key.In_state_id = inState.Id
	key.From_state_id = fromState.Id
	interactionIds := Query.interaction_id_by_in_state_and_from_state[key]
	for _, interactionId := range interactionIds {
		interaction := &Inputs.Interactions[interactionId]
		if interaction.From_state_id == fromState.Id && interaction.In_state_id == inState.Id {
			interactionIdsToReturn = append(interactionIdsToReturn, interaction.Id)
		}
	}

	// if len(interactionIdsToReturn) > 0 {
	// 	//fmt.Println(interactionIdsToReturn[0:1])
	// 	return interactionIdsToReturn[0:1]
	// }

	return interactionIdsToReturn
}

func (Query *Query_t) getTpByRAS(raceState State, ageState State, sexState State, model Model) []TPByRAS {
	var key RASkey
	key.Age_state_id = ageState.Id
	key.Race_state_id = raceState.Id
	key.Sex_state_id = sexState.Id
	key.Model_id = model.Id
	RASs := Query.TP_by_RAS[key]

	// if ras.Model_id != model.Id || ras.Age+22 != ageState.Id || ras.Race_state_id != raceState.Id || ras.Sex_state_id != sexState.Id {
	// 	fmt.Println("cannot find by RAS")
	// 	os.Exit(1)
	// }
	return RASs
}
