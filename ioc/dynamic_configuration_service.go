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
	"sync"
)

// DynamicConfigurationService This service is bound into every
// service locator and can be used to add application services
// to the context sensitive registry
type DynamicConfigurationService interface {
	CreateDynamicConfiguration() (DynamicConfiguration, error)
}

// DynamicConfiguration use this to add and remove descriptors to
// the service locator
type DynamicConfiguration interface {
	// BindWithCreator adds a descriptor to be bound into the service locator
	// returns the copy of the descriptor that will bound in if
	// commit succeeds
	Bind(desc Descriptor) (Descriptor, error)

	// AddRemoveFilter adds a filter that will be run over all
	// existing descriptors to determing which ones to remove
	// from the locator
	AddRemoveFilter(Filter) error

	Commit() error
}

type dynamicConfigData struct {
	parent *serviceLocatorData
}

func newDynamicConfigurationService(parent *serviceLocatorData) DynamicConfigurationService {
	return &dynamicConfigData{
		parent: parent,
	}
}

func (dcd *dynamicConfigData) CreateDynamicConfiguration() (DynamicConfiguration, error) {
	return newModifier(dcd.parent), nil
}

type dynamicConfigModificationData struct {
	lock               sync.Mutex
	state              int
	parent             *serviceLocatorData
	originalGeneration uint64
	binds              []Descriptor
	removeFilters      []Filter
}

func newModifier(parent *serviceLocatorData) DynamicConfiguration {
	return &dynamicConfigModificationData{
		parent:             parent,
		originalGeneration: parent.getGeneration(),
		binds:              make([]Descriptor, 0),
		removeFilters:      make([]Filter, 0),
	}
}

func (mod *dynamicConfigModificationData) checkState() error {
	if mod.state != 0 {
		return fmt.Errorf("This dynamic configuration has been committed or closed for other reasons")
	}

	return nil
}

func (mod *dynamicConfigModificationData) Bind(desc Descriptor) (Descriptor, error) {
	mod.lock.Lock()
	defer mod.lock.Unlock()

	err := mod.checkState()
	if err != nil {
		return nil, err
	}

	serviceID := mod.parent.getNextServiceID()
	locatorID := mod.parent.GetID()

	retVal, err := NewDescriptor(desc, serviceID, locatorID)
	if err != nil {
		return nil, err
	}

	mod.binds = append(mod.binds, retVal)

	return retVal, nil
}

func (mod *dynamicConfigModificationData) AddRemoveFilter(f Filter) error {
	mod.lock.Lock()
	defer mod.lock.Unlock()

	err := mod.checkState()
	if err != nil {
		return err
	}

	mod.removeFilters = append(mod.removeFilters, f)

	return nil
}

func (mod *dynamicConfigModificationData) Commit() error {
	mod.lock.Lock()
	defer mod.lock.Unlock()

	err := mod.checkState()
	if err != nil {
		return err
	}

	err = mod.parent.update(mod.binds, mod.removeFilters, mod.originalGeneration)
	mod.state = 1
	if err != nil {
		_, ok := err.(MultiError)
		if !ok {
			err = NewMultiError(err)
		}

		mod.parent.runErrorHandlers(DynamicConfigurationFailure, nil, nil, err)

		return err
	}

	return nil
}
