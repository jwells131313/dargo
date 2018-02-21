package utilities

import (
    "testing"
    "reflect"
)

type iFace3 interface {}

func TestConstantDescriptor(t *testing.T) {
	i1 := new(iFace3)
	
	cDesc, err := NewConstantDescriptor(i1)
	if err != nil {
		t.Errorf("Could not create a new constant descriptor %v", err)
	}
	
	var contracts []reflect.Type
	contracts = []reflect.Type { reflect.TypeOf(i1) }
	
	cDesc.SetAdvertisedInterfaces(contracts)
	
	i2, err2 := cDesc.Create()
	if err2 != nil {
		t.Errorf("Could not create the thing from the constant descriptor %v", err)
	}
	
	if i1 != i2 {
		t.Errorf("Did not get my original constant back! %v/%v", i1, i2)
	}
}

