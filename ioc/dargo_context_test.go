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
	"github.com/stretchr/testify/assert"
	"strings"
	"sync/atomic"
	"testing"
)

const (
	testDargoContextLocator1 = "TestDargoContextLocator1"
	testDargoService         = "testDargoService"
	testToUpperService       = "testToUpperService"
)

var creationGeneration int32

type testDargoContextHolder struct {
	creationNumber int32
	context        context.Context
}

func TestSingleCallToDargoContext(t *testing.T) {
	locator, err := CreateAndBind(testDargoContextLocator1, func(binder Binder) error {
		binder.Bind(testDargoService, createDargoService).InScope(ContextScope)
		binder.Bind(testToUpperService, createToUpperService)
		return nil
	})
	if err != nil {
		t.Errorf("could not create locator %v", err)
		return
	}

	EnableContextScope(locator)

	DumpAllDescriptors(locator)

	bg := context.Background()

	myContext, err := createDargoContext(t, locator, bg)
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

func createDargoContext(t *testing.T, locator ServiceLocator, parentContext context.Context) (context.Context, error) {
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

func createDargoService(locator ServiceLocator, key ServiceKey) (interface{}, error) {
	val := atomic.AddInt32(&creationGeneration, 1)

	return &testDargoContextHolder{
		creationNumber: val,
	}, nil
}

type toUpper interface {
	toUpper(string) string
}

type toUpperData struct{}

func createToUpperService(locator ServiceLocator, key ServiceKey) (interface{}, error) {
	return &toUpperData{}, nil
}

func (tud *toUpperData) toUpper(in string) string {
	return strings.ToUpper(in)
}
