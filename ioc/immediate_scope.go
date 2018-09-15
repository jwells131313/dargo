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
	"github.com/jwells131313/goethe/cache"
	"sync"
	"time"
)

var immediateFilter = &immediateFilterData{}

// ImmediateScopeData is the implementation of ContextualScope for ImmediateScope
type ImmediateScopeData struct {
	Locator ServiceLocator `inject:"system#ServiceLocator"`
	cache   cache.Cache
}

// GetScope implements the ContextualScope interface
func (isd *ImmediateScopeData) GetScope() string {
	return ImmediateScope
}

// FindOrCreate implements the ContextualScope interface
func (isd *ImmediateScopeData) FindOrCreate(locator ServiceLocator, desc Descriptor) (interface{}, error) {
	return isd.cache.Compute(idKey{desc: desc})
}

// ContainsKey implements the ContextualScope interface
func (isd *ImmediateScopeData) ContainsKey(locator ServiceLocator, desc Descriptor) bool {
	return isd.cache.HasKey(idKey{desc: desc})
}

// DestroyOne implements the ContextualScope interface
func (isd *ImmediateScopeData) DestroyOne(locator ServiceLocator, desc Descriptor) error {
	lookForMe := idKey{desc: desc}

	isd.cache.Remove(func(key interface{}, value interface{}) bool {
		if key == lookForMe {
			isd.actualDestruction(desc, value)

			return true
		}

		return false
	})

	return nil
}

// GetSupportsNilCreation implements the ContextualScope interface
func (isd *ImmediateScopeData) GetSupportsNilCreation(locator ServiceLocator) bool {
	return false
}

// IsActive implements the ContextualScope interface
func (isd *ImmediateScopeData) IsActive(locator ServiceLocator) bool {
	return true
}

// Shutdown implements the ContextualScope interface
func (isd *ImmediateScopeData) Shutdown(locator ServiceLocator) {
	tid := threadManager.GetThreadID()
	if tid < 0 {
		c := make(chan bool)

		threadManager.Go(isd.channelShutdown, c)

		<-c

		return
	}

	isd.internalShutdown()
}

// DargoInitialize initializes the scope
func (isd *ImmediateScopeData) DargoInitialize(desc Descriptor) error {
	c, err := cache.NewCache(isd, func(in interface{}) error {
		return fmt.Errorf("cycle detected in immediate scope involving %v", in)
	})

	if err != nil {
		return err
	}

	isd.cache = c

	return nil
}

// Compute gets the singleton-like service value
func (isd *ImmediateScopeData) Compute(in interface{}) (interface{}, error) {
	key, ok := in.(idKey)
	if !ok {
		return nil, fmt.Errorf("incomding key not the expected type %v", in)
	}

	return isd.Locator.CreateServiceFromDescriptor(key.desc)
}

func (isd *ImmediateScopeData) channelShutdown(replyChan chan bool) {
	isd.internalShutdown()

	replyChan <- true
}

func (isd *ImmediateScopeData) internalShutdown() {
	isd.cache.Remove(func(key interface{}, value interface{}) bool {
		idKey := key.(idKey)

		isd.actualDestruction(idKey.desc, value)

		return true
	})

}

func (isd *ImmediateScopeData) actualDestruction(desc Descriptor, value interface{}) {
	if desc == nil {
		return
	}

	df := desc.GetDestroyFunction()
	if df == nil {
		return
	}

	df(isd.Locator, desc, value)
}

// ImmediateConfigurationListerData structure for the ImmediateService configuration listener service
type ImmediateConfigurationListerData struct {
	lock              sync.Mutex
	Locator           ServiceLocator  `inject:"system#ServiceLocator"`
	Context           ContextualScope `inject:"sys/scope#ImmediateScope"`
	immediateServices map[Descriptor]string
	workQueue         goethe.FunctionQueue
	threadPool        goethe.Pool
}

// DargoInitialize initializes the configuration listener
func (listener *ImmediateConfigurationListerData) DargoInitialize(desc Descriptor) error {
	descriptors, err := listener.Locator.GetDescriptors(immediateFilter)
	if err != nil {
		return err
	}

	listener.workQueue = goethe.NewBoundedFunctionQueue(10000000)

	p, err := threadManager.NewPool(listener.Locator.GetName(), 0, 1,
		5*time.Minute, listener.workQueue, nil)
	if err != nil {
		return err
	}

	listener.threadPool = p

	err = listener.threadPool.Start()
	if err != nil {
		return err
	}

	services := make(map[Descriptor]string)
	for _, desc := range descriptors {
		services[desc] = desc.GetFullName()
	}

	listener.immediateServices = services

	for desc, _ := range services {
		listener.workQueue.Enqueue(func() {
			listener.Locator.GetServiceFromDescriptor(desc)
		})
	}

	return nil
}

// ConfigurationChanged is called whenever there was a change to the set of descriptors
func (listener *ImmediateConfigurationListerData) ConfigurationChanged() {
	listener.lock.Lock()
	defer listener.lock.Unlock()

	descriptors, err := listener.Locator.GetDescriptors(immediateFilter)
	if err != nil {
		return
	}

	removed := make(map[Descriptor]string)
	for key, value := range listener.immediateServices {
		removed[key] = value
	}

	newValue := make(map[Descriptor]string)
	added := make([]Descriptor, 0)
	for _, desc := range descriptors {
		delete(removed, desc)
		_, found := listener.immediateServices[desc]
		if !found {
			added = append(added, desc)
		}

		newValue[desc] = desc.GetFullName()
	}

	listener.immediateServices = newValue

	for desc, _ := range removed {
		listener.workQueue.Enqueue(func() {
			listener.Context.DestroyOne(listener.Locator, desc)
		})
	}

	for _, desc := range added {
		listener.workQueue.Enqueue(func() {
			listener.Locator.GetServiceFromDescriptor(desc)
		})
	}

}

type immediateFilterData struct{}

// Filter gets all the services in the Immediate scope
func (filter *immediateFilterData) Filter(desc Descriptor) bool {
	if desc.GetScope() == ImmediateScope {
		return true
	}

	return false
}
