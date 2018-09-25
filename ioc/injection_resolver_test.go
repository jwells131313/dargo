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
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

const (
	injectionResolverTest1 = "InjectionResolverTest1"
	injectionResolverTest2 = "InjectionResolverTest2"

	alternateServiceName = "AlternateService"
	normalServiceName    = "NormalService"
)

func TestAlternateInject(t *testing.T) {
	locator, err := CreateAndBind(injectionResolverTest1, func(binder Binder) error {
		binder.Bind(InjectionResolverName, &AlternateInjectionResolver{}).InNamespace(UserServicesNamespace)
		binder.Bind(SimpleServiceName, &SimpleService{})
		binder.Bind(alternateServiceName, &AlternateService{})

		return nil
	})
	if !assert.Nil(t, err, "could not create locator") {
		return
	}

	raw, err := locator.GetDService(alternateServiceName)
	if !assert.Nil(t, err, "could not find alt service") {
		return
	}

	alt := raw.(*AlternateService)

	assert.NotNil(t, alt.SS, "Did not have alternate injection point")
}

func TestOverrideSystemInject(t *testing.T) {
	locator, err := CreateAndBind(injectionResolverTest2, func(binder Binder) error {
		binder.Bind(InjectionResolverName, &OverrideInjectResolver{}).Ranked(1).InNamespace(UserServicesNamespace)
		binder.Bind(SimpleServiceName, &SimpleService{})
		binder.Bind(normalServiceName, &NormalService{})

		return nil
	})
	if !assert.Nil(t, err, "could not create locator") {
		return
	}

	raw, err := locator.GetDService(normalServiceName)
	if !assert.Nil(t, err, "could not find normal service %v", err) {
		return
	}

	norm := raw.(*NormalService)

	assert.NotNil(t, norm.SS, "Did not have alternate injection point")

	rawResolver, err := locator.GetService(USK(InjectionResolverName))
	if !assert.Nil(t, err, "could not find resolver service") {
		return
	}

	override := rawResolver.(*OverrideInjectResolver)

	if !assert.Equal(t, 1, len(override.injectees), "should have been one resolution") {
		return
	}

	injectee := override.injectees[0]
	desc := injectee.GetDescriptor()

	assert.Equal(t, normalServiceName, desc.GetName())
}

type AlternateInjectionResolver struct {
}

func (air *AlternateInjectionResolver) Resolve(locator ServiceLocator, injectee Injectee) (*reflect.Value, bool, error) {
	fieldVal := injectee.GetField()

	injectString := fieldVal.Tag.Get("alternate")

	if injectString != "" {
		serviceKey, err := parseInjectString(injectString)
		if err != nil {
			return nil, false, err
		}

		var dependency interface{}
		dependency, err = locator.GetService(serviceKey)

		if err != nil {
			return nil, false, err
		}

		dependencyAsValue := reflect.ValueOf(dependency)

		return &dependencyAsValue, true, nil
	}

	return nil, false, nil
}

type AlternateService struct {
	SS *SimpleService `alternate:"SimpleService"`
}

type OverrideInjectResolver struct {
	SystemResolver InjectionResolver `inject:"user/services#InjectionResolver@SystemInjectResolverQualifier"`
	injectees      []Injectee
}

func (oir *OverrideInjectResolver) Resolve(locator ServiceLocator, injectee Injectee) (*reflect.Value, bool, error) {
	if oir.injectees == nil {
		oir.injectees = make([]Injectee, 0)
	}

	oir.injectees = append(oir.injectees, injectee)

	return oir.SystemResolver.Resolve(locator, injectee)
}

type NormalService struct {
	SS *SimpleService `inject:"SimpleService"`
}
