package internal

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

import (
	"github.com/jwells131313/dargo/api"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// ServiceLocatorImpl An internal implementation of ServiceLocator
type ServiceLocatorImpl struct {
	countMux, mux sync.Mutex
	lockCount     uint64

	name string
	id   int64

	allDescriptors []api.Descriptor
	byName         map[string][]api.Descriptor
	byType         map[reflect.Type][]api.Descriptor

	lastTakenServiceID int64
	lastChange         int64

	perLookupContext, singletonContext api.Context
}

// NewServiceLocator creates a new ServiceLocator with the given name and ID
func NewServiceLocator(lName string, lID int64) api.ServiceLocator {
	retVal := &ServiceLocatorImpl{
		name:               lName,
		id:                 lID,
		allDescriptors:     make([]api.Descriptor, 2),
		byName:             make(map[string][]api.Descriptor),
		byType:             make(map[reflect.Type][]api.Descriptor),
		lastTakenServiceID: 2,
		lastChange:         0,
		perLookupContext:   &PerLookupContext{},
	}

	cDesc := NewConstantDescriptor(retVal)
	cDesc.AddAdvertisedInterface(reflect.TypeOf(new(api.ServiceLocator)).Elem())

	sDesc0 := CopyDescriptor(cDesc, lID, 0)
	retVal.addDescriptorToStructures(sDesc0)

	dcs := NewDynamicConfigurationService(retVal)
	dDesc := NewConstantDescriptor(dcs)
	dDesc.AddAdvertisedInterface(reflect.TypeOf(new(api.DynamicConfigurationService)).Elem())

	sDesc1 := CopyDescriptor(dDesc, lID, 1)
	retVal.addDescriptorToStructures(sDesc1)

	return retVal
}

func (locator *ServiceLocatorImpl) addDescriptorToStructures(desc api.Descriptor) {
	locator.allDescriptors = append(locator.allDescriptors, desc)

	name := desc.GetName()
	if name != "" {
		nameSlice, found := locator.byName[name]
		if found {
			locator.byName[name] = append(nameSlice, desc)
		} else {
			locator.byName[name] = []api.Descriptor{desc}
		}
	}

	advertised := desc.GetAdvertisedInterfaces()
	for _, contract := range advertised {
		typeSlice, found := locator.byType[contract]
		if found {
			locator.byType[contract] = append(typeSlice, desc)
		} else {
			locator.byType[contract] = []api.Descriptor{desc}
		}

	}
}

// GetService gets the service associated with the type
func (locator *ServiceLocatorImpl) GetService(toMe reflect.Type) (interface{}, error) {
	return locator.GetServiceWithName(toMe, "")
}

// GetServiceWithName gets the service of the given type with the given name
func (locator *ServiceLocatorImpl) GetServiceWithName(toMe reflect.Type, name string) (interface{}, error) {
	locator.lock()
	defer locator.unlock()

	myDescriptor := locator.getDescriptorOfTypeWithName(toMe, name)

	if myDescriptor == nil {
		return nil, nil
	}

	retVal, err := locator.GetServiceFromDescriptor(myDescriptor)

	return retVal, err
}

// GetServiceFromDescriptor gets the contextual service from the given descriptor
func (locator *ServiceLocatorImpl) GetServiceFromDescriptor(desc api.Descriptor) (interface{}, error) {
	locator.lock()
	defer locator.unlock()

	context, err := locator.getActiveContext(desc)
	if err != nil {
		return nil, err
	}

	if context == nil {
		return nil, fmt.Errorf("No active context was found for %s", desc.GetScope())

	}

	retVal, err := context.FindOrCreate(locator, desc)

	return retVal, err
}

func (locator *ServiceLocatorImpl) getActiveContext(desc api.Descriptor) (api.Context, error) {
	scope := desc.GetScope()

	if api.PerLookup == scope {
		return locator.perLookupContext, nil
	}

	if api.Singleton == scope {
		return locator.singletonContext, nil
	}

	namedDescriptors := locator.getDescriptorsOfTypeWithName(reflect.TypeOf(new(api.Context)).Elem(), scope)
	for _, contextDescriptor := range namedDescriptors {
		rawContext, err := locator.GetServiceFromDescriptor(contextDescriptor)
		if err != nil {
			return nil, err
		}

		context := rawContext.(api.Context)

		if context.IsActive(locator) {
			return context, nil
		}
	}

	return nil, nil
}

// GetDescriptors Returns all descriptors that return true when passed through the input function
// will not return nil, but may return an empty list
func (locator *ServiceLocatorImpl) GetDescriptors(filter func(api.Descriptor) bool) []api.Descriptor {
	return locator.GetDescriptorsWithNameOrType(filter, nil, "")

}

// GetBestDescriptor returns the best descriptor found returning true through the input function
// The best descriptor is the one with the highest rank, or if rank is equal the one with the
// lowest serviceId or if the serviceId are the same the one with the highest locatorId
func (locator *ServiceLocatorImpl) GetBestDescriptor(filter func(api.Descriptor) bool) api.Descriptor {
	return locator.GetBestDescriptorWithNameOrType(filter, nil, "")
}

func (locator *ServiceLocatorImpl) getDescriptorsOfTypeWithName(toMe reflect.Type, name string) []api.Descriptor {
	return locator.GetDescriptorsWithNameOrType(func(api.Descriptor) bool {
		return true
	}, toMe, name)
}

func (locator *ServiceLocatorImpl) getDescriptorOfTypeWithName(toMe reflect.Type, name string) api.Descriptor {
	return locator.GetBestDescriptorWithNameOrType(func(api.Descriptor) bool {
		return true
	}, toMe, name)
}

// GetDescriptorsWithNameOrType Returns all descriptors that return true when passed through the input function
// and which have the given name.  Can drastically reduce the number of descriptors passed to the method
// will not return nil, but may return an empty list
func (locator *ServiceLocatorImpl) GetDescriptorsWithNameOrType(filter func(api.Descriptor) bool, toMe reflect.Type, name string) []api.Descriptor {
	var originalList []api.Descriptor
	var found bool

	if name != "" {
		originalList, found = locator.byName[name]
		if !found {
			// None with given name, can actually just return now
			return []api.Descriptor{}
		}

		if toMe != nil {
			var ofTypeSlice []api.Descriptor
			ofTypeSlice = []api.Descriptor{}

			for _, iDesc := range originalList {
				contracts := iDesc.GetAdvertisedInterfaces()

				for _, contract := range contracts {
					if contract == toMe {
						ofTypeSlice = append(ofTypeSlice, iDesc)
					}
				}

			}

		}
	} else if toMe != nil {
		originalList, found = locator.byType[toMe]

		if !found {
			// None with given type, can actually just return now
			return []api.Descriptor{}
		}
	} else {
		originalList = locator.allDescriptors
	}

	// Run the original list through the filter
	filteredDescriptors := []api.Descriptor{}
	for _, desc := range originalList {
		if filter(desc) {
			filteredDescriptors = append(filteredDescriptors, desc)
		}
	}

	// And now sort the returned filtered descriptors
	sort.Slice(filteredDescriptors, func(i, j int) bool {
		iSid := filteredDescriptors[i].GetServiceID()
		jSid := filteredDescriptors[j].GetServiceID()

		if iSid < jSid {
			return true
		}
		if iSid > jSid {
			return false
		}

		// They are the same, take largest locator ID
		iLid := filteredDescriptors[i].GetLocatorID()
		jLid := filteredDescriptors[j].GetLocatorID()

		if iLid > jLid {
			return true
		}

		return false
	})

	return filteredDescriptors
}

// GetBestDescriptorWithNameOrType returns the best descriptor found returning true through the input function
// and which have the given name
// The best descriptor is the one with the highest rank, or if rank is equal the one with the
// lowest serviceId or if the serviceId are the same the one with the highest locatorId
func (locator *ServiceLocatorImpl) GetBestDescriptorWithNameOrType(filter func(api.Descriptor) bool, toMe reflect.Type, name string) api.Descriptor {
	filteredDescriptors := locator.GetDescriptorsWithNameOrType(filter, toMe, name)
	if filteredDescriptors == nil {
		return nil
	}
	if len(filteredDescriptors) == 0 {
		return nil
	}

	return filteredDescriptors[0]
}

// GetName gets the name associated with this locator
func (locator *ServiceLocatorImpl) GetName() string {
	return locator.name
}

// GetID gets the id associated with this locator
func (locator *ServiceLocatorImpl) GetID() int64 {
	return locator.id
}

// Shutdown shuts down this locator
func (locator *ServiceLocatorImpl) Shutdown() {
	// do nothing
}

func (locator *ServiceLocatorImpl) getLastChange() int64 {
	locator.lock()
	defer locator.unlock()

	return locator.lastChange
}

func (locator *ServiceLocatorImpl) getAndIncrementSID() int64 {
	locator.lock()
	defer locator.unlock()

	retVal := locator.lastTakenServiceID
	locator.lastTakenServiceID = locator.lastTakenServiceID + 1

	return retVal
}

func (locator *ServiceLocatorImpl) updateDescriptors(binds []api.Descriptor, unbinds []func(api.Descriptor) bool) error {
	locator.lock()
	locator.unlock()

	return nil
}

func (locator *ServiceLocatorImpl) lock() {
	locator.countMux.Lock()
	defer locator.countMux.Unlock()

	if locator.lockCount == 0 {
		locator.lockCount = 1
		locator.mux.Lock()
	} else {
		locator.lockCount = locator.lockCount + 1
	}
}

func (locator *ServiceLocatorImpl) unlock() {
	locator.countMux.Lock()
	defer locator.countMux.Unlock()

	if locator.lockCount == 0 {
		return
	}

	locator.lockCount = locator.lockCount - 1
	if locator.lockCount == 0 {
		locator.mux.Unlock()
	}
}
