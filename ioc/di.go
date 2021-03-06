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
	"reflect"
)

type diData struct {
	ty      reflect.Type
	locator *serviceLocatorData
}

func newCreatorFunc(ty reflect.Type, parent *serviceLocatorData) func(ServiceLocator, Descriptor) (interface{}, error) {
	diVal := &diData{
		ty:      ty,
		locator: parent,
	}

	retVal := func(locator ServiceLocator, desc Descriptor) (interface{}, error) {
		return diVal.create(locator, desc)
	}

	return retVal
}

type indexAndValueOfDependency struct {
	index int
	value *reflect.Value
}

type errorReturn struct {
	err error
}

func (di *diData) create(rawLocator ServiceLocator, desc Descriptor) (interface{}, error) {
	locator, ok := rawLocator.(*serviceLocatorData)
	if !ok {
		return nil, fmt.Errorf("unknown service locator type")
	}

	return createAndInject(locator, desc, di.ty, nil)
}

func isProvider(ty reflect.Type) bool {
	if ty == nil {
		return false
	}

	providerType := reflect.TypeOf((*Provider)(nil)).Elem()
	return ty.AssignableTo(providerType)
}

func safeSet(v reflect.Value, to *reflect.Value, ret *errorReturn) {
	defer func() {
		if r := recover(); r != nil {
			ret.err = fmt.Errorf("%v", r)
		}
	}()

	v.Set(*to)
}

func safeDargoInitialize(dargoI DargoInitializer, desc Descriptor, ret *errorReturn) {
	defer func() {
		if r := recover(); r != nil {
			ret.err = fmt.Errorf("%v", r)
		}
	}()

	ret.err = dargoI.DargoInitialize(desc)
}

type hasRunErrorHandlersError interface {
	error
	GetHasRunErrorHandlers() bool
	GetUnderlyingError() MultiError
}

type hasRunHandlers struct {
	hasRunHandlers  bool
	underlyingError MultiError
}

func (hrh *hasRunHandlers) Error() string {
	return hrh.underlyingError.Error()
}

func (hrh *hasRunHandlers) GetHasRunErrorHandlers() bool {
	return hrh.hasRunHandlers
}

func (hrh *hasRunHandlers) GetUnderlyingError() MultiError {
	return hrh.underlyingError
}

type providerData struct {
	locator *serviceLocatorData
	key     ServiceKey
	mother  Descriptor
}

func newProvider(locator *serviceLocatorData, serviceKey ServiceKey, mother Descriptor) Provider {
	return &providerData{
		locator: locator,
		key:     serviceKey,
		mother:  mother,
	}
}

func (pd *providerData) Get() (interface{}, error) {
	return pd.locator.getServiceFor(pd.key, pd.mother)
}

func (pd *providerData) GetAll() ([]interface{}, error) {
	return pd.locator.getAllServicesFor(pd.key, pd.mother)
}

func (pd *providerData) QualifiedBy(qualifier string) Provider {
	currentQualifiers := pd.key.GetQualifiers()
	currentQualifiers = append(currentQualifiers, qualifier)

	serviceKey, err := NewServiceKey(pd.key.GetNamespace(), pd.key.GetName(), currentQualifiers...)
	if err != nil {
		panic(err.Error())
	}

	return newProvider(pd.locator, serviceKey, pd.mother)
}
