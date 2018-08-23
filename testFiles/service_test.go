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

package testFiles

import (
	"github.com/jwells131313/dargo/ioc"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ServiceTestLocatorName = "ServiceTestLocatorName"
	EchoServiceName        = "EchoService"
)

// TestBasicServiceLocatorLookup.  This uses the raw DynamicConfigurationService
// in order to add a service (echo) and then look up the service
func TestAddServiceWithDCS(t *testing.T) {
	locator, err := ioc.NewServiceLocator(ServiceTestLocatorName, ioc.FailIfPresent)
	assert.Nil(t, err, "error creating locator")

	wd := ioc.NewWriteableDescriptor()
	wd.SetName(EchoServiceName)
	wd.SetCreateFunction(createEcho)

	dcs, err := getDCS(t, locator)
	if err != nil {
		return
	}

	config, err := dcs.CreateDynamicConfiguration()
	assert.Nil(t, err, "could not create a dynamic configuration")

	wdb, err := config.Bind(wd)
	assert.Nil(t, err, "could not bind user descriptor")

	assert.True(t, wdb.GetLocatorID() >= 0, "Got incorrect locator ID")
	assert.True(t, wdb.GetServiceID() > 0, "Gog incorrect service ID")

	err = config.Commit()
	assert.Nil(t, err, "commit failed")

	// Now lets get the service
	if false {
		t.Log("Fix this test once we have a nice cache for singleton service")
		return
	}

	raw, err := locator.GetService(ioc.DSK(EchoServiceName))
	assert.Nil(t, err, "error getting the user service")
	assert.NotNil(t, raw, "returned service is nil")

	echoService, ok := raw.(EchoApplication)
	assert.True(t, ok, "Returned service is not an EchoApplication")

	reply := echoService.Echo("hi")
	assert.Equal(t, reply, "hi", "Echo didn't echo?")

}

func getDCS(t *testing.T, locator ioc.ServiceLocator) (ioc.DynamicConfigurationService, error) {
	raw, err := locator.GetService(ioc.SSK(ioc.DynamicConfigurationServiceName))
	assert.Nil(t, err, "could not get dynamic configuration service")

	dcs, ok := raw.(ioc.DynamicConfigurationService)
	assert.True(t, ok, "Service returned is not a dynamic configuration service")

	return dcs, nil
}

func createEcho(locator ioc.ServiceLocator, key ioc.ServiceKey) (interface{}, error) {
	return NewEchoApplication(), nil
}

func destroyEcho(locator ioc.ServiceLocator, key ioc.ServiceKey, obj interface{}) error {
	return nil
}
