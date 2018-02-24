package utilities

import (
    "testing"
    "reflect"
    "../internal"
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
	
	fCreate := cDesc.GetCreateFunction()
	
	i2, err2 := fCreate()
	if err2 != nil {
		t.Errorf("Could not call create method from descriptor %v", err2)
	}
	
	if i1 != i2 {
		t.Errorf("Did not get my original constant back! %v/%v", i1, i2)
	}
}

func TestSystemDescriptor(t *testing.T) {
	i1 := new(iFace3)
	
	cDesc, err := NewConstantDescriptor(i1)
	if err != nil {
		t.Errorf("Could not create a new constant descriptor %v", err)
	}
	
	var contracts []reflect.Type
	contracts = []reflect.Type { reflect.TypeOf(i1) }
	
	cDesc.SetAdvertisedInterfaces(contracts)
	
	sDesc := internal.CopyDescriptor(cDesc)
	
	fCreate := sDesc.GetCreateFunction()
	
	i2, err2 := fCreate()
	if err2 != nil {
		t.Errorf("Could not call create method from descriptor %v", err2)
	}
	
	if i1 != i2 {
		t.Errorf("Did not get my original constant back! %v/%v", i1, i2)
	}
}

