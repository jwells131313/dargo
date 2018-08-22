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

/*

import (
	"github.com/jwells131313/dargo/api"
	"github.com/jwells131313/dargo/utilities"
	"reflect"
	"testing"
)

// TestBasicServiceLocatorLookup.  This uses the raw DynamicConfigurationService
// in order to add a service (echo) and then look up the service
func TestBasicServiceLocatorLookup(t *testing.T) {
	locatorFactory := utilities.GetSystemLocatorFactory()

	locator, found := locatorFactory.FindOrCreateRootLocator("BasicServiceLocatorLookup")
	defer locator.Shutdown()

	if found != false {
		t.Errorf("There was an error creating BasicServiceLocatorLookup service locator %v", locator)
		return
	}

	dynamicConfigurationServiceRaw, err2 := locator.GetService(reflect.TypeOf(new(api.DynamicConfigurationService)).Elem())
	if dynamicConfigurationServiceRaw == nil {
		t.Errorf("There must always be an implementation of the DCS in any ServiceLocator")
		return
	}
	if err2 != nil {
		t.Errorf("There was an error getting the service %v", err2)
		return
	}

	dynamicConfigurationService, ok := dynamicConfigurationServiceRaw.(api.DynamicConfigurationService)

	if !ok {
		t.Errorf("Could not do the cast")
		return
	}

	dConfig := dynamicConfigurationService.CreateDynamicConfiguration()
	if dConfig == nil {
		t.Errorf("Got a nil dynamic configuration, that's a fail")
		return
	}

	if true {
		t.Log("Below not yet implemented")
		return
	}

	wd, err := api.Bind(func(locator api.ServiceLocator) (interface{}, error) {
		return NewEchoApplication(), nil
	}).Build()
	if err != nil {
		t.Error("Could not bind the echo application", err)
		return
	}

	sd := dConfig.Bind(wd)
	if sd == nil {
		t.Error("Should have returned read-only system descriptor")
		return
	}

	err = dConfig.Commit()
	if err == nil {
		t.Error("DConfig commit failed", err)
		return
	}

	esi, err3 := locator.GetService(reflect.TypeOf(new(EchoApplication)).Elem())
	if err3 != nil {
		t.Error("Could not find EchoApplication", err3)
		return
	}

	ea, ok := esi.(EchoApplication)
	if !ok {
		t.Error("Could not cast returned object to EchoApplication")
		return
	}

	ret := ea.Echo("hello")
	if ret != "hello" {
		t.Errorf("Expected hello got %s", ret)
		return
	}
}
*/
