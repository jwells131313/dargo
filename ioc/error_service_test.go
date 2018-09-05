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
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

const (
	expectedErrorString = "expected error creating serviceB"
)

type ServiceB struct {
}

type ServiceA struct {
	B ServiceB `inject:"ServiceB"`
}

func serviceBCreator(locator ServiceLocator, desc Descriptor) (interface{}, error) {
	return nil, errors.New(expectedErrorString)
}

type errorServiceData struct{}

var lastErrorInformation []ErrorInformation

func (esd *errorServiceData) OnFailure(ei ErrorInformation) error {
	lastErrorInformation = append(lastErrorInformation, ei)

	return ei.GetAssociatedError()
}

func TestCreationError(t *testing.T) {
	lastErrorInformation = make([]ErrorInformation, 0)

	locator, err := CreateAndBind("TestCreationErrorLocator", func(binder Binder) error {
		binder.BindWithCreator("ServiceB", serviceBCreator)
		binder.Bind("ServiceA", ServiceA{})
		binder.Bind(ErrorServiceName, errorServiceData{}).InNamespace(UserServicesNamespace)

		return nil
	})
	if !assert.Nil(t, err, "could not create locator") {
		return
	}

	raw, err := locator.GetDService("ServiceA")
	if !assert.Nil(t, raw, "serviceA should be nil") {
		return
	}

	if !assert.Equal(t, 2, len(lastErrorInformation), "should be two errors") {
		return
	}

	ei0 := lastErrorInformation[0]

	if !checkErrorInformation(t, ei0, ServiceCreationFailure, "ServiceB",
		nil, expectedErrorString) {
		return
	}

	ei1 := lastErrorInformation[1]

	expectedType := reflect.TypeOf(ServiceA{})

	checkErrorInformation(t, ei1, ServiceCreationFailure, "ServiceA",
		expectedType, "an error occurred while getting the dependencies of")
}

func checkErrorInformation(t *testing.T, ei ErrorInformation, infoType string,
	descriptorName string, typ reflect.Type, expectedError string) bool {
	if !assert.NotNil(t, ei, "ErrorInformation should not be nil") {
		return false
	}

	if !assert.Equal(t, infoType, ei.GetType(), "unexpected error information type") {
		return false
	}

	if descriptorName != "" {
		if !assert.Equal(t, descriptorName, ei.GetDescriptor().GetName(), "unexpected descriptor name") {
			return false
		}
	}

	if typ != nil {
		if !assert.Equal(t, typ, ei.GetInjectee(), "type is not the same") {
			return false
		}
	} else {
		if !assert.Nil(t, ei.GetInjectee(), "injectee should be nil") {
			return false
		}
	}

	if expectedError != "" {
		rawError := ei.GetAssociatedError()
		multi, ok := rawError.(MultiError)
		if !assert.True(t, ok, "unexpected type of error from GetAssociatedError") {
			return false
		}

		var found bool
		for _, internalErr := range multi.GetErrors() {
			if strings.Contains(internalErr.Error(), expectedError) {
				found = true
				break
			}
		}

		if !assert.True(t, found, "did not find expected string %s in %v", expectedError, rawError) {
			return false
		}
	}

	return true
}
