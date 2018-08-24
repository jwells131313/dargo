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
	"sort"
	"sync"
)

const (
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
)

// ServiceLocator The main registry for dargo.  Use it to get context sensitive lookups
// for application services
type ServiceLocator interface {
	// GetService gets the service that is correct for the current context with the given key.
	// It returns the best implementation of the interface
	GetService(toMe ServiceKey) (interface{}, error)

	// GetDService gets the service with the given name in the default namespace and
	// with the provided qualifiers
	// It returns the best implementation of the interface
	GetDService(name string, qualifiers ...string) (interface{}, error)

	// GetAllServices returns all the services matching the service key
	GetAllServices(toMe ServiceKey) ([]interface{}, error)

	// GetService gets the service that is correct for the current context with the given
	// descriptor and any other error if there was an error creating the interface
	GetServiceFromDescriptor(desc Descriptor) (interface{}, error)

	// CreateServiceFromDescriptor creates the service without using contextual data,
	// simply properly invoking the creation method
	CreateServiceFromDescriptor(desc Descriptor) (interface{}, error)

	// GetDescriptors Returns all descriptors that return true when passed through the input function
	// will not return nil, but may return an empty list
	GetDescriptors(Filter) ([]Descriptor, error)

	// GetBestDescriptor returns the best descriptor found returning true through the input function
	// The best descriptor is the one with the highest rank, or if rank is equal the one with the
	// lowest serviceId or if the serviceId are the same the one with the highest locatorId
	GetBestDescriptor(Filter) (Descriptor, error)

	// GetName gets the name of this ServiceLocator
	GetName() string

	// GetID Gets the id of this ServiceLocator
	GetID() int64

	// Will shut down all services associated with this ServiceLocator
	Shutdown()
}

type serviceLocatorData struct {
	lock sync.Mutex
	name string
	ID   int64

	allDescriptors []Descriptor

	nextServiceID int64

	perLookupContext ContextualScope
	singletonContext ContextualScope

	generation uint64
}

// NewServiceLocator this will find or create a service locator with the given name, and
// return errors based on the value of qos
func NewServiceLocator(name string, qos int) (ServiceLocator, error) {
	locatorsLock.Lock()
	defer locatorsLock.Unlock()

	err := checkNameCharacters(name)
	if err != nil {
		return nil, err
	}

	if qos != FailIfPresent && qos != FailIfNotPresent && qos != ReturnExistingOrCreateNew {
		return nil, fmt.Errorf("Unkonwn quality of service %d", qos)
	}

	retVal, found := locators[name]
	if found {
		if qos == FailIfPresent {
			return nil, fmt.Errorf("Quality of service is FailIfPresent and there is a locator with name %s", name)
		}

		return retVal, nil
	}

	// Not found
	if qos == FailIfNotPresent {
		return nil, fmt.Errorf("Quality of service is FailIfNotPresent and there is no locator named %s", name)
	}

	ID := currentID
	currentID = currentID + 1

	retVal = &serviceLocatorData{
		name:             name,
		ID:               ID,
		allDescriptors:   make([]Descriptor, 0),
		perLookupContext: newPerLookupContext(),
	}

	retVal.singletonContext, err = newSingletonScope(retVal)
	if err != nil {
		return nil, err
	}

	serviceLocatorDescriptor := NewConstantDescriptor(SSK(ServiceLocatorName), retVal)
	serviceLocatorSystemDescriptor, err := NewDescriptor(serviceLocatorDescriptor, 0, ID)
	if err != nil {
		return nil, err
	}

	dcs := newDynamicConfigurationService(retVal)
	dynamicConfigurationDescriptor := NewConstantDescriptor(SSK(DynamicConfigurationServiceName), dcs)
	dcsSystemDescriptor, err := NewDescriptor(dynamicConfigurationDescriptor, 1, ID)
	if err != nil {
		return nil, err
	}

	retVal.allDescriptors = append(retVal.allDescriptors, serviceLocatorSystemDescriptor)
	retVal.allDescriptors = append(retVal.allDescriptors, dcsSystemDescriptor)

	retVal.nextServiceID = 2

	locators[name] = retVal

	return retVal, nil
}

func (locator *serviceLocatorData) GetService(toMe ServiceKey) (interface{}, error) {
	f := NewServiceKeyFilter(toMe)

	desc, err := locator.GetBestDescriptor(f)
	if err != nil {
		return nil, err
	}

	if desc == nil {
		return nil, fmt.Errorf("Service %v not found", toMe)
	}

	return locator.createService(desc)
}

func (locator *serviceLocatorData) GetDService(name string, qualifiers ...string) (interface{}, error) {
	return locator.GetService(DSK(name, qualifiers...))
}

