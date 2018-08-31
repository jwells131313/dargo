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

// Descriptor description of a dargo service description
type Descriptor interface {
	// GetCreateFunction create creates the instance of the type
	GetCreateFunction() func(ServiceLocator, Descriptor) (interface{}, error)

	// GetDestroyFunction destroys this service
	GetDestroyFunction() func(ServiceLocator, Descriptor, interface{}) error

	// GetNamespace returns the namespace this service is in (may not be empty string)
	GetNamespace() string

	// GetName Returns the name of this service (may not be empty string)
	GetName() string

	// GetScope Returns the scope of this service
	GetScope() string

	// GetQualifiers Returns the qualifiers of this service
	GetQualifiers() []string

	// GetVisibility One of NORMAL or LOCAL
	GetVisibility() int

	// GetMetadata returns the metadata for this service
	GetMetadata() map[string][]string

	// GetRank Returns the rank of this descriptor
	GetRank() int32

	// SetRank Sets the rank of this service, returns the old rank
	SetRank(rank int32) int32

	// GetServiceID The serviceid, or -1 if this does not have a serviceid
	GetServiceID() int64

	// GetLocatorID The locator id for this service, or -1 if there is not associated locator id
	GetLocatorID() int64
}

// WriteableDescriptor A writeable version of a descriptor
type WriteableDescriptor interface {
	Descriptor

	// SetCreateFunction create creates the instance of the type
	SetCreateFunction(func(ServiceLocator, Descriptor) (interface{}, error)) error

	// SetDestroyFunction destroys this service
	SetDestroyFunction(func(ServiceLocator, Descriptor, interface{}) error) error

	// SetNamespace sets the namespace of this descriptor, may not be empty
	SetNamespace(string) error

	// SetName sets the name of this descriptor, may not be empty
	SetName(string) error

	// SetScope sets the scope of this service
	SetScope(string) error

	// SetQualifiers sets the qualifiers of this service
	SetQualifiers([]string) error

	// SetVisibility setsOne of NORMAL or LOCAL
	SetVisibility(int) error

	// SetMetadata sets the metadata for this service
	SetMetadata(map[string][]string) error
}

type baseDescriptor struct {
	lock                   sync.Mutex
	namespace, name, scope string
	qualifiers             []string
	visibility             int
	metadata               map[string][]string
	rank                   int32
	serviceID, locatorID   int64
}

type descriptorImpl struct {
	baseDescriptor
	creator   func(ServiceLocator, Descriptor) (interface{}, error)
	destroyer func(ServiceLocator, Descriptor, interface{}) error
}

type writeableDescriptorImpl struct {
	baseDescriptor
	creator   func(ServiceLocator, Descriptor) (interface{}, error)
	destroyer func(ServiceLocator, Descriptor, interface{}) error
}

type constantDescriptorImpl struct {
	writeableDescriptorImpl
	cnstnt interface{}
}

// NewDescriptor create a read-only descriptor deep copy of the incoming descriptor
func NewDescriptor(desc Descriptor, serviceID, locatorID int64) (Descriptor, error) {
	creator := desc.GetCreateFunction()
	if creator == nil {
		return nil, fmt.Errorf("descriptor must have create function")
	}

	namespace := desc.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("descriptor must have namespace")
	}

	name := desc.GetName()
	if name == "" {
		return nil, fmt.Errorf("descriptor must have name")
	}

	scope := desc.GetScope()
	if scope == "" {
		return nil, fmt.Errorf("descriptor must have scope")
	}

	qualifiers := desc.GetQualifiers()
	if qualifiers == nil {
		qualifiers = make([]string, 0)
	}

	visibility := desc.GetVisibility()
	if visibility != LocalVisibility && visibility != NormalVisibility {
		return nil, fmt.Errorf("descriptor must have visibility LocalVisibility or Normal, it is %d", visibility)
	}

	retVal := &descriptorImpl{
		creator:   creator,
		destroyer: desc.GetDestroyFunction(),
	}

	retVal.namespace = namespace
	retVal.name = name
	retVal.scope = scope
	retVal.qualifiers = qualifiers
	retVal.visibility = visibility
	retVal.metadata = copyMetadata(desc.GetMetadata())
	retVal.rank = desc.GetRank()
	retVal.serviceID = serviceID
	retVal.locatorID = locatorID

	return retVal, nil
}

// NewWriteableDescriptor creates a writeable descriptor
func NewWriteableDescriptor() WriteableDescriptor {
	retVal := &writeableDescriptorImpl{}

	retVal.namespace = DefaultNamespace
	retVal.qualifiers = make([]string, 0)
	retVal.metadata = make(map[string][]string)
	retVal.visibility = NormalVisibility
	retVal.scope = Singleton

	return retVal
}

// NewConstantDescriptor creates a descriptor that always resolves to exactly the constant
// passed in.  It will by default be put into the PerLookup scope
func NewConstantDescriptor(sk ServiceKey, cnstnt interface{}) WriteableDescriptor {
	retVal := &constantDescriptorImpl{
		cnstnt: cnstnt,
	}

	retVal.namespace = sk.GetNamespace()
	retVal.name = sk.GetName()
	retVal.qualifiers = sk.GetQualifiers()
	retVal.metadata = make(map[string][]string)
	retVal.visibility = NormalVisibility
	retVal.scope = PerLookup

	return retVal

}

