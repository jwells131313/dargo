/*
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS HEADER.
 *
 * Copyright (c) 2018 Oracle and/or its affiliates. All rights reserved.
 *
 * The contents of this file are subject to the terms of either the GNU
 * General Public License Version 2 only ("GPL") or the Common Development
 * and Distribution License("CDDL") (collectively, the "License").  You
 * may not use this file except in compliance with the License.  You can
 * obtain a copy of the License at
 * https://glassfish.dev.java.net/public/CDDL+GPL_1_1.html
 * or packager/legal/LICENSE.txt.  See the License for the specific
 * language governing permissions and limitations under the License.
 *
 * When distributing the software, include this License Header Notice in each
 * file and include the License file at packager/legal/LICENSE.txt.
 *
 * GPL Classpath Exception:
 * Oracle designates this particular file as subject to the "Classpath"
 * exception as provided by Oracle in the GPL Version 2 section of the License
 * file that accompanied this code.
 *
 * Modifications:
 * If applicable, add the following below the License Header, with the fields
 * enclosed by brackets [] replaced by your own identifying information:
 * "Portions Copyright [year] [name of copyright owner]"
 *
 * Contributor(s):
 * If you wish your version of this file to be governed by only the CDDL or
 * only the GPL Version 2, indicate your decision by adding "[Contributor]
 * elects to include this software in this distribution under the [CDDL or GPL
 * Version 2] license."  If you don't indicate a single choice of license, a
 * recipient has the option to distribute your version of this file under
 * either the CDDL, the GPL Version 2 or to extend the choice of license to
 * its licensees as provided above.  However, if you add GPL Version 2 code
 * and therefore, elected the GPL Version 2 license, then the option applies
 * only if the new code is made subject to such option by the copyright
 * holder.
 */

package ioc

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type stack interface {
	Push(interface{}) error
	Pop() (interface{}, bool)
	Peek() (interface{}, bool)
}

// IsServiceNotFound returns true if the given error is due to a service
// not being found in the locator.  If the incoming error is a MultiError
// this will return true if any of the contained errors is ServiceNotFoundError
func IsServiceNotFound(e error) bool {
	if e == nil {
		return false
	}

	_, ok := e.(ServiceNotFoundInfo)
	if ok {
		return true
	}

	multi, ok := e.(MultiError)
	if ok {
		for _, e := range multi.GetErrors() {
			_, ok = e.(ServiceNotFoundInfo)
			if ok {
				return true
			}
		}
	}

	return false
}

type stackData struct {
	lock  sync.Mutex
	stack []interface{}
}

func newStack() stack {
	return &stackData{
		stack: make([]interface{}, 0),
	}
}

func (stack *stackData) Push(in interface{}) error {
	stack.lock.Lock()
	defer stack.lock.Unlock()

	if in == nil {
		return fmt.Errorf("nil cannot be pushed")
	}

	stack.stack = append(stack.stack, in)

	return nil
}

func (stack *stackData) Pop() (interface{}, bool) {
	stack.lock.Lock()
	defer stack.lock.Unlock()

	sLen := len(stack.stack)
	if sLen <= 0 {
		return nil, false
	}

	retVal := stack.stack[sLen-1]
	stack.stack = stack.stack[:sLen-1]

	return retVal, true
}

func (stack *stackData) Peek() (interface{}, bool) {
	stack.lock.Lock()
	defer stack.lock.Unlock()

	sLen := len(stack.stack)
	if sLen <= 0 {
		return nil, false
	}

	return stack.stack[sLen-1], true
}

func createAndInject(locator *serviceLocatorData, desc Descriptor, dity reflect.Type, preCreated *reflect.Value) (interface{}, error) {
	numFields := dity.NumField()

	dependencies := make([]*indexAndValueOfDependency, 0)
	depErrors := NewMultiError()

	for lcv := 0; lcv < numFields; lcv++ {
		fieldVal := dity.Field(lcv)

		injectee := newInjectee(desc, dity, fieldVal)

		for _, resolver := range locator.injectionResolvers {
			dependencyAsValue, gotValue, err := resolver.Resolve(locator, injectee)
			if err != nil {
				depErrors.AddError(err)
			} else if gotValue {
				dependencies = append(dependencies, &indexAndValueOfDependency{
					index: lcv,
					value: dependencyAsValue,
				})

				break
			}
		}
	}

	if depErrors.HasError() {
		depErrors.AddError(fmt.Errorf("an error occurred while getting the dependencies of %v", desc))

		var replyError error
		if preCreated == nil {
			locator.runErrorHandlers(ServiceCreationFailure, desc, dity, nil, depErrors)

			replyError = &hasRunHandlers{
				hasRunHandlers:  true,
				underlyingError: depErrors,
			}
		} else {
			replyError = depErrors
		}

		return nil, replyError
	}

	var retVal *reflect.Value
	if preCreated != nil {
		retVal = preCreated
	} else {
		newVal := reflect.New(dity)
		retVal = &newVal
	}

	indirect := reflect.Indirect(*retVal)

	for _, iav := range dependencies {
		index := iav.index
		value := iav.value

		fieldValue := indirect.Field(index)
		errRet := &errorReturn{}

		findAssignableSubStruct(indirect, fieldValue, value)
		safeSet(fieldValue, value, errRet)
		if errRet.err != nil {
			depErrors.AddError(errRet.err)
		}
	}

	if depErrors.HasError() {
		depErrors.AddError(fmt.Errorf("an error occurred while injecting the dependencies of %v", desc))

		var replyError error
		if preCreated == nil {
			locator.runErrorHandlers(ServiceCreationFailure, desc, dity, nil, depErrors)

			replyError = &hasRunHandlers{
				hasRunHandlers:  true,
				underlyingError: depErrors,
			}
		} else {
			replyError = depErrors
		}

		return nil, replyError
	}

	iFace := retVal.Interface()

	initializer, ok := iFace.(DargoInitializer)
	if preCreated == nil && ok {
		errRet := &errorReturn{}
		safeDargoInitialize(initializer, desc, errRet)
		err := errRet.err

		if err != nil {
			_, isMulti := err.(MultiError)
			if !isMulti {
				err = NewMultiError(err)
			}

			locator.runErrorHandlers(ServiceCreationFailure, desc, dity, nil, err)

			replyError := &hasRunHandlers{
				hasRunHandlers:  true,
				underlyingError: err.(MultiError),
			}

			return nil, replyError
		}
	}

	return iFace, nil
}

