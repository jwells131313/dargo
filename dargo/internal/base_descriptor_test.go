package internal

import (
    "testing"
    "strings"
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
