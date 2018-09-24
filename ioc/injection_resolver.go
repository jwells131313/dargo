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
	"reflect"
)

// Injectee describes a potential injection point that an injection
// resolver might want to inject a value into
type Injectee interface {
	// GetDescriptor returns the Descriptor being created
	GetDescriptor() Descriptor
	// GetType returns the structure type being injected into
	GetType() reflect.Type
	// GetField returns the structure field that is being considered for injection
	GetField() reflect.StructField
}

// InjectionResolver is a service that must be in namespace ioc.UserServicesNamespace
// and have name ioc.InjectionResolverName.  It will be created as soon as it is bound
// into the ServiceLocator so anything it depends on will also be instantiated right
// away.  The system injection resolver which handles structure fields that have
// the annotation "inject" will be qualified with the value
type InjectionResolver interface {
	// Resolve This method must return the value
	// that should be injected, true if the returned value should be used (which allows
	// for nil injection values) and an error if there should be an error associated
	// with this injection point.  All injection resolvers will be called until one
	// of them returns a useable value.  If no injection resolver returns true then
	// that field is not injected into
	Resolve(ServiceLocator, Injectee) (*reflect.Value, bool, error)
}

type injecteeData struct {
	desc  Descriptor
	typ   reflect.Type
	field reflect.StructField
}

func newInjectee(desc Descriptor, typ reflect.Type, field reflect.StructField) Injectee {
	return &injecteeData{
		desc:  desc,
		typ:   typ,
		field: field,
	}
}

func (i *injecteeData) GetDescriptor() Descriptor {
	return i.desc
}

func (i *injecteeData) GetType() reflect.Type {
	return i.typ
}

func (i *injecteeData) GetField() reflect.StructField {
	return i.field
}

type systemInjectionResolver struct {
}

func (sir *systemInjectionResolver) Resolve(locator ServiceLocator, injectee Injectee) (*reflect.Value, bool, error) {
	iLocator := locator.(*serviceLocatorData)

	fieldVal := injectee.GetField()
	desc := injectee.GetDescriptor()

	injectString := fieldVal.Tag.Get("inject")

	if injectString != "" {
		serviceKey, err := parseInjectString(injectString)
		if err != nil {
			return nil, false, err
		}

		fieldType := fieldVal.Type

		var dependency interface{}
		if !isProvider(fieldType) {
			dependency, err = iLocator.getServiceFor(serviceKey, desc)
		} else {
			dependency = newProvider(iLocator, serviceKey, desc)
		}

		if err != nil {
			return nil, false, err
		}

		dependencyAsValue := reflect.ValueOf(dependency)

		return &dependencyAsValue, true, nil
	}

	return nil, false, nil
}
