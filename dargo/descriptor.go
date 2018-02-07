package dargo

import "reflect"

type Descriptor interface {
    create(myType reflect.Type) interface{}
}
