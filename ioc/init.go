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
	"github.com/jwells131313/goethe"
	"sync"
)

var (
	locatorsLock sync.Mutex
	locators           = make(map[string]*serviceLocatorData)
	currentID    int64 = 1
)

// constant values for the ioc package
const (
	// Indicates that this is normal descriptor, visible to children
	NormalVisibility = iota

	// Indicates taht this is a local descriptor, only visible to its own locator
	LocalVisibility = iota

	// PerLookup Every new lookup is a new service
	PerLookup = "PerLookup"

	// Singleton Is created one time only
	Singleton = "Singleton"

	// ContextScope scope is per context
	ContextScope = "ContextScope"

	// SystemNamespace The namespace for system services
	SystemNamespace = "system"

	// DefaultNamespace A default namespace
	DefaultNamespace = "default"

	// ContextualScopeNamespace A namespace specifically for ContextualScope services
	ContextualScopeNamespace = "sys/scope"

	// UserServicesNamespace A namespace for user-supplied implementations of
	// special services, such as the ValidationService or ErrorService
	UserServicesNamespace = "user/services"

	// ErrorServiceName the name implementations of ErrorService must have
	ErrorServiceName = "ErrorService"

	// FailIfPresent Return an error if a service locator with that name exists
	FailIfPresent = 0

	// ReturnExistingOrCreateNew Return the existing service locator if found, otherwise create it
	ReturnExistingOrCreateNew = 1

	// FailIfNotPresent Return an existing locator with the given name, and fail if it does not already exist
	FailIfNotPresent = 2

	// ServiceLocatorName The name of the ServiceLocator service (in the system namespace)
	ServiceLocatorName = "ServiceLocator"

	// DynamicConfigurationServiceName The name of the DynamicConfigurationService (in the system namespace)
	DynamicConfigurationServiceName = "DynamicConfigurationService"

	// DargoContextCreationServiceName The name of the DargoCreationContextService
	DargoContextCreationServiceName = "DargoContextCreationService"

	// ServiceWithNameNotFoundExceptionString is the string used in the error when a service descriptor is not found
	ServiceWithNameNotFoundExceptionString = "service was not found: %s"

	// LocatorStateRunning This is the state when a locator is currently open and running
	LocatorStateRunning = "Running"

	// LocatorStateShutdown This is the state when a locator has been shut down
	LocatorStateShutdown = "Shutdown"

	// DynamicConfigurationFailure is a type of error returned by ErrorInformation.GetType
	DynamicConfigurationFailure = "DYNAMIC_CONFIGURATION_FAILURE"

	// ServiceCreationFailure is a type of error returned by ErrorInformation.GetTypeß
	ServiceCreationFailure = "SERVICE_CREATION_FAILURE"
)

var (
	// AllFilter is a filter that returns true for every Descriptor
	AllFilter Filter = &allFilterData{}

	// ErrLocatorIsShutdown is returned if the locator you are using has been shut down
	ErrLocatorIsShutdown = fmt.Errorf("locator has been shut down")

	threadManager = goethe.GG()
)

const (
	dargoContextThreadLocal = "DargoContextThreadLocal"
)

func init() {
	threadManager.EstablishThreadLocal(dargoContextThreadLocal, func(tl goethe.ThreadLocal) error {
		tl.Set(newStack())

		return nil

	}, nil)
}

// MultiError contains any number of other errors, useful for cases when
// multiple operations might happen and the user would like to know the
// details of all of them
type MultiError interface {
	// The Error method of this should return a description of all errors
	error
	// AddError adds an error to this MultiError
	AddError(error)
	// GetErrors returns a copy of the errors in this error
	GetErrors() []error
	// GetFinalError returns this MultiError itself if there is at
	// least one embedded error, or nil if this MultiError has no
	// associated errors
	GetFinalError() error
	// HasError returns true if this error has at least one error
	HasError() bool
}

type multiErrorData struct {
	lock   sync.Mutex
	errors []error
}

// NewMultiError creates a multi-error error with the
// provided errors
func NewMultiError(errors ...error) MultiError {
	errorsArray := make([]error, 0)
	for _, err := range errors {
		errorsArray = append(errorsArray, err)
	}

	return &multiErrorData{
		errors: errorsArray,
	}
}

func (med *multiErrorData) Error() string {
	med.lock.Lock()
	defer med.lock.Unlock()

	numErrors := len(med.errors)
	if numErrors == 0 {
		return "there are no errors"
	}
	if numErrors == 1 {
		return med.errors[0].Error()
	}

	first := true
	retVal := ""
	for index, err := range med.errors {
		count := index + 1

		if first {
			first = false

			retVal = fmt.Sprintf("%d. %s", count, err.Error())
		} else {
			retVal = fmt.Sprintf("%s\n%d. %s", retVal, count, err.Error())
		}
	}

	return retVal
}

func (med *multiErrorData) AddError(err error) {
	med.lock.Lock()
	defer med.lock.Unlock()

	multi, ok := err.(MultiError)
	if !ok {
		med.errors = append(med.errors, err)
	} else {
		// Get all the internal errors and add them to this error
		for _, internalErr := range multi.GetErrors() {
			med.errors = append(med.errors, internalErr)
		}
	}
}

func (med *multiErrorData) GetErrors() []error {
	med.lock.Lock()
	defer med.lock.Unlock()

	retVal := make([]error, len(med.errors))
	copy(retVal, med.errors)

	return retVal
}

func (med *multiErrorData) GetFinalError() error {
	med.lock.Lock()
	defer med.lock.Unlock()

	if len(med.errors) == 0 {
		return nil
	}

	return med
}

func (med *multiErrorData) HasError() bool {
	med.lock.Lock()
	defer med.lock.Unlock()

	return len(med.errors) > 0
}

func (med *multiErrorData) String() string {
	return fmt.Sprintf("MultiError(%s)", med.Error())
}