func (locator *serviceLocatorData) GetAllServices(toMe ServiceKey) ([]interface{}, error) {
	f := NewServiceKeyFilter(toMe)

	descs, err := locator.GetDescriptors(f)
	if err != nil {
		return nil, err
	}

	retVal := make([]interface{}, 0)
	for _, desc := range descs {
		us, err := locator.createService(desc)
		if err != nil {
			return retVal, err
		}

		retVal = append(retVal, us)
	}

	return retVal, nil
}

func (locator *serviceLocatorData) GetServiceFromDescriptor(desc Descriptor) (interface{}, error) {
	return locator.createService(desc)
}

func (locator *serviceLocatorData) GetDescriptors(filter Filter) ([]Descriptor, error) {
	all, err := locator.internalGetDescriptors(filter)
	if err != nil {
		return nil, err
	}

	return all, nil
}

func (locator *serviceLocatorData) GetBestDescriptor(filter Filter) (Descriptor, error) {
	all, err := locator.internalGetDescriptors(filter)
	if err != nil {
		return nil, err
	}

	if len(all) == 0 {
		return nil, nil
	}

	return all[0], nil
}

func (locator *serviceLocatorData) GetName() string {
	return locator.name
}

func (locator *serviceLocatorData) GetID() int64 {
	return locator.ID
}

func (locator *serviceLocatorData) Shutdown() {
	panic("implement me")
}

func (locator *serviceLocatorData) createService(desc Descriptor) (interface{}, error) {
	scope := desc.GetScope()

	var cs ContextualScope
	if scope == PerLookup {
		cs = locator.perLookupContext
	} else if scope == Singleton {
		cs = locator.singletonContext
	} else {
		service, err := locator.GetService(CSK(scope))
		if err != nil {
			return nil, err
		}

		if service == nil {
			return nil, fmt.Errorf("could not find a scope named %s", scope)
		}

		cs = service.(ContextualScope)
	}

	if cs == nil {
		return nil, fmt.Errorf("Could not find scope named %s in catchall", scope)
	}

	userService, err := cs.FindOrCreate(locator, desc)
	if err != nil {
		return nil, err
	}

	return userService, nil
}

// TODO: This will one day need to, you know, honor the rank and maybe keep caches
func (locator *serviceLocatorData) internalGetDescriptors(filter Filter) ([]Descriptor, error) {
	locator.lock.Lock()
	defer locator.lock.Unlock()

	retVal := make([]Descriptor, 0)
	for _, desc := range locator.allDescriptors {
		if filter.Filter(desc) {
			retVal = append(retVal, desc)
		}
	}

	sort.Slice(retVal, func(i, j int) bool {
		if retVal[i].GetRank() > retVal[j].GetRank() {
			return true
		} else if retVal[i].GetRank() < retVal[j].GetRank() {
			return false
		}

		if retVal[i].GetLocatorID() > retVal[j].GetLocatorID() {
			return true
		} else if retVal[i].GetLocatorID() < retVal[j].GetLocatorID() {
			return false
		}

		if retVal[i].GetServiceID() < retVal[j].GetLocatorID() {
			return true
		}

		return false
	})

	return retVal, nil
}

func (locator *serviceLocatorData) getGeneration() uint64 {
	locator.lock.Lock()
	defer locator.lock.Unlock()

	return locator.generation
}

func (locator *serviceLocatorData) getNextServiceID() int64 {
	locator.lock.Lock()
	defer locator.lock.Unlock()

	retVal := locator.nextServiceID

	locator.nextServiceID = locator.nextServiceID + 1

	return retVal
}

func (locator *serviceLocatorData) update(newDescs []Descriptor, removers []Filter, originalGeneration uint64) error {
	locator.lock.Lock()
	defer locator.lock.Unlock()

	if originalGeneration != locator.generation {
		return fmt.Errorf("Their was an update to the ServiceLocator after this DynamicConfiguration was created")
	}

	newAllDescs := make([]Descriptor, 0)

	removedDescriptors := make([]Descriptor, 0)
	for _, myDesc := range locator.allDescriptors {
		removeMe := false

		for _, removeFilter := range removers {
			removeMe = removeMe || removeFilter.Filter(myDesc)
		}

		if !removeMe {
			newAllDescs = append(newAllDescs, myDesc)
		} else {
			removedDescriptors = append(removedDescriptors, myDesc)
		}
	}

	// TODO: Here the validation service would verify these descriptors were legal removals

	for _, newDesc := range newDescs {
		// TODO: Here the validation service would check to see if the new descriptor could be added

		newAllDescs = append(newAllDescs, newDesc)
	}

	// Seems like at this point we've done all the checking and we can do the actual swap
	locator.allDescriptors = newAllDescs

	return nil
}

func (locator *serviceLocatorData) CreateServiceFromDescriptor(desc Descriptor) (interface{}, error) {
	cf := desc.GetCreateFunction()
	serviceKey, err := newServiceKeyFromDescriptor(desc)
	if err != nil {
		return nil, err
	}

	return cf(locator, serviceKey)
}
