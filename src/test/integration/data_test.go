package main

import (
	"fmt"
	"testing"
	dataclient "extremeWorkload.com/daytrader/lib/data"
	modelsdata "extremeWorkload.com/daytrader/lib/models/data"
)

func TestUsers(t *testing.T) {
	var dataClient = dataclient.DataClient{}

	//Creating and reading
	newUser := modelsdata.User{"1234", 68, []modelsdata.Investment{}}
	dataClient.CreateUser(newUser)

	users, readAllErr := dataClient.ReadUsers()
	if readAllErr != nil {
		t.Errorf("there was an error while reading all users")
		fmt.Println(readAllErr)
	}
	fmt.Println(users)

	user, readErr := dataClient.ReadUser("1234")
	if readErr != nil {
		t.Errorf("there was an error while reading a specific user")
		fmt.Println(readAllErr)
	}
	fmt.Println(user)

	//Updating
	var investments []modelsdata.Investment
	investments = append(investments, modelsdata.Investment{"TTT", 96});
	updateUser := modelsdata.User{"1234", 1000, investments}

	updateErr := dataClient.UpdateUser(updateUser)
	if updateErr != nil {
		t.Errorf("there was an error while updating the user")
		fmt.Println(updateErr);
	}
	
	usersAfterUpdate, updateReadAllErr := dataClient.ReadUsers();
	if updateReadAllErr != nil {
		t.Errorf("there was an error while reading all users after one has been updated")
		fmt.Println(updateReadAllErr)
	}
	fmt.Println(usersAfterUpdate)

	//Deleting
	deleteErr := dataClient.DeleteUser("1234");
	if deleteErr != nil {
		t.Errorf("there was an error while deleting a user");
		fmt.Println(deleteErr)
	}

	usersAfterDelete, deleteReadAllErr := dataClient.ReadUsers()
	if deleteReadAllErr != nil {
		t.Errorf("there was an error while reading all users after one has been deleted")
		fmt.Println(deleteReadAllErr);
	}
	fmt.Println(usersAfterDelete)
}

func TestTriggers(t *testing.T) {
	var dataClient = dataclient.DataClient{}

	//Creating and reading
	newTrigger := modelsdata.Trigger{"1234", "ABC", 100, 100, false}
	createErr := dataClient.CreateTrigger(newTrigger)
	if createErr != nil {
		t.Errorf("There was an error when creating a trigger");
		fmt.Println(createErr);
	}

	triggers, readAllErr := dataClient.ReadTriggers()
	if readAllErr != nil {
		t.Errorf("There was an error when reading all triggers");

		fmt.Println(readAllErr)
	}
	fmt.Println(triggers)

	trigger, readErr := dataClient.ReadTrigger("1234", "ABC")
	if readErr != nil {
		t.Errorf("There was an error when reading a specific trigger");
		fmt.Println(readErr)
	}
	fmt.Println(trigger)

	//Updating
	updateTrigger := modelsdata.Trigger{"1234", "ABC", 200, 200, true}
	updateErr := dataClient.UpdateTrigger(updateTrigger);
	if updateErr != nil {
		t.Errorf("There was an error when updating a trigger");
		fmt.Println(updateErr);
	}

	triggersAfterUpdate, updateReadAllErr := dataClient.ReadTriggers()
	if updateReadAllErr != nil {
		t.Errorf("There was an error when reading triggers after updating")
		fmt.Println(updateReadAllErr)
	}
	fmt.Println(triggersAfterUpdate)

	//Deleting
	deleteErr := dataClient.DeleteTrigger("1234", "ABC")
	if deleteErr != nil {
		t.Errorf("There was an error when deleting a trigger");
	}

	triggersAfterDelete, deleteReadAllErr := dataClient.ReadTriggers();
	if deleteReadAllErr != nil {
		t.Errorf("There was an error reading triggers after deleting one")
	}
	fmt.Println(triggersAfterDelete)
}