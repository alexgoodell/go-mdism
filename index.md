
get states, models, interactions, people, cycles from the database

set the initial population within each state by transitioning from an uninitialized state in each model into their starting state

now, we will do the same for cycles that are not zero

foreach cycle
	foreach person
		randomize the order of the models
		foreach model
			get the current state of the person in this model (should be the uninitialized state for cycle 0)
			get the transition probabilities from the given state

			get any interactions that will effect this model
			if there are interactions
				foreach interaction
					apply the interactions to the transition probabilities
				end foreach interaction
			end if there are interactions

			using  final transition probabilities, assign new state to person
			store new state in master object 
		end foreach model
	end foreach person
end foreach cycle


	
functions needed







deal with ORs


