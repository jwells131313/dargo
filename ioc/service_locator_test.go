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
	"testing"
)

const (
	testLocatorName  = "TestLocator"
	testLocatorName2 = "TestLocator2"
)

func TestGetSystemServices(t *testing.T) {
	locator, err := NewServiceLocator(testLocatorName, ReturnExistingOrCreateNew)
	assert.Nil(t, err, "Could not create/find locator")

	foundService, err := locator.GetService(SSK(ServiceLocatorName))
	assert.Nil(t, err, "Did not find service locator service")

	foundLocator, ok := foundService.(ServiceLocator)
	assert.True(t, ok, "Does not implement ServiceLocator")

	assert.Equal(t, locator, foundLocator, "The locators are not the same")
	assert.Equal(t, testLocatorName, foundLocator.GetName(), "Invalid name")

	locatorID := locator.GetID()
	assert.True(t, locatorID > 0, "locator id should be greater than zero %d", locatorID)

	foundService, err = locator.GetService(SSK(DynamicConfigurationServiceName))
	assert.Nil(t, err, "Could not find dynamic configuration service")

	_, ok = foundService.(DynamicConfigurationService)
	assert.True(t, ok, "does not implement DynamicConfigurationService")

	locator2, err := NewServiceLocator(testLocatorName2, FailIfPresent)
	assert.Nil(t, err, "Could not create/find locator2")

	assert.NotEqual(t, locator2, locator, "The locators are not the same")
	assert.Equal(t, testLocatorName2, locator2.GetName(), "Invalid name")

	locatorID2 := locator2.GetID()
	assert.True(t, locatorID2 > 0, "locator id should be greater than zero %d", locatorID2)
	assert.NotEqual(t, locatorID, locatorID2, "The locator serviceID should not be equal")

	locator3, err := NewServiceLocator(testLocatorName, FailIfNotPresent)
	assert.Equal(t, locator, locator3, "Already existing locator returned")
}