func findAssignableSubStruct(indirect, toField reflect.Value, fromFieldRef *reflect.Value) {
	fromField := *fromFieldRef
	if fromField.Kind() == reflect.Pointer {
		fromField = reflect.Indirect(fromField)
	}
	fromKind := fromField.Kind()
	fromType := fromField.Type()

	if toField.Kind() == reflect.Pointer {
		toField.Set(reflect.New(toField.Type().Elem()))
		toField = toField.Elem()
	}
	// toKind := toField.Kind()
	toType := toField.Type()

	// Used for debugging
	// fmt.Printf("FROM %s %s\n", fromKind, fromType.Name())
	// fmt.Printf("TO   %s %s\n", toKind, toType.Name())

	if fromKind == reflect.Struct {
		fields := reflect.VisibleFields(fromType)
		for _, field := range fields {
			isAnonymous := field.Anonymous
			isAssignable := field.Type.AssignableTo(toType)
			if isAnonymous && isAssignable {
				val := indirect.FieldByIndex(field.Index)
				*fromFieldRef = val
				return
			}
		}
	}
}

func safeValidate(validator Validator, info ValidationInformation, ret *errorReturn) {
	defer func() {
		if r := recover(); r != nil {
			ret.err = fmt.Errorf("%v", r)
		}
	}()

	ret.err = validator.Validate(info)
}

func safeGetFilter(validationService ValidationService, ret *errorReturn) Filter {
	defer func() {
		if r := recover(); r != nil {
			ret.err = fmt.Errorf("%v", r)
		}
	}()

	return validationService.GetFilter()
}

func safeGetValidator(validationService ValidationService, ret *errorReturn) Validator {
	defer func() {
		if r := recover(); r != nil {
			ret.err = fmt.Errorf("%v", r)
		}
	}()

	return validationService.GetValidator()
}

func safeCreatorFunctions(cf func(ServiceLocator, Descriptor) (interface{}, error),
	l ServiceLocator, d Descriptor, ret *errorReturn) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			ret.err = fmt.Errorf("%v", r)
		}
	}()

	return cf(l, d)
}

// Pesky users can panic, lets not allow that
func safeCallUserErrorService(errorService ErrorService, ei ErrorInformation) error {
	defer func() {
		if r := recover(); r != nil {
			// Ignore me
		}
	}()

	return errorService.OnFailure(ei)
}

func safeConfigurationChanged(configurationListener ConfigurationListener) {
	defer func() {
		if r := recover(); r != nil {
			// Ignore me
		}
	}()

	configurationListener.ConfigurationChanged()
}

func isErrorService(desc Descriptor) bool {
	if UserServicesNamespace == desc.GetNamespace() &&
		ErrorServiceName == desc.GetName() {
		return true
	}

	return false
}

func isValidationService(desc Descriptor) bool {
	if UserServicesNamespace == desc.GetNamespace() &&
		ValidationServiceName == desc.GetName() {
		return true
	}

	return false
}

func isConfigurationListener(desc Descriptor) bool {
	if UserServicesNamespace == desc.GetNamespace() &&
		ConfigurationListenerName == desc.GetName() {
		return true
	}

	return false
}

func isInjectionResolver(desc Descriptor) bool {
	if UserServicesNamespace == desc.GetNamespace() &&
		InjectionResolverName == desc.GetName() {
		return true
	}

	return false
}

func descriptorToIDString(desc Descriptor) string {
	return fmt.Sprintf("%d.%d", desc.GetLocatorID(), desc.GetServiceID())
}

type parseData struct {
	serviceKey ServiceKey
	isOptional bool
}

func parseInjectString(parseMe string) (*parseData, error) {
	if parseMe == "" {
		return nil, fmt.Errorf("no injection string to parse")
	}

	isOptional := false

	namespaceAndName := []string{}

	namespaceNameQualifiersOptions := strings.Split(parseMe, ",")
	for index, value := range namespaceNameQualifiersOptions {
		if index == 0 {
			namespaceAndName = strings.SplitN(value, "#", 2)
		} else {
			if value != "optional" {
				return nil, fmt.Errorf("unknown option %s", value)
			}

			isOptional = true
		}
	}

	var namespace, name string
	if len(namespaceAndName) == 2 {
		namespace = namespaceAndName[0]
		name = namespaceAndName[1]
	} else {
		namespace = DefaultNamespace
		name = namespaceAndName[0]
	}

	qualifiers := make([]string, 0)
	nameAndQualifiers := strings.Split(name, "@")

	for index, val := range nameAndQualifiers {
		if index == 0 {
			name = val
		} else {
			qualifiers = append(qualifiers, val)
		}
	}

	sk, err := NewServiceKey(namespace, name, qualifiers...)
	if err != nil {
		return nil, err
	}

	return &parseData{
		serviceKey: sk,
		isOptional: isOptional,
	}, nil
}
