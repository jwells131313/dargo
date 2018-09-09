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
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ValidationTestLocatorName1 = "ValidationTestLocator1"
	ValidationTestLocatorName2 = "ValidationTestLocator2"
	ValidationTestLocatorName3 = "ValidationTestLocator3"
	ValidationTestLocatorName4 = "ValidationTestLocator4"

	DoNotBindService   = "DoNotBindService"
	NeverUnbindService = "NeverUnbindService"
	SimpleServiceName  = "SimpleService"

	ClientOrServerService = "ClientOrServerService"
	Server                = "ServerQualifier"
	Client                = "ClientQualifier"

	NoServerString = "no server allowed in client environment"
)

var isServer bool

func TestBindValidation(t *testing.T) {
	locator, err := CreateAndBind(ValidationTestLocatorName1, func(binder Binder) error {
		binder.Bind(ValidationServiceName, ValidationServiceData{}).InNamespace(UserServicesNamespace)
		return nil
	})
	if !assert.Nil(t, err, "error creating locator") {
		return
	}

	err = BindIntoLocator(locator, func(binder Binder) error {
		binder.Bind(DoNotBindService, InvalidService{})
		binder.Bind(SimpleServiceName, SimpleService{})
		return nil
	})

	if !assert.NotNil(t, err, "should have failed to bind") {
		return
	}

	// Make sure we can't get SimpleService now, since it was in Bind that failed
	raw, err := locator.GetDService(SimpleServiceName)
	if !assert.NotNil(t, err, "we throw an error if service is not there") {
		return
	}

	assert.Nil(t, raw, "there should be no SimpleService")
}

func TestUnBindValidation(t *testing.T) {
	locator, err := CreateAndBind(ValidationTestLocatorName2, func(binder Binder) error {
		binder.Bind(ValidationServiceName, ValidationServiceData{}).InNamespace(UserServicesNamespace)
		binder.Bind(NeverUnbindService, SimpleService{}).InScope(PerLookup)
		return nil
	})
	if !assert.Nil(t, err, "error creating locator") {
		return
	}

	raw, err := locator.GetDService(NeverUnbindService)
	if !assert.Nil(t, err, "service never to be unbound should not be unbound") {
		return
	}

	if !assert.NotNil(t, raw, "got nil NeverUnbind service") {
		return
	}

	err = UnbindDServices(locator, NeverUnbindService)
	if !assert.NotNil(t, err, "should not have been able to unbind the service") {
		return
	}

	raw, err = locator.GetDService(NeverUnbindService)
	if !assert.Nil(t, err, "service never to be unbound should not be unbound (2)") {
		return
	}

	if !assert.NotNil(t, raw, "got nil NeverUnbind service (2)") {
		return
	}

}

func TestLookukpValidation(t *testing.T) {
	locator, err := CreateAndBind(ValidationTestLocatorName3, func(binder Binder) error {
		binder.Bind(ValidationServiceName, ValidationServiceData{}).InNamespace(UserServicesNamespace)
		binder.Bind(ClientOrServerService, ClientService{}).QualifiedBy(Client).InScope(PerLookup)
		binder.Bind(ClientOrServerService, ServerService{}).QualifiedBy(Server).InScope(PerLookup)
		return nil
	})
	if !assert.Nil(t, err, "error creating locator") {
		return
	}

	isServer = false

	_, err = getClientService(t, locator)
	if !assert.Nil(t, err, "in client should have been able to see client") {
		return
	}

	_, err = getServerService(t, locator)
	if !assert.NotNil(t, err, "should think there is no server service") {
		return
	}

	cos, err := locator.GetAllServices(DSK(ClientOrServerService))
	if !assert.Nil(t, err, "should have gotten one service") {
		return
	}

	if !assert.Equal(t, 1, len(cos), "should be only one service") {
		return
	}

	_, ok := cos[0].(*ClientService)
	if !assert.True(t, ok, "One service returned should be client") {
		return
	}

	isServer = true

	_, err = getClientService(t, locator)
	if !assert.NotNil(t, err, "in server should not have been able to see client") {
		return
	}

	_, err = getServerService(t, locator)
	if !assert.Nil(t, err, "should have been able to see server service") {
		return
	}

	sin, err := locator.GetAllServices(DSK(ClientOrServerService))
	if !assert.Nil(t, err, "should have gotten one service") {
		return
	}

	if !assert.Equal(t, 1, len(sin), "should be only one service") {
		return
	}

	_, ok = sin[0].(*ServerService)
	if !assert.True(t, ok, "One service returned should be server") {
		return
	}

}

func TestLookukpValidationErrorService(t *testing.T) {
	locator, err := CreateAndBind(ValidationTestLocatorName4, func(binder Binder) error {
		binder.Bind(ValidationServiceName, ValidationServiceData{}).InNamespace(UserServicesNamespace)
		binder.Bind(ClientOrServerService, ServerService{}).QualifiedBy(Server).InScope(PerLookup)
		binder.Bind(ErrorServiceName, ValidationErrorServiceData{}).InNamespace(UserServicesNamespace)
		return nil
	})
	if !assert.Nil(t, err, "error creating locator") {
		return
	}

	lastValidationErrorInformation = nil
	isServer = false

	_, err = getServerService(t, locator)
	if !assert.NotNil(t, err, "should think there is no server service") {
		return
	}

	assert.NotNil(t, lastValidationErrorInformation, "error service not called")

	assert.Equal(t, LookupValidationFailure, lastValidationErrorInformation.GetType(),
		"incorrect type")
	assert.Equal(t, ClientOrServerService, lastValidationErrorInformation.GetDescriptor().GetName(),
		"incorrect descriptor %v", lastValidationErrorInformation.GetDescriptor())
	assert.Nil(t, lastValidationErrorInformation.GetInjectee())
	assert.Equal(t, NoServerString, lastValidationErrorInformation.GetAssociatedError().Error(),
		"invalid error")
}

func getClientService(t *testing.T, locator ServiceLocator) (*ClientService, error) {
	raw, err := locator.GetDService(ClientOrServerService, Client)
	if err != nil {
		return nil, err
	}

	client, ok := raw.(*ClientService)
	if !assert.True(t, ok, "incorrect type for client service") {
		return nil, err
	}

	return client, nil
}

func getServerService(t *testing.T, locator ServiceLocator) (*ServerService, error) {
	raw, err := locator.GetDService(ClientOrServerService, Server)
	if err != nil {
		return nil, err
	}

	server, ok := raw.(*ServerService)
	if !assert.True(t, ok, "incorrect type for server service") {
		return nil, err
	}

	return server, nil
}

type ValidationServiceData struct {
}

func (vsd *ValidationServiceData) GetFilter() Filter {
	// We check everything
	return AllFilter
}

func (vsd *ValidationServiceData) GetValidator() Validator {
	return vsd
}

func (vsd *ValidationServiceData) Validate(info ValidationInformation) error {
	switch info.GetOperation() {
	case BindOperation:
		if DoNotBindService == info.GetCandidate().GetName() {
			return fmt.Errorf("we will not bind %v", info.GetCandidate())
		}
		break
	case UnbindOperation:
		if NeverUnbindService == info.GetCandidate().GetName() {
			return fmt.Errorf("we will not unbind %v", info.GetCandidate())
		}
		break
	case LookupOperation:
		if isServer {
			if hasQualifier(Client, info.GetCandidate()) {
				return fmt.Errorf("no client allowed in server environment")
			}
		} else {
			if hasQualifier(Server, info.GetCandidate()) {
				return fmt.Errorf(NoServerString)
			}
		}
		break
	default:
		return fmt.Errorf("unexpected operation %s", info.GetOperation())
	}

	return nil
}

func hasQualifier(qualifier string, desc Descriptor) bool {
	for _, q := range desc.GetQualifiers() {
		if qualifier == q {
			return true
		}
	}

	return false
}

type SimpleService struct {
}

type InvalidService struct {
}

type ClientService struct {
}

type ServerService struct {
}

type UsesAClientServiceData struct {
	InjectedClientService *ClientService `inject:"ClientOrServerService@Client"`
}

type ValidationErrorServiceData struct {
}

var lastValidationErrorInformation ErrorInformation

func (vesd *ValidationErrorServiceData) OnFailure(ei ErrorInformation) error {
	lastValidationErrorInformation = ei
	return nil
}