func (di *descriptorImpl) GetCreateFunction() func(sl ServiceLocator, desc Descriptor) (interface{}, error) {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.creator
}

func (di *descriptorImpl) GetDestroyFunction() func(ServiceLocator, Descriptor, interface{}) error {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.destroyer
}

func (di *baseDescriptor) GetNamespace() string {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.namespace
}

func (di *baseDescriptor) GetName() string {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.name
}

func (di *baseDescriptor) GetScope() string {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.scope
}

func (di *baseDescriptor) GetQualifiers() []string {
	di.lock.Lock()
	defer di.lock.Unlock()

	retVal := make([]string, len(di.qualifiers))
	copy(retVal, di.qualifiers)
	return retVal
}

func (di *baseDescriptor) GetVisibility() int {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.visibility
}

func copyMetadata(in map[string][]string) map[string][]string {
	retVal := make(map[string][]string)
	if in == nil {
		return retVal
	}

	for k, v := range in {
		v1 := make([]string, len(v))
		copy(v1, v)

		retVal[k] = v1
	}

	return retVal
}

func (di *baseDescriptor) GetMetadata() map[string][]string {
	di.lock.Lock()
	defer di.lock.Unlock()

	return copyMetadata(di.metadata)
}

func (di *baseDescriptor) GetRank() int32 {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.rank
}

func (di *baseDescriptor) SetRank(rank int32) int32 {
	di.lock.Lock()
	defer di.lock.Unlock()

	retVal := di.rank

	di.rank = rank

	return retVal
}

func (di *baseDescriptor) String() string {
	return fmt.Sprintf("%s/%s/%d/%d", di.namespace, di.name, di.locatorID, di.serviceID)
}

func (di *descriptorImpl) GetServiceID() int64 {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.serviceID
}

func (di *descriptorImpl) GetLocatorID() int64 {
	di.lock.Lock()
	defer di.lock.Unlock()

	return di.locatorID
}

func (wdi *writeableDescriptorImpl) SetCreateFunction(in func(ServiceLocator, Descriptor) (interface{}, error)) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	wdi.creator = in

	return nil
}

func (wdi *writeableDescriptorImpl) SetDestroyFunction(in func(ServiceLocator, Descriptor, interface{}) error) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	wdi.destroyer = in

	return nil
}

func (wdi *writeableDescriptorImpl) SetNamespace(in string) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	err := checkNamespaceCharacters(in)
	if err != nil {
		return err
	}

	wdi.namespace = in

	return nil
}

func (wdi *writeableDescriptorImpl) SetName(in string) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	err := checkNameCharacters(in)
	if err != nil {
		return err
	}

	wdi.name = in

	return nil
}

func (wdi *writeableDescriptorImpl) SetScope(in string) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	err := checkNameCharacters(in)
	if err != nil {
		return err
	}

	wdi.scope = in

	return nil
}

func (wdi *writeableDescriptorImpl) SetQualifiers(in []string) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	for _, q := range in {
		err := checkNameCharacters(q)
		if err != nil {
			return err
		}
	}

	copied := make([]string, len(in))
	copy(copied, in)

	wdi.qualifiers = copied

	return nil
}

func (wdi *writeableDescriptorImpl) SetVisibility(in int) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	if in != LocalVisibility && in != NormalVisibility {
		return fmt.Errorf("visibility %d is not LocalVisibility or Normal", in)
	}

	wdi.visibility = in

	return nil
}

func (wdi *writeableDescriptorImpl) SetMetadata(md map[string][]string) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	wdi.metadata = copyMetadata(md)

	return nil
}

func (wdi *writeableDescriptorImpl) GetCreateFunction() func(sl ServiceLocator, sk Descriptor) (interface{}, error) {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	return wdi.creator
}

func (wdi *writeableDescriptorImpl) GetDestroyFunction() func(ServiceLocator, Descriptor, interface{}) error {
	wdi.lock.Lock()
	defer wdi.lock.Unlock()

	return wdi.destroyer
}

func (wdi *writeableDescriptorImpl) GetServiceID() int64 {
	return -1
}

func (wdi *writeableDescriptorImpl) GetLocatorID() int64 {
	return -1
}

func (cdi *constantDescriptorImpl) GetCreateFunction() func(sl ServiceLocator, sk Descriptor) (interface{}, error) {
	cdi.lock.Lock()
	defer cdi.lock.Unlock()

	return func(ServiceLocator, Descriptor) (interface{}, error) {
		return cdi.cnstnt, nil
	}
}

func (cdi *constantDescriptorImpl) GetDestroyFunction() func(ServiceLocator, Descriptor, interface{}) error {
	cdi.lock.Lock()
	defer cdi.lock.Unlock()

	return func(ServiceLocator, Descriptor, interface{}) error {
		return nil
	}
}
