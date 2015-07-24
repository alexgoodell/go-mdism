Explanation of each line of code
-> cli.go -> Line 1
Open package main => standard for all go codes

Import all the relevant packages, these include functions that
somebody else already coded and are therefore ready to use.

-> line 14 -> function main
This is where the main function start. When the main function ends, the GO program exits.
16: I think this is where the app function is started. I do not really understand these lines but I do not think they are very relevant either.
=============QUESTION============================================
22: You can set 'flags', when you type those after the go-mdism (after go build) in the terminal, you can set these values (so you don't have to change the code). There are currently only 3.

30: here it says which actions the app function has to take.
31: show greeting
	-> greeting.go
	This program is really only a fancy graphic
32: print 2 things, why one with fmt and the other not? I do not understand this thing
======================QUESTION====================================

36: This is where the simulation options are listed

38: The first one is a single run, with a name and it processes the flags you put up before.
Then it startsRunWithSingle, which is another function
	




-> Line 27
28: Make a variable called begintime, which is set at the time the program commences
30: Make a new structure type, called State (global). 
This structure has 9 properties, you can access each by typing:
State.Id or State.Cost_per_Cycle

42: Make a new structure for the Models, it includes an Id (starting at 0) and the name of each model

47: Make a new structure called MasterRecord. This MasterRecord will contain all states per model per person per cycle.

54: Define a new structure for cycles. Same as for models, Id and name.

59: Define a struct for persons, only Id.

63: Define a struct for Interaction. It has 6 properties that can be accessed by the . symbol.

72: Define a structure for TransitionProbability. It has the from_id, to_id and the value of the probability.

80: Define a structure for state populations. This is similar to the MasterRecord, but it does not hold individual people, it holds accumulated total populations per state. See it as population per state per model per cycle instead of state per model per cycle per person.
I wonder if it really needs the Id? Since MAsterRecord does not have one either

88: Define the Query struct. This one is more complicated. The way I understand it, it holds all sorts of combinations. So for example the first one; state_id_by_cycle_and_person_and_model. It holds 3 slices and an integer? Or three slices of integers? So does that mean it holds a state id (integer), that is based on the cycle you select, the model and the person (slices)? I am not sure I understand this triple slice...
The others I do not fully comprehend either. Why for instance, does state_populations_by_cycle have two slices, while model_id_by_state has only one?
======================QUESTION=========================================

98: Define the input structure. This Input structure holds currentcycle, all Querydata (defined in the Query struct), all models, people, states, TPs, interactions, cycles and also all MasterRecords?
So basically this Input structure holds everything there is to know?
-> The way this works is that the Input structure holds slices of the other structures. So withint this structure are new structures. And if we look at Query, then within that second structure, it holds more slices (this goes quite deep and is hard to grasp).

110: Define the TPByRas structure: this structure holds all the properties that are necessary to define the initiation TPs for diseases determined by race, age and sex.

122: make a new variable that is called GlobalTpsByRas. It is a slice of the type TPByRas (so this slice will have all of the TPByRAS'). Since no value is given, all properties are set to their equivalent 0.

127: create the variable GlobalMasterRecords. This is also a slice. Now with this one there is an = in between. This means that we set the value for this variable at the values of the MasterRecord{}. Since we did not put anything between the curly brackets, I presume all elements of this slice of types MasterRecord are empty. What is important to understand is that MasterRecord itself does not hold slices. So to be able to keep track of all of the generated elements that have the type MasterRecord, we need to make a slice of all of them to add them to.

128: the same goes for statepopulations.

130: We make slices for these 4 variables. The make function is just anyther way of designing a slice. We use this form here because we want to define a certain length (150=max states) for the slices to keep them manageable. We could also have written 'var Global = [150]float64{}'.
Instead of a slice with a definite length, you would have an array with a fixed length, which would not make a difference.

135: Make a new variable that gives the GlobalMasterRecords by iteration-person-cycle-model? I am not sure what this variable holds. I do not understand the naming of multiple slices behind a variable.
==========================QUESTION======================================

137: create the variable output-dir (not global). This is the output directory that we want to use. It is a string variable set as "tmp". (which is the output folder)

139: create the variable numberOfPeople (not global). Of the type integer. The value is not yet set.
140: numberofIterations variable is also an integer.
141: inputsPath is set a string variable
142: isProfile is also set as a string variable (I don't know what this one is for yet).

144: create a variable called timer, which is a pointer to the function nitro.B? I do not know what this does.
===============================QUESTION================================

146: Here the main function starts.

148: Timer = nitro.Initialize. This refers to a function in the nitro package. I presume it holds the time? Or is this a function that keeps track of the time spent on each function? Need to ask.
==============================QUESTION=================================

150: Alex explained that you can use the flag package to say in the command line how much you want from something. So the function IntVar and StringVar set the values for these variables.
The & sign means that you refer to a pointer? But the variables that are referred weren't pointers, so how can we refer to them? 
================================QUESTION=========================

154: what does Flag.Parse do?
========================QUESTION=================================

156: Here we are defining whether we want to use CPU or MEM? I am not sure what this does but I presume it is used for setting up whether you want to run the model on your own CPU or if you want to run it on a server? So when we call isProfile "mem" we automatically run the model on a server?
=========================QUESTION================================

174: the runtime.GOMAXPROCS makes sure that you use all cores that you can possibly use? This goes for "mem" as well? Or is it not applicable there?
=========================QUESTION===============================

176: print how many cores are being used.
179-181: print amount of people, amount of iterations and where the inputs are gotten from.

183: Derive the variable Inputs (global), of the type Input. Remember, the type Input structure has slices of virtually all the information of the model.
What is interesting is that this seems similar to the MasterRecords global variable. There the global variable is a slice and the structure is fixed. Here the global variable is fixed and the structure conatins slices. Would these two methods be interchangeable? So say, if we would make a slice of the global variable Inputs, could we set the elements of the structure Input as fixed?
============================QUESTION============================

184: Now we are assigning values to the variable Inputs. This is done by the function initializeInputs, which takes two arguments, the variable you want to return the outcome of the function to (which is the variable Inputs), and the location where the inputs for the function are stored on the computer (which is inputsPath).

	1488: here the function initialize inputs is defined. It says it takes the arguments Inputs (the variable that is assigned), of the type Input, and the inputsPath, of the type string. When finished, it will return values of the type Input (so it will return values for all the elements of the structure Input).

	1491: The subset of the Inputs variable called CurrentCycle is set at 0.

	1496: the variable filename is made (type: string) and it is assigned the value of the location where the inputs are stored, followed by 'models' (so it is the model input file).
	1497: the numberOfRecords variable is made (type int) and is assigned the value of total number of records in the input file, according to the getNumberOfRecords function.
		1403: the getNumberOfRecords function takes a filename of the string type as its only input argument. It will return a integer.
		1404: here we start referring to functions outside of this program. I do not know what these functions do exactly, and the notation of 'csvFile, err' as a name for a variable seems strange (I thought they could not contain spaces). But what I do understand is that it reads the amount of rows that have data in them in the csv file. It subtracts 1 (the headers), to find the amount of lines of data that you want as inputs. It returns this value as an int.

	1498: back the the initialize function, we are assigning the values of Inputs.Models. We make a slice of the type Model, and give it the length of the inputs we just found.
	I do not understand why we MAKE a slice here, the Models parameter in the Input struct was already a slice right? So we could just adjust the length to the numberOfRecords?
	======================QUESTION===========================

	1499: we derive the variable ptrs, which is a slice of the type interface, and the values are not assigned.
	1500: Now, for all models (numberOfRecords), we give the value of ptrs to the slice of ptrs... I don't understand this anymore.. What is happening?
	============================QUESTION=====================

	1503: the value of ptrs is changed into what the fromCsv returns, with the arguments that go to the function: filename, inputs.Models[0], and the value of ptrs.
		1417: the function fromCsv. Takes as arguments: a filename of type string, a record with a certain interface, and the recordpointers it should create, with a certain interface. It returns a slice of interfaces.
		This whole function is beyond what I understand.. too many references to functions outside of this program that I have no idea what is going on.
		I do not even understand what it is exactly that the function returns.
		=====================QUESTION=========================

	1504: So with these values of ptrs that we got from the fromCSV function, we make pointers to location of the models. I think I only understand this half, pointers are hard to grasp.
	===========================QUESTION=======================
	1509-1589: it does exactly the same for states, TPs, interactions, cycles and TPs by Ras. So now all input variables are assigned to the global Input variable, and pointers are formed to their respective locations?

189: Make the variable numberOfPeopleEntering and give it a value. This value is not really used for a function though.

191: The Inputs global variable gets aasigned new values. This is done by the function setUpQueryData, which takes the arguments: Inputs (the variable), numberOfPeople and numberofpeopleentering.

	849: the function setUpQueryData takes the variable Inputs of the type Input, and the numbers of people starting and entering the simulation later. It returns values for Input? So it does not return values for the globalvariable Inputs? It gives values for the structure Input?

	853: create the variable for total people, which is the starting people + the people entering.
	854: Print the total amount of people.

	856: Within the global variable Inputs (of struct Input), within the variable Querydata (of struct Query), within the variable state_id_by_cycle_and_person_and_model, we want to make a slice (I thought it was already a slice), and we define the length of this slice being the total length of cycles, plus 1. What is happening here and why is it plus 1?
	===================QUESTION================================

	857: For the full range of this slice of cycles, we want each element to contain a slice of the length of the number of people.

	860: For the full range of the slice of cycles, and the full length of the slice of people, we want each element to contain a slice of the length of the number of models.

	So what we are doing here is creating a slice within a slice within a slice within a structure within another structure that is the type of a global variable, called Inputs? Haha this goes a little deep but I think I sort of understand. I do not yet understand the point of this.

	868: It goes on for all the other Querydata struct Query elements. But I do not really understand this. First I need to comprehend how the triple slice inside the struct Query works.
	=======================QUESTION=========================

195: assign values to the Inputs global variable according to the function createInitialPeople. This function takes the Inputs variable and numberOfPeople as arguments.
	1017: function create initialpeople, takes Inputs variable of the type Input, and a number of the type int as inputs and will return something of the type Input (struct).
	1018: it makes a slice with the length of the total number of people you want in the initiation. The slice is in location Inputs.People.
	1022: For the whole range of these Inputs.People, and within that, for the whole range of Inputs.Models, we want to get the uninitialized state per model. We do that with the function model.get_uninitialized_state, which takes the argument &Inputs, which is the location where Inputs are stored.
		1240: Function (model * Model) get_uninitialized_state takes Inputs as a pointer to Input type (?!) as arguments, and gives the State for the specific model.
		1241: Create the variable modelId which is the Id of the specific model, that you can find by model.id (which refers to the Model pointer), it is of type int.
		1242: For all states, we want to check if the model id is equal to the model we are assessing now, and if it is the uninitialized state. If it is, return the state of the type State (so that contains all 9 elements).
		1247: If it cannot find the state, it exits the program and says that it cannot find it.

	1025: create variable mr of the Type MasterRecord.
	1026: Assign the values of this variable according to the results you just got and the model you are in and the person you are handeling right now.
	1032: create the variable qd which has the value of a state that is specific for the cycle and the person and the model we are currently processing.
	1034: Value of qd state is assigned as the value of the state we just found.
	I do not understand what the variable qd is for? Since it is not used in any formula. It just gets derived with a certain value, then changed, and then never used again.
	======================QUESTION===========================

	1037: Add the states you just found to the Inputs.MasterRecords and to the global variable GlobalMasterRecords.

	1051: timer.step? This is probably to keep track of processing time. Also a function outside of this code.

	1053: this is where the Inputs get returned to the function.

197: The value of the Inputs variable get changed to set up the starting values of the GlobalStatePopulation with the function initializeGlobalStatePopulations. This function only takes Inputs as argument.
	932: The function initializeGlobalStatePopulations takes Inputs of the type Input as input, and returns the something of the type Input.
	936: create the variable for number of calculated cycles., which is the length of cycles +1 (why plus 1?)
	=====================QUESTION=============================
	937: The global variable GlobalStatePopulations gets assigned a slice with types StatePopulation, that has the length of the cycles times the number of states.
	938: set variable q to 0 (type int)
	939: for each of the cycles (+1), and within each cycle, for each of the states, set the value for each element of the GlobalStatePopulations to their respective starting value.
	949: Return the Inputs. I do not really understand why we have to return anything? It does not seem like we change the Inputs, right?
	======================QUESTION=============================

199: Execute function setUpGlobalMasterRecordsByIPCM, which takes as its only argument the Inputs variable.
	952: Function setUpGlobalRecordsByIPCM takes Inputs of Type input as its only argument, and does not return anything.
	954: This functions i making slices within slices again, a concept I do not fully comprehend yet.
	I also am not sure yet of the use of this IPCM at all.
	=========================QUESTION===========================
	965: Another step in the timer.

201: Make the variable that says whether we are handling an intervention right now. Set it at false to say we are not.

203: Make the variable interventionAsInteraction, which is an interaction. Shouldn't we put a number in here for which interaction we want to change? Besides, we don't change the interaction right? We change the baseline TP to go from uninitialized to high fructose group (added risk).
The way it is coded now is that we make an extra interaction? Where we multiply that baseline TP by 0.8.
===================QUESTION====================================

205: The factor of adjustment.
206: the from state id
207: the to state id of the interaction

209: make a new variable called cycle, of which we set the value at Cycle{}. Because there is nothing between the curly brackets, I am not sure what the value is exactly?
==================QUESTION=====================================

211: make the new variable newTps which will be a slice of the type TransitionProbability.

213: If there is an intervention, 
214: the variable unitFructoseState gets assigned all the State values of state 37 through the function get_state_by_id.
	1058: Function get_state_by_id takes localInputs of type *Input as inputs and stateId as an int. It returns a type State. 
	1060: Create the variable theState which has the contents of the specified stateId.
	1062: if the Id's are the same, return that state.
	1066: if the Ids are not the same, give an error and exit
	(still return the State).

215: tPs gets assigned all the values of the destination probabilities of this state according to the function get_destination_probabilities, for the unitFructoseState. It takes 
&Inputs as argument
	1256: Function (state *State) get_destination_probabilities gets all destination TPs (which he returns in a slice) related to the specified state. It further takes localinputs as argument with type *Inputs. 
	I do not understand why we use localInputs here? What does that mean?
	======================QUESTION============================
	1257: create tPIdsToReturn which is a slice of integers.
	1258: the value of this variable is the values of all TPs from this particular state that we get from localInputs.Querydata.Tp_ids_by_from_state.
	1259: variable TpstoReturn is made as a slice of types TransitionProbability, with the length of the amount of Tps we just found. (where you can go from that state).
	1260: For the range of all possible Tps, set the TpsToReturn as the TransitionPossibilities specific to this id.
	1263: If there are TpsToReturn, return them
	1266: If there aren't any destination Tps to return, give error and exit (still return them as well).
216: newTps gets the recalculated value of transitioning when the intervention is on, with the function adjust_transitions, which takes the Inputs pointer, the Tps, the interventionInteraction and the cycle as arguments.
	1083: The function adjust_transitions takes localInputs as type Input pointer as inputs, the Tps that you want to adjust, which is a slice of the type TransitionProbability, an interaction, which is of the type Interaction, and a cycle which is of the type Cycle. It returns a slice of TransitionProbabilities.
	1086: The adjustmentFactor variable is assigned the value of the Interaction.Adjustment
	1093: hasTimeEffect variable gets created. It gets a boolean value (true/false). It is true if the interaction.to_state_id is 13, 14 or 8 (CHD inc, mort and natural mort).
	1094: if the cycle.id is bigger than 1 and there is a time effect, make a slice of type float 64 of length 15 and set element 8, 13 and 14 to their respective values.
	The new adjustmentfactor = the old one * time effect^cycle.id-2
	Now, I am not sure if that if cycle.Id > 1 is correct, and neither am I sure about the -2 for the cycle Id's. We should check if it really functions properly like this.
	Can we build a function where these TPs get printed for each cycle so we can check?
	==================QUESTION==============================

	1104: for the range of all the Tps you said you wanted to have adjusted (the full slice): 
	1107: tp variable value is the specific TP
	1108: original TpBase is created and gets the original value.
	1109: If the from_id is equal to the from_id of the interaction and the to_id is equal to the to_id of the interaction, the new Tp_base will be the old Tp_base * the adjustmentFactor.
	1111: Check: if the new Tp_base is equal to the original Tp_base but the adjustmentfactor was not 1, than something went wrong and it should give an error with the specific interaction, and exit the program.

	1121: Get the sum of all the Tps after adjustment, by using the get_sum function, which takes the slice of TheseTPs as an argument.
		1160: the get_sum function takes the slice of the TPs you want to have checked (types TransitionProbabilities) and returns a float64.
		1161: Derive the variable sum as type float64 and give it value 0.0
		1162: For the full range of TheseTPs, add each to the previous one to get the total sum.
		1165: return the sum to the function that called the get_sum function.
	1122: now derive the variable remain, which is the sum minus 1.0 (so now we will have only the adjustment that we calculated, which can either be - or + something).

	1124: Make the variable recursiveTp of type float64.
	1126: For the range of the full slice of TheseTPs, the tp is assigned the value of the respective TheseTPs we are treating.

	1129: if the tp.From_id is equal to the To_id; meaning that the tp is the one considering the people that stay in the same state, adjust the Tp_base by subtracting the remaining value from the original value of Tp_base.
	1131: the value of recursiveTp gets set as the value for Tp.Tp_base (necessary for check see hereafter)
	1141: the variable model is set as the model that is affected by this interaction.
	1142: the variable unitState is set as the uninitialized state within this model by using the function get_uninitialized_state, that takes the localInputs as an argument.
	1143: if the from_Id (state) within the first TP within theseTPs is equal to the uninitialized state and the recursiveTp (the adjusted Tp_base) is not equal to 0, give an error and exit the program. This function confirms that the adjustments to the TPs cannot be done for the ones in the uninitialized state, since there should not be an interaction there.
	1148: return the newly derived theseTPs to the function that called for the adjustment.
219: for each newTps, set the transitionprobabilities within Inputs according to the newly found value.
This part from 209 until 221 makes sure that not only the TP of the intervention is changed, but also the TPs of the other states that the from state of this TP goes to. Else they wouldn't add up to 1 anymore.
This method is nice because it makes sure that it doesn't matter which value you enter for the intervention. In our case, it would probably be faster to just change the two relevant TPs to the values we have in mind directly, without interaction or adjusting the others according to these functions.

225: The variable concurrencyBy by gets assigned the string value "person-within-cycle"

227: we are making a new variable which is a channel of strings?
I have a vague idea, but am not sure yet how channels exactly work, so need some more background here.
=========================QUESTION=============================

229: For the number of iterations we put in, go for the function runModel, which takes the arguments Inputs, concorrencyBy and iterationChan.
		
	242: function runModel takes the arguments Inputs of type Input, the concurrencyBy of type string and the iterationChan of type channel of strings.
	244: print that initialization is complete and gives the elapsed time.
	245: set a new value for begintime.
	247: make a new channel with slices of type MasterRecord. we call the channel masterRecordsToAdd.

	250: make variable localInputs of the type Input
	251: set the value of localInputs as the deepCopy of Inputs. Deepcopy is a function that takes Inputs as an argument.
	I do not understand why it is only here that you make the variable localinputs, while you already have used it before.
	=======================QUESTION=============================
		383: function deepCopy takes Inputs of type Input as arguments and will return something of type Input.
		385: I have no idea what is happening in this function.. all these functions refer to packages outside of this program and I have no clue what they do.
		==========================QUESTION======================

	253: here a switch is built into the program. When concurrencyBy is set as person, it will run the first model. If ConcurrencyBy is set as "person-within-cycle", it will run the second model. Since we use the second model all the time, I am going to skip the first one.

	255-269: Not relevant for our current model, is another way of doing the concurrency.

	271: if our concurrencyBy is "person-within-cycle", we run the following:
	273: for each cycle within the range of cycles
	277: if cycle Id is bigger than 0, perform the function createNewPeople. The function takes the pointer to localInputs, the cycle and a number, which is the number of people you want to have entering. It is set now at cycle>0 because we do not yet want new people to enter the model at cycle 0.
		981: The function createNewPeople takes Inputs of type *Input (Input pointer), a cycle of the type Cycle and a number of the type int as arguments and it does not return anything.
		982: the variable idForFirstPerson is set at 1 more than the amount of people we already had.
		984: for each new person, set the Person value at 1 more than the id for the first person.
		987: Make the slice of people in the Inputs bigger for each person
		989: for each model within Inputs:
		992: set the new variable uninitialized state as a type state with the value of the uninitialized state specific for this model
		994: if the model name is age, the value for uninitializedstate will be 42 (instead of the 41 that is actually the uninitialized state).
		This is because we want them to enter at age 20 instead of in the uninitialized state? But the problem here is of course that when they are 20 in cycle 0, that they will be 21 when they really enter the model. But we do not want them to really be in the uninitialized state, because then they would be distributed among all ages. We want all of them to be 20. So how are we going to code this?
		====================QUESTION===========================
		1000: create the variable mr of the type MasterRecord
		1001: set the values for each of the elements of mr
		1006: Again the variable qd which I don't understand.
		=======================QUESTION=======================
		1010: Add the new values to the Inputs.MasterRecords

	281: for each person within the full range of people in Inputs, run concurrently the function runOneCycleForOnePerson. The function takes the localInputs, the cycle, the person and the masterRecordsToAdd as arguments.
		813: The function runOneCycleForOnePerson takes the localInputs of type pointer to Input, a cycle of type Cycle, a person of type Person and a masterRecordsToAdd of type a channel of slices of MasterRecord as arguments. It returns nothing.
		815: we create the variable localInputsPointer with the values of localInputs
		Why do we do this? Can't we just use localInputs?
		=======================QUESTION========================

		817: the new variable mrSize has the value of the length of the amount of models.
		818: the new variable theseMasterRecordsToAdd is a slice of the type MasterRecord with the length of the models.
		812: mrIndex is set as 0
		820: for each model in the range of models, run the function runCyclePersonModel, which takes the localInputsPointer, the cycle, the modelm the person, the pointer to theseMasterRecordsToAdd and mrIndex as arguments.
			403: the function runCyclePersonModel takes the pointer to localInputsPointer, a cycle of type Cycle, a model of type Model, a person of type Person, mrIndex of type int and theseMasterRecordsToAddPtr of type a pointer to a slice of type MasterRecord.
			407: get the state in the current model with the function get_state_by_model, which takes the localinputsPointer and the model as arguments.
				1187: the function get_state_by_model goes for a specific person and takes the localInputs as an argument of the type pointer to Input, and wants a model of the type Model as an argument. It returns something of type State.
				1188: the variable thisModelId gets the value of the id of the model that you want the state from.
				1189: the new variable stateToReturn is of the type State
				1190: the new variable stateToReturnId is of the type integer
				1192: the value of statetoreturnId is the value we find in Inputs.Querydata under the right cycle, model and person.
				1194: this does not do anything, can be deleted?
				==================QUESTION=====================

				1198: stateToReturn is the state that has the Id we just found that was specific for model, cycle and person.
				1199: if the Ids match, return that State.
				1202: if the ids don't match give an error and exit, but still return the state.
			410: get all the destinationprobabilities for the state you just found the person was in for this model and this cycle. get_destination_probabilities function is already explained.

			413: the new variable states gets the value of the states found by the function get_states that takes the localInputsPointer as an argument.
				1208: function get_states gets all the states a specific person is occupying in the current cycle. It takes a pointer to the localInputs as argument and returns a slice of States.
				1213: the variable statesToReturnIds gets the value of all state Ids for this person in this cycle. It is not specified but I presume this is a slice?
				===============QUESTION======================
				1215: the states to return variable is made as a slice of type States with a length of all the states we just found the person occupies.
				1217: for the full range of the state the person occupies, if the Id of a state in localInputs matches the Id of that specific state, add that state to the slice of statesToReturn.
				1220: if the Id's do not match, give an error with that specific state that fails and exit the program.
				1227: If the length of the states to return is bigger than 0 (there is something to return), return that slice of states.
				1229: if there are no states to return, give an error and exit the program, but still return the statesToReturn.

			420: if the currentstateinthis model in this cycle for this specific person are the uninitialized states for CHD, T2D and BMI, then change the transitionprobabilities according to the function getTransitionProbabilitiesByRas, which takes the localInputsPointer, currentStateInThisModel, states and person as arguments.
			-> I was wondering, shouldn't this be in the main function? Or at least earlier? Because this is something that should be in the initialization phase right? Where everyone's states are assigned for year 0?
			===============QUESTION=======================
				1591: function get TransitionProbByRas takes the localInputsPointer of type a pointer to Input, the currentstate of type State, the states of type a slice of States, and the person of type Person as arguments. It returns a slice of transitionprobabilities.
				1593: the variable tpsToReturn is a slice of the type TransitionProbability
				1595-1605: getting all the Ids of the states this person occupies within the other models.
				1607: for the full range of the GlobalTPsByRas, if the id's of the current person matches the model and state within the other models, set the newTp.Tp_base as the probability you found in tpByRas.
				1613: then make a slice of Tps of all of them.
				1617: if if does not find anything to return, give error and exit.
				1622: return the slice of tpsToReturn.
			423: check the sum of the found transitionprobabilities by the function check_sum, that takes the tPs as it's only argument
				1151: function check_sum takes theseTPs as an argument, which is of the type a slice of type TransitionProbability.
				1152: sum is the sum of those TPs with the get_sum formula I already explained.
				1154: if the sum is not equal to 10, give error and exit.
			429: make a new interactions variable that contains all the relevant interactions via the functions get_relevant_interactions.
				1276: the function get_relevant_interactions is specific for a certain inState (with a State pointer) and uses localInputs *Input and allstates of type slice of State as arguments. It returns a slice of types Interaction.
				1277: variable modelId is the modelId of the inState you put in the arguments.
				1279: relevantInteractions is the variable that holds the slice of types Interaction.
				1280: for the full range of all the States you want the interactions for, it gets the relevant interactionId. 
				1282: when there is a relevantInteractionId and it is equal to the localInputs.Interactions.Id, then add this interaction to the slice of relevantInteractions.
				1286: If it is not the same as the Id of localInputs, give error.
				1291: return the slice of Interactions from all the states this person occupies. (but you only do it for the current state in this model in this cycle for this person, so you will only get the interactions for that state).

			435: If there are interactions, adjust the transitionProbabilities for the full range of interactions. 
			445: Check the sum again.
			The only thing I can think of that might be going wrong with these interactions is that there is a pointer wrong somewhere? And that because of that he does not actually change the underlying value but only the value within the secondary function?
			But if you change the underlying values, how does this influence the TPs in the next cycles?
			========================QUESTION=================
			448: new_state variable is the state that you pick with the function pickstate. The function takes the localInputsPointer and the transitionProbabilities as arguments.
				1318: function pickstate. Takes a pointer to localInputs and a slice of TransitionProbabilities as arguments. Returns something of type State.
				1319: make a slice of float64 types and give it the length of the amount ot TPs you want to incorporate.
				1320: for the full range of TPs, fill the full slice of probs with the values of Tp_base (so here we convert the transitionP struct into only a value for transitionprobability).
				1324: chosenIndex variable is determined by the function pick, which takes the slice of probs as an argument.
					1345: function pick takes a slice of probabilities of type float64 as argument and returns an int.
					1346: random number is derived by the rand function. How many decimals are used? If that is low, we will not find results for the rare incidences.
					==================QUESTION==============
					1347: sum variable is set at 0.0 as a float 64.
					1348: for the whole range of probabilities, you want to add the probs (TP_base) to find a sum. But before you add each next one, you want to check if the random is smaller or equal to the sum. If so, return the id of the probability that is chosen.
					1355: error if nothing gets returned -> this error is not yet coded completely.
				1325: stateId to transfer to is the one that is chosen by the pick function.
				1326: if the stateId that is chosen is 0, an error has occured and the program needs to exit.
				1331: find the state that belongs to the id we just found. (this variable is of type State)
				1333: if the pointer to this State is not equal to 0, return the State. I do not understand this? But that is probably because I do not really understand pointers. Why should it not be 0?
				=====================QUESTION==================
				1336: if it is not equal to zero, give error and exit, but still return the State.
			453: calculate the discountvalue based on the current cycle we are in.
			455: define the costs for each state and put them in a slice. Because the costs for the states are also in the input document for State structure, we should not have to define these here again do we?
			=========================QUESTION===================
			456-463: fill the slice with values specific for the states.
			465: if the cycle Id is bigger than 0, add the costs for the newly chosen state * the discountrate to the GlobalCostsByState variable.
			Is 0 the correct number here?
			======================QUESTION======================
			468: the variable statespecificYLDs holds the disability weight for the current state. (we can do this for the costs as well right?)
			======================QUESTION======================
			469: if the disabilityweight is not a number, give an error, print which ones we are talking about, and exit. This error is not necessary anymore since I changed the way we calculate YLD's (no more discounting).
			=======================QUESTION=====================
			475: Add the YLDs for this current state to the GlobalYLDsByState.
			476: Add this value also to the GlobalDALYsByState.
			480: if the new found state is diseasespecific death and the previous state (last cycle) was not, say justDiedOfDiseaseSpecific is true.
			482: Same thing for naturalcauses.
			484: If he died a diseasespecificdeath, get the YLL for this person because of this death and add them to the GlobalYLLsByState and to the GlobalDALYsByState. This is done with the getYLLFromDeath function, which takes the localInputsPointer and the person as arguments.
				565: the function getYLLFromDeath takes the localinputsPointer and a person of type Person as arguments and returns a float64.
				569: the variable agesmodel gets the values of the model type for age.
				570: the stateinAge model is found by the function get_state_by_model.
				572: the real age of this person is found by getting the Id of the State this person is in in the age model and subtracting 22.
				574: We make slices of the HALE for men and women and fill them up with data.
				761: get the model for sex.
				762: get the state in that model.
				763: get the Id of that state.
				764: Make the variable we want to return and set it at 0.00 for now.
				766: if the Id of the state in the model for sex is 34 (male), get the lifeexpectancy of this person from the MALE table. If it is 35 (female) get it from the female table.
				If it is neither, give an error because no sex is found and exit the program.
				Return the LifeExpectancy for this person.
			491: If he died a diseasespecific death or a natural death, go over the full range of models and do the following:
			501: sub_model.Id? I do not get this? why is this in there?
			======================QUESTION======================
			503: otherDeathState variable is made and it is gotten by the function getOtherDeathStateByModel, which takes the localInputsPointer and sub_model as arguments.
				779: function getOtherDeathStateByModel takes the pointer to localInputsPointer and a model of the type Model as inputs, and returns a State.
				780: otherDeathStateId is the Id we find in the QueryData.
				781: otherDeathState is the state that belong sto that Id, we find that via the get_state_by_id function.
				782: return the otherDeathState.
			508: the variable prev_cycle is of the struct cycle?
			509: the variable prev.cycle.Id is cycle.Id minus 1.
			510: Now we add this to the QueryDataMasterRecord by the function addToQueryDataMasterRecord.
			I am completely lost here, the QueryData within the Inputs and the Query struct do not even contain the item MasterRecord? I get that this function is supposed to put everyone in the otherdeath state within the other models, but I do not understand how he does that.
			=========================QUESTION==================
				1301: function addToQueryDataMasterRecord takes the localInputsPointer, a cycle of type Cycle, a person of type Person and a newState of type State as arguments. It returns something of type boolean.
				1302: the variable ogLen gets the number (integer) of the length of the localInputs.MasterRecords
				1303: we make a new variable called newMasterRecord of the type MasterRecord.
				1304: We assign all the attributes of this newMasterRecord.
				1309: we change the localInputs.QueryData? 
				1311: _ ogLen? This does not make sense.
				1313: return false? 
				I really don't get this part, what is the point of all this?
				===================QUESTION=====================
			518: Now a whole set of error tests are here. first one gives an error if new_state is other death but currentstateinthismodel is not other death.
			523: second one says that there should be a new_state.id bigger than 0.
			528: third one says that the current cycle should be the same as cycle Id.

			534: make the variable err and add stuff to the QueryDataMasterRecord. This is a similar thing that I do not understand.
			=====================QUESTION=======================

			536: the id of the new state of the next cycle is searched.
			Why of the next cycle? It should be the currentcycle.id of the next cycle shouldn't it?
			======================QUESTION======================
			538: if the new state of the next cycle is not the same as the new state we just assigned, give an error.
			543: If error is not false (the return of the addToQueryDataMasterRecord), give an error.

			End of the function runCyclePersonModel (did all models, see line 820)

		826: for each model in the range of models, update the values within newMasterRecord. 
		832: set the values of theseMasterRecordsToAdd for each of the models to the values we have assigned to the newMasterRecord.

		836: channel all theseMasterRecordsToAdd into the primary channel; masterRecordsToAdd.

		End function runOneCycleForOnePerson (did all models for this person within this cycle and added them to the channel that will add to the masterrecord, and did that for the full range of people (see line 281)).

	285: For the full range of all people, channel the derived masterRecordsToAdd into the newly made variable mRstoAdd. 
	Then append the GlobalMasterRecords slice with all the data that was in that channel.

	289-292: This does not seem to be doing anything, delete?
	========================QUESTION============================

	293 & 294: need to add these 2 because else the for loop would give a problem. But it is also possible to replace the 'person' in the for statement with a _ right? and would cycle really give a warning? Since you do you cycle in your if statement right?
	========================QUESTION============================

	297: go to the next cycle here, to repeat the process for each person and each model.

	315: All calculations are done. Print the time that expired since the calculations started (after initiation).

	317: For the full range of GlobalMasterRecords, if preson.Id is bigger than 100? This seems like a nonsense statement.
	Can be deleted?
	========================QUESTION============================
	323: add a person to each state_population by cycle if the masterRecord.Person_id is bigger than 100? This makes no sense. What is going on?
	========================Question============================
	326: for the full range of GlobalStatePopulations, the number of the statepopulations.Population is that what is stored in Inputs.QueryData.State_Populations_By_cycle -> this is where the Global variable of the state populations gets filled.

	333-346: for the first 150 states (all of them), print the state specific YLD's, YLL's, Costs and DALYs.

	374: write the state populations to an output document.
	I am wondering, since we don't really use the masterrecords as an ouput document? Can't we trim it down? Since the state populations are derived from the QueryData, do we even need the MasterRecords?
	======================QUESTION=============================

	377: print the total time elapsed (excluding initialization though). Maybe we want to change that? Since you reset the begintime variable now, maybe we shouldn't do that, but make a second one, so we can also see the total time elapsed instead of the time excluding initiation?
	======================QUESTION=============================

	379: Says that the channel is done? I am not sure what this does. Does it close the channel otherwise the program can't exit? I am not sure.
	======================QUESTION=============================

	381: end of function. The outputs are completely written, and now we move to the next iteration (see line 229) until all iterations are done.

233: So now everything is completely done, all outputs are written for all iterations. We want the variable ToPrint to be changed according to the iterationChan? I do not understand what this means.
=======================QUESTION===============================

235: print the variable toPrint, I know it prints "done" every time an iteration is complete, I just don't really understand how, see above.

238: Reset the timer? Or something, I am not sure, why do we set a timer step as "main"?
=======================QUESTION================================

240: Finishing the main function and therefore, the complete program.

So as I understand it, the models are run in order every time? How does that work with moving people to the otherdeath state in the other models and does it count the costs and DALYs before moving them? Even when they die in a later model when the first model is already simulated?
===========================QUESTION====================================