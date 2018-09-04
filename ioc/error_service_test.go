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

var lastErrorInformation ErrorInformation

func (esd *errorServiceData) OnFailure(ei ErrorInformation) error {
	lastErrorInformation = ei
	return ei.GetAssociatedError()
}

func TestCreationError(t *testing.T) {
	locator, err := CreateAndBind("TestCreationErrorLocator", func(binder Binder) error {
		binder.BindWithCreator("ServiceB", serviceBCreator)
		binder.Bind("ServiceA", ServiceA{})
		binder.Bind(ErrorServiceName, errorServiceData{}).InNamespace(UserServicesNamespace)

		return nil
	})
	if err != nil {
		assert.NotNil(t, err, "could not create locator")
		return
	}

	raw, err := locator.GetDService("ServiceA")
	assert.Nil(t, raw, "serviceA should be nil")

	assert.NotNil(t, lastErrorInformation, "LastError information should not be nil")

	assert.Equal(t, ServiceCreationFailure, lastErrorInformation.GetType(), "should be ServiceCreationFailure")
	assert.Equal(t, err, lastErrorInformation.GetAssociatedError(), "should be same error")

	desc := lastErrorInformation.GetDescriptor()
	assert.NotNil(t, desc, "should have an associated descriptor")

	assert.Equal(t, "ServiceA", desc.GetName(), "A failed it should be the descriptor that shows up")

	multi, ok := err.(MultiError)
	assert.True(t, ok, "returned error must be multi error")

	found := false
	for _, internalErr := range multi.GetErrors() {
		if internalErr.Error() == expectedErrorString {
			found = true
			break
		}
	}

	assert.True(t, found, "Did not get the expected error up through the stack")

}
