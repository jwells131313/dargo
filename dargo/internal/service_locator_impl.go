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
	"fmt"
	"reflect"
	"../api"
)

// ServiceLocatorImpl An internal implementation of ServiceLocator
type ServiceLocatorImpl struct {
	name string
	id int64
	
	allDescriptors []api.Descriptor
}

// NewServiceLocator creates a new ServiceLocator with the given name and ID
func NewServiceLocator(lName string, lID int64) api.ServiceLocator {
	retVal := &ServiceLocatorImpl {
		name: lName,
		id: lID,
		allDescriptors: make([]api.Descriptor, 10),
	}
	
	
	
	return retVal
}

// GetService gets the service associated with the type
func (locator *ServiceLocatorImpl) GetService(toMe reflect.Type) (interface{}, error) {
	fmt.Println("Looking for something of type", toMe)
	return nil, nil
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
