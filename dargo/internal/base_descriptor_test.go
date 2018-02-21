package internal

import (
    "testing"
    "strings"
    "../api"
    "reflect"
)

func TestBaseDescriptorSetLocatorID(t *testing.T) {
	base := &BaseDescriptor{}
	
	base.SetLocatorID(10)
	
	lid := base.GetLocatorID()
	
	if lid != 10 {
		t.Errorf("SetLocatorID did not work in BaseDescriptor")
	}
}

func TestBaseDescriptorSetServiceID(t *testing.T) {
	base := &BaseDescriptor{}
	
	base.SetServiceID(10)
	
	lid := base.GetServiceID()
	
	if lid != 10 {
		t.Errorf("SetServiceID did not work in BaseDescriptor")
	}
}

func TestBaseDescriptorSetMetadata(t *testing.T) {
	lMetadata := make(map[string][]string)
	
	fooSlice := []string { "a", "b", "c" }
	barSlice := []string { "x", "y", "z" }
	
	lMetadata["foo"] = fooSlice
	lMetadata["bar"] = barSlice
	
	base := &BaseDescriptor{}
	
	base.SetMetadata(lMetadata)
	
	metadata := base.GetMetadata()
	
	fooResult := metadata["foo"]
	
	if compareSlices(fooSlice, fooResult) == false {
		t.Errorf("fooSlice and fooResult were not the same %v/%v", fooSlice, fooResult)
	}
	
	barResult := metadata["bar"]
	
	if compareSlices(barSlice, barResult) == false {
		t.Errorf("barSlice and barResult were not the same %v/%v", barSlice, barResult)
	}
	
	var fooSliceCopy = make([]string, len(fooSlice))
	copy(fooSliceCopy, fooSlice)
	
	fooSlice = append(fooSlice, "d")
	
	if compareSlices(fooSlice, fooResult) == true {
		t.Errorf("Should now be different, confirming internal copy %v/%v", fooSlice, fooResult)
	}
	
	if compareSlices(fooSliceCopy, fooResult) == false {
		t.Errorf("Confirming that the internal metadata did not change %v/%v", fooSliceCopy, fooResult)
	}
}

func TestBaseDescriptorSetVisibility(t *testing.T) {
	base := &BaseDescriptor{}
	
	base.SetVisibility(api.LOCAL)
	
	lvis := base.GetVisibility()
	
	if lvis != api.LOCAL {
		t.Errorf("SetServiceID did not work in BaseDescriptor")
	}
}

func TestBaseDescriptorSetQualifiers(t *testing.T) {
	base := &BaseDescriptor{}
	
	inputQualifiers := []string { "Red", "Secure" }
	
	base.SetQualifiers(inputQualifiers)
	
	outputQualifiers := base.GetQualifiers()
	
	if compareSlices(inputQualifiers, outputQualifiers) == false {
		t.Errorf("Input and output Qualifiers didn't match")
	}
	
	// Make sure modifying output slice doesn't do anything to internals
	outputQualifiers = append(outputQualifiers, "Plastic")
	outputQualifiers2 := base.GetQualifiers()
	
	if compareSlices(outputQualifiers, outputQualifiers2) == true {
		t.Errorf("Should not have been able to modify internal qualifiers")
	}
}

func TestBaseDescriptorSetName(t *testing.T) {
	base := &BaseDescriptor{}
	
	bob := "Bob"
	
	base.SetName(bob)
	
	lname := base.GetName()
	
	if strings.Compare(bob, lname) != 0 {
		t.Errorf("SetName did not work in BaseDescriptor")
	}
}

func TestBaseDescriptorSetScope(t *testing.T) {
	base := &BaseDescriptor{}
	
	one := "Singleton"
	
	base.SetScope(one)
	
	lscope := base.GetScope()
	
	if strings.Compare(one, lscope) != 0 {
		t.Errorf("SetName did not work in BaseDescriptor")
	}
}

type iFace1 interface {}
type iFace2 interface {}

func TestBaseDescriptorSetContracts(t *testing.T) {
	base := &BaseDescriptor{}
	
	var t1,t2 reflect.Type
	
	t1 = reflect.TypeOf(new(iFace1))
	t2 = reflect.TypeOf(new(iFace2))
	
	contracts := []reflect.Type { t1, t2 }
	
	base.SetAdvertisedInterfaces(contracts)
	
	rContracts := base.GetAdvertisedInterfaces()
	
	if len(contracts) != len(rContracts) {
		t.Errorf("Returned contracts have different lenghts %d/%d", len(contracts), len(rContracts))
	}
	
	for index := range contracts {
		expected := contracts[index]
		got := rContracts[index]
		
		if expected != got {
			t.Errorf("Failure at index %d, contracts not the same %v/%v", index, expected, got)
		}
	}
	
}

func compareSlices(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	for c := range a {
		d := a[c]
		e := b[c]
		
		if strings.Compare(d, e) != 0 {
			return false
		}
		
	}
	
	return true
}
