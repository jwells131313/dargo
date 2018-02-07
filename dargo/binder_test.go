package dargo

import "testing"
import "reflect"

type shape interface {
	
}

func TestBinder(t *testing.T) {
  desc := new(Descriptor)
  name := "Nick Foles"
  
  Bind(desc, reflect.TypeOf(new(shape)).Elem(), name, nil)
}

func TestPointerBinderFailure(t *testing.T) {
  desc := new(Descriptor)
  name := "Nick Foles"
  
  error := Bind(desc, reflect.TypeOf(new(shape)), name, nil)
  if (error == nil) {
  	t.Errorf("Should have been a failure because reflect.TypeOf(new(shape)) is a pointer")
  }
}
