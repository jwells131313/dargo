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

	// ImmediateScope scope services are started immediately
	ImmediateScope = "ImmediateScope"

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

	// ValidationServiceName the name an implementation of ValidationService must have
	ValidationServiceName = "ValidationService"

	// ConfigurationListenerName the name an implementation of ConfigurationListener must have
	ConfigurationListenerName = "ConfigurationListener"

	// InjectionResolver the name an an implementation of InjectionResolver must have
	InjectionResolverName = "InjectionResolver"

	// SystemInjectionResolverQualifierName A qualifier that is put on the system injection
	// resolver for the "inject" field annotation
	SystemInjectionResolverQualifierName = "SystemInjectResolverQualifier"

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

	// ServiceCreationFailure is a type of error returned by ErrorInformation.GetType
	ServiceCreationFailure = "SERVICE_CREATION_FAILURE"

	// LookupValidationFailure is a type of error returned by ErrorInformation.GetType
	LookupValidationFailure = "LOOKUP_VALIDATION_FAILURE"

	// BindOperation is the Bind operation passed in the ValidationInformation
	BindOperation = "BIND"

	// UnbindOperation is the Unbind operation passed in the ValidationInformation
	UnbindOperation = "UNBIND"

	// LookupOperation is the Lookup operation passed in the ValidationInformation
	LookupOperation = "LOOKUP"
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
