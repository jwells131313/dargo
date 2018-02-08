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

import "reflect"

// Values for the visibility field of the Descriptor
const (
	// Indicates that this is normal descriptor, visibile to children
	NORMAL = iota
	
	// Indicates taht this is a local descriptor, only visible to its own locator
	LOCAL = iota
)

// Descriptor description of a dargo service description
type Descriptor interface {
	// create creates the instance of the type
    create(myType reflect.Type) (interface{}, error)
    
    // Returns all interfaces advertised by this service
    getAdvertisedInterfaces() ([]interface{}, error)
    
    // Returns the scope of this service
    getScope() string
    
    // Returns the name of this service (or nil)
    getName() string
    
    // Returns the qualifiers of this service
    getQualifiers() []string
    
    // One of NORMAL or LOCAL
    getVisibility() int
    
    // returns the metadata for this service
    getMetadata() map[string][]string
    
    // Returns the rank of this descriptor
    getRank() int32
    
    // Sets the rank of this service
    setRank(int32)
    
    // The serviceid, or -1 if this does not have a serviceid
    getServiceId() int64
    
    // The locator id for this service, or -1 if there is not associated locator id
    getLocatorId() int64
}
