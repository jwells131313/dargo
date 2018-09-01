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
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync/atomic"
	"testing"
)

const (
	testDargoContextLocator1 = "TestDargoContextLocator1"
	testDargoContextLocator2 = "TestDargoContextLocator2"
	testDargoContextLocator3 = "TestDargoContextLocator3"

	testDargoService   = "testDargoService"
	testToUpperService = "testToUpperService"
)

var creationGeneration int32

type testDargoContextHolder struct {
	creationNumber int32
	context        context.Context
	destroyed      bool
}

func TestSingleCallToDargoContext(t *testing.T) {
	locator, err := CreateAndBind(testDargoContextLocator1, func(binder Binder) error {
		binder.BindWithCreator(testDargoService, createDargoService).InScope(ContextScope)
		binder.BindWithCreator(testToUpperService, createToUpperService)
		return nil
	})
	if err != nil {
		t.Errorf("could not create locator %v", err)
		return
	}

	EnableContextScope(locator)

	bg := context.Background()

	myContext, err := createDargoContext(bg, t, locator)
	if err != nil {
		t.Errorf("could not create context %v", err)
		return
	}

	tusRaw := myContext.Value(DSK(testToUpperService))
	assert.NotNil(t, tusRaw, "Should have found service from the context")

	tus := tusRaw.(toUpper)

	input := "go eagles"
	expected := strings.ToUpper(input)

	assert.Equal(t, expected, tus.toUpper(input), "did not get expected result")
}

func TestManyDargoContexts(t *testing.T) {
	creationGeneration = 0

	locator, err := CreateAndBind(testDargoContextLocator2, func(binder Binder) error {
		binder.BindWithCreator(testDargoService, createDargoService).InScope(ContextScope)
		return nil
	})
	if err != nil {
		t.Errorf("could not create locator %v", err)
		return
	}

	EnableContextScope(locator)

	bg := context.Background()

	myContext1, err := createDargoContext(bg, t, locator)
	if err != nil {
		t.Errorf("could not create context %v", err)
		return
	}

	myContext2, err := createDargoContext(bg, t, locator)
	if err != nil {
		t.Errorf("could not create context %v", err)
		return
	}

	myContext3, err := createDargoContext(bg, t, locator)
	if err != nil {
		t.Errorf("could not create context %v", err)
		return
	}

	ret1, err := getDargoServiceValue(myContext1)
	if err != nil {
		t.Errorf("value 1 was not returned %s", err.Error())
		return
	}

	ret2, err := getDargoServiceValue(myContext2)
	if err != nil {
		t.Errorf("value 2 was not returned %s", err.Error())
		return
	}

	ret3, err := getDargoServiceValue(myContext3)
	if err != nil {
		t.Errorf("value 2 was not returned %s", err.Error())
		return
	}

	hasValue(t, 1, ret1, ret2, ret3)
	hasValue(t, 2, ret1, ret2, ret3)
	hasValue(t, 3, ret1, ret2, ret3)
}

func TestDargoContextCancel(t *testing.T) {
	parentContext1, canceller1 := context.WithCancel(context.Background())
	defer canceller1()

	parentContext2, canceller2 := context.WithCancel(context.Background())
	defer canceller2()

	locator, err := CreateAndBind(testDargoContextLocator3, func(binder Binder) error {
		binder.BindWithCreator(testDargoService, createDargoService).
			InScope(ContextScope).
			AndDestroyWith(destroyDargoService)

		return nil
	})
	if err != nil {
		t.Errorf("could not create locator %v", err)
		return
	}

	EnableContextScope(locator)

	context1, err := createDargoContext(parentContext1, t, locator)
	if err != nil {
		t.Errorf("could not create context 1 %v", err)
		return
	}

	// context 2 is not destroyed to ensure its services are not destroyed when 1 dies
	context2, err := createDargoContext(parentContext2, t, locator)
	if err != nil {
		t.Errorf("could not create context 2  %v", err)
		return
	}

	ds, err := getDargoService(context1)
	if err != nil {
		t.Errorf("could not get test dargo service %v", err)
		return
	}

	ds1, err := getDargoService(context2)
	if err != nil {
		t.Errorf("could not get test dargo service %v", err)
		return
	}

	assert.False(t, ds.destroyed, "Have not cancelled context yet so should not be destroyed")
	assert.False(t, ds1.destroyed, "Have not cancelled context yet so should not be destroyed ds1")

	canceller1()

	// Wait for it to be done
	<-context1.Done()

	assert.True(t, ds.destroyed, "Cancelled and waited, so... yeah, should be destroyed now")
	assert.False(t, ds1.destroyed, "Have not cancelled context yet so should not be destroyed ds1")

	canceller2()

	// Wait for it to be done
	<-context2.Done()

	assert.True(t, ds.destroyed, "Cancelled and waited, so... yeah, should be destroyed now")
	assert.True(t, ds1.destroyed, "Have not cancelled context yet so should not be destroyed ds1")

}

func hasValue(t *testing.T, expected int32, a, b, c int32) {
	if a != expected && b != expected && c != expected {
		t.Errorf("There was no expected return value of %d.  Instead got %d,%d,%d", expected, a, b, c)
	}
}

func getDargoServiceValue(context context.Context) (int32, error) {
	ds, err := getDargoService(context)
	if err != nil {
		return -1, errors.Wrap(err, "Could not get dargo service")
	}

	return ds.creationNumber, nil
}

func getDargoService(context context.Context) (*testDargoContextHolder, error) {
	raw := context.Value(testDargoService)
	if raw == nil {
		return nil, fmt.Errorf("Could not find the testDargoService from context %v", context)
	}

	ds, ok := raw.(*testDargoContextHolder)
	if !ok {
		return nil, fmt.Errorf("TestDargoService was not the correct type")
	}

	return ds, nil
}

func createDargoContext(parentContext context.Context, t *testing.T, locator ServiceLocator) (context.Context, error) {
	retVal, err := NewDargoContext(parentContext, locator)
	if err != nil {
		assert.NotNil(t, err, "Could not create new DargoContext")
		return nil, err
	}

	dsRaw := retVal.Value(testDargoService)
	if dsRaw == nil {
		assert.NotNil(t, dsRaw, "Did not find bound testDargoService")
		return nil, fmt.Errorf("did not find bound testDargoService")
	}

	ds, ok := dsRaw.(*testDargoContextHolder)
	assert.True(t, ok, "raw DS service was not of the correct type")

	ds.context = retVal

	return retVal, nil
}

func createDargoService(locator ServiceLocator, key Descriptor) (interface{}, error) {
	val := atomic.AddInt32(&creationGeneration, 1)

	return &testDargoContextHolder{
		creationNumber: val,
	}, nil
}

func destroyDargoService(locator ServiceLocator, key Descriptor, raw interface{}) error {
	ds, ok := raw.(*testDargoContextHolder)
	if !ok {
		return fmt.Errorf("Incorrect type to destroy")
	}

	ds.destroyed = true

	return nil
}

func (tdch *testDargoContextHolder) String() string {
	return fmt.Sprintf("testDargoContextHolder(%d)", tdch.creationNumber)
}

type toUpper interface {
	toUpper(string) string
}

type toUpperData struct{}

func createToUpperService(locator ServiceLocator, key Descriptor) (interface{}, error) {
	return &toUpperData{}, nil
}

func (tud *toUpperData) toUpper(in string) string {
	return strings.ToUpper(in)
}
