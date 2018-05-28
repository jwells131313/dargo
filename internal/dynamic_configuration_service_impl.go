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

package internal

import (
	"fmt"
	"github.com/jwells131313/dargo/api"
	"sync"
)

type dynamicConfigurationService struct {
	locator *ServiceLocatorImpl
}

type dynamicConfiguration struct {
	locator *ServiceLocatorImpl

	mux                 sync.Mutex
	locatorChangeNumber int64
	binds               []api.Descriptor
	destroys            []func(api.Descriptor) bool
	closed              bool
}

// NewDynamicConfigurationService Creates a dynamic configuration service with the given locator
func NewDynamicConfigurationService(l *ServiceLocatorImpl) api.DynamicConfigurationService {
	retVal := &dynamicConfigurationService{
		locator: l,
	}

	return retVal
}

// CreateDynamicConfiguration creates a dynamic configuration that can be used to bind into the locator
func (dcs *dynamicConfigurationService) CreateDynamicConfiguration() api.DynamicConfiguration {
	lChange := dcs.locator.getLastChange()

	retVal := &dynamicConfiguration{
		locator:             dcs.locator,
		locatorChangeNumber: lChange,
		binds:               []api.Descriptor{},
		destroys:            []func(api.Descriptor) bool{},
		closed:              false,
	}

	return retVal
}

// Bind adds a descriptor to be bound into the service locator
// returns the copy of the descriptor that will bound in if
// commit succeeds
func (dConfig *dynamicConfiguration) Bind(desc api.Descriptor) api.Descriptor {
	dConfig.mux.Lock()
	defer dConfig.mux.Unlock()

	lID := dConfig.locator.GetID()
	sID := dConfig.locator.getAndIncrementSID()

	copied := CopyDescriptor(desc, lID, sID)
	dConfig.binds = append(dConfig.binds, copied)

	return copied
}

// AddRemoveFilter adds a filter that will be run over all
// existing descriptors to determing which ones to remove
// from the locator
func (dConfig *dynamicConfiguration) AddRemoveFilter(killer func(api.Descriptor) bool) {
	dConfig.mux.Lock()
	defer dConfig.mux.Unlock()

	dConfig.destroys = append(dConfig.destroys, killer)
}

func (dConfig *dynamicConfiguration) Commit() error {
	dConfig.mux.Lock()
	defer dConfig.mux.Unlock()
	if dConfig.closed {
		return fmt.Errorf("This configuration has been closed possibly because its already been committed")
	}

	lChange := dConfig.locator.getLastChange()
	if dConfig.locatorChangeNumber != lChange {
		dConfig.closed = true
		return fmt.Errorf("The locator %s has been changed since this dynamic configuration was created.  Will not commit it",
			dConfig.locator.GetName())
	}

	dConfig.closed = true
	return dConfig.locator.updateDescriptors(dConfig.binds, dConfig.destroys)
}
