package dargo

import "reflect"
import "errors"
import "fmt"

/**
 * Binds the descriptor to the string I guess
 */
func Bind(desc *Descriptor, toMe reflect.Type, name string, metadata map[string]string) error {
	kind := toMe.Kind();
	
	fmt.Println("kind=", kind.String())
	
	if (kind != reflect.Interface) {
		return errors.New("toMe must be an interface")
	}
	
	return nil
}
