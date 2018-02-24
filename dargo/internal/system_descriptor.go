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
	"reflect"
	"../api"

)

type systemDescriptor struct {
    creator func() (interface{}, error)
    destroyer func(interface{}) error
	myContracts []reflect.Type
	scope string
	name string
	qualifiers []string
	visibility int
	metadata map[string][]string
	rank int32
	serviceid int64
	locatorid int64
}

// CopyDescriptor Makes a full copy of the incoming descriptor and returns
// a strictly read only copy of the descriptor (except where it is normally
// writeable such as Rank)
func CopyDescriptor(desc api.Descriptor) api.Descriptor {
	retVal := &systemDescriptor{}
	
	retVal.locatorid = desc.GetLocatorID()
	retVal.serviceid = desc.GetServiceID()
	retVal.rank = desc.GetRank()
	retVal.metadata = copyMetadata(desc.GetMetadata())
	retVal.visibility = desc.GetVisibility()
	retVal.qualifiers = copyStringArray(desc.GetQualifiers())
	retVal.name = desc.GetName()
	retVal.scope = desc.GetScope()
	retVal.myContracts = copyAdvertised(desc.GetAdvertisedInterfaces())
	retVal.destroyer = desc.GetDestroyFunction()
	retVal.creator = desc.GetCreateFunction()
	
	return retVal
}

// Create create creates the instance of the type
func (base *systemDescriptor) GetCreateFunction() func() (interface{}, error) {
	return base.creator
}

func (base *systemDescriptor) GetDestroyFunction() func(interface{}) error {
	return base.destroyer
}

// GetAdvertisedInterfaces Returns all interfaces advertised by this service
func (base *systemDescriptor) GetAdvertisedInterfaces() []reflect.Type {
	return copyAdvertised(base.myContracts)
}

// GetScope Returns the scope of this service
func (base *systemDescriptor) GetScope() string {
	return base.scope
}

// GetName Returns the name of this service (or nil)
func (base *systemDescriptor) GetName() string {
	return base.name
}

// GetQualifiers Returns the qualifiers of this service
func (base *systemDescriptor) GetQualifiers() []string {
	return copyStringArray(base.qualifiers)
}

// GetVisibility One of NORMAL or LOCAL
func (base *systemDescriptor) GetVisibility() int {
	return base.visibility
}

// GetMetadata returns the metadata for this service
func (base *systemDescriptor) GetMetadata() map[string][]string {
	return copyMetadata(base.metadata)
}

// GetRank Returns the rank of this descriptor
func (base *systemDescriptor) GetRank() int32 {
	return base.rank
}

// SetRank Sets the rank of this service
func (base *systemDescriptor) SetRank(rank int32) {
	base.rank = rank
}

// GetServiceID The serviceid, or -1 if this does not have a serviceid
func (base *systemDescriptor) GetServiceID() int64 {
	return base.serviceid
}

// GetLocatorID The locator id for this service, or -1 if there is not associated locator id
func (base *systemDescriptor) GetLocatorID() int64 {
	return base.locatorid
}

