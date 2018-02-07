package dargo

import "testing"
import "reflect"

func TestBinder(t *testing.T) {
  desc := new(Descriptor)
  name := "Nick Foles"
  
  Bind(desc, reflect.TypeOf(new(Shape)).Elem(), name, nil)
}

func TestPointerBinderFailure(t *testing.T) {
  desc := new(Descriptor)
  name := "Nick Foles"
  
  error := Bind(desc, reflect.TypeOf(new(Shape)), name, nil)
  if (error == nil) {
  	t.Errorf("Should have been a failure because reflect.TypeOf(new(Shape)) is a pointer")
  }
}
