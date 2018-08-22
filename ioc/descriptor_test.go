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
	someKey    = "someKey"
	someValue1 = "someValue1"
	someValue2 = "someValue2"
	nonValue   = "nonValue"
)

func createBenchmarkWriteableDescriptor(t *testing.T) WriteableDescriptor {
	wd := NewWriteableDescriptor()

	qualifiers := []string{blue, green}

	metadata := make(map[string][]string)

	values := []string{someValue1, someValue2}

	metadata[someKey] = values

	err := wd.SetCreateFunction(testCreator)
	assert.Nil(t, err, "got error %v", err)
	err = wd.SetDestroyFunction(testDestroyer)
	assert.Nil(t, err, "got error %v", err)
	err = wd.SetNamespace(foo)
	assert.Nil(t, err, "got error %v", err)
	err = wd.SetName(bar)
	assert.Nil(t, err, "got error %v", err)
	err = wd.SetScope(PerLookup)
	assert.Nil(t, err, "got error %v", err)
	err = wd.SetQualifiers(qualifiers)
	assert.Nil(t, err, "got error %v", err)
	err = wd.SetVisibility(Local)
	assert.Nil(t, err, "got error %v", err)
	err = wd.SetMetadata(metadata)
	assert.Nil(t, err, "got error %v", err)

	// Post sets update the qualifiers and metadata to be sure they are copies
	qualifiers = append(qualifiers, red)
	values = append(values, nonValue)

	return wd
}

func validateBenchmarkDescriptor(t *testing.T, wd Descriptor) {
	assert.NotNil(t, wd.GetCreateFunction(), "did not get expected creator")
	assert.NotNil(t, wd.GetDestroyFunction(), "did not get expected destoyer")
	assert.Equal(t, foo, wd.GetNamespace(), "did not get expected namespace")
	assert.Equal(t, bar, wd.GetName(), "did not get expected name")
	assert.Equal(t, PerLookup, wd.GetScope(), "did not get expected scope")
	assert.Equal(t, Local, wd.GetVisibility(), "did not get expected visibility")

	assert.Equal(t, 2, len(wd.GetQualifiers()), "qualifiers should have had two elements")
	assert.Equal(t, blue, wd.GetQualifiers()[0], "zero qualifier should have been blue")
	assert.Equal(t, green, wd.GetQualifiers()[1], "one qualifier should have been green")

	assert.Equal(t, 1, len(wd.GetMetadata()), "metadata should have had one element")
	fValues := wd.GetMetadata()[someKey]
	assert.Equal(t, 2, len(fValues), "values of metadata should have been len 2")
	assert.Equal(t, someValue1, fValues[0], "zero index metadata value should have been value1")
	assert.Equal(t, someValue2, fValues[1], "one index metadata value should have been value2")
}

func TestWriteableDescriptor(t *testing.T) {
	wd := createBenchmarkWriteableDescriptor(t)

	validateBenchmarkDescriptor(t, wd)

	assert.Equal(t, int64(-1), wd.GetServiceID(), "serviceID is always -1 in this implementation")
	assert.Equal(t, int64(-1), wd.GetLocatorID(), "locatorID is always -1 in this implementation")
}

func TestReadOnlyDescriptor(t *testing.T) {
	desc := createBenchmarkWriteableDescriptor(t)

	wod, err := NewDescriptor(desc, 100, 1)
	if err != nil {
		t.Errorf("Error creating new descriptor %v", err)
		return
	}

	validateBenchmarkDescriptor(t, wod)

	assert.Equal(t, int64(100), wod.GetServiceID(), "serviceID should have been 100")
	assert.Equal(t, int64(1), wod.GetLocatorID(), "locatorID should have been 1")
}

type iFace3 interface{}

func TestConstantDescriptor(t *testing.T) {
	i1 := new(iFace3)

	key := DSK(foo, red, green)

	cDesc := NewConstantDescriptor(key, i1)

	fCreate := cDesc.GetCreateFunction()

	i2, err2 := fCreate(nil, key)
	if err2 != nil {
		t.Errorf("Could not call create method from descriptor %v", err2)
	}

	if i1 != i2 {
		t.Errorf("Did not get my original constant back! %v/%v", i1, i2)
	}
}

func testCreator(locator ServiceLocator, key ServiceKey) (interface{}, error) {
	return nil, nil
}

func testDestroyer(locator ServiceLocator, key ServiceKey, killMe interface{}) error {
	return nil
}
