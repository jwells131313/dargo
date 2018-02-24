package api
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
	"reflect"
)

// Values for the visibility field of the Descriptor
const (
	// Indicates that this is normal descriptor, visibile to children
	NORMAL = iota
	
	// Indicates taht this is a local descriptor, only visible to its own locator
	LOCAL = iota
)

// Descriptor description of a dargo service description
type Descriptor interface {
	// GetCreateFunction create creates the instance of the type
    GetCreateFunction() func(ServiceLocator) (interface{}, error)
    
    // GetDestroyFunction destroys this service
    GetDestroyFunction() func(ServiceLocator, interface{}) error
    
    // GetAdvertisedInterfaces Returns all interfaces advertised by this service
    GetAdvertisedInterfaces() []reflect.Type
    
    // GetScope Returns the scope of this service
    GetScope() string
    
    // GetName Returns the name of this service (or nil)
    GetName() string
    
    // GetQualifiers Returns the qualifiers of this service
    GetQualifiers() []string
    
    // GetVisibility One of NORMAL or LOCAL
    GetVisibility() int
    
    // GetMetadata returns the metadata for this service
    GetMetadata() map[string][]string
    
    // GetRank Returns the rank of this descriptor
    GetRank() int32
    
    // SetRank Sets the rank of this service
    SetRank(rank int32)
    
    // GetServiceID The serviceid, or -1 if this does not have a serviceid
    GetServiceID() int64
    
    // GetLocatorID The locator id for this service, or -1 if there is not associated locator id
    GetLocatorID() int64
}
