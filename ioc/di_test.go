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
	DILocator1 = "DITestLocator1"
	DILocator2 = "DITestLocator2"

	AServiceName = "A"
	BServiceName = "B"

	BNamespaceName2 = "some/user/namespace"
	BServiceName2   = "BService"
	BRed            = "Red"
	BBlue           = "Blue"
	BGreen          = "Green"

	CServiceName = "C"
)

func TestInitializerSuccess(t *testing.T) {
	locator, err := CreateAndBind(DILocator1, func(binder Binder) error {
		binder.Bind(AServiceName, ASimpleService{})
		binder.Bind(BServiceName, BSimpleService{})

		return nil
	})
	if err != nil {
		t.Error("", err)
		return
	}

	aRaw, err := locator.GetDService(AServiceName)
	if err != nil {
		t.Error("", err)
		return
	}

	a, ok := aRaw.(*ASimpleService)
	if !ok {
		assert.True(t, ok, "invalid type")
		return
	}

	assert.True(t, a.initialized, "initializer not called")
}

func TestComplexInjectionName(t *testing.T) {
	locator, err := CreateAndBind(DILocator2, func(binder Binder) error {
		binder.Bind(CServiceName, CSimpleService{})
		binder.Bind(BServiceName2, BSimpleService{}).InNamespace(BNamespaceName2).
			QualifiedBy(BBlue).QualifiedBy(BRed).QualifiedBy(BGreen)

		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator2) {
		return
	}

	cRaw, err := locator.GetDService(CServiceName)
	if !assert.Nil(t, err, "couldn't create CService") {
		fmt.Println("", err)
		return
	}

	c, ok := cRaw.(*CSimpleService)
	if !assert.True(t, ok, "Invalid type for CService") {
		return
	}

	assert.True(t, c.initialized, "initializer not called")
}

type BSimpleService struct {
	initialized bool
}

func (b *BSimpleService) DargoInitialize() error {
	b.initialized = true
	return nil
}

type ASimpleService struct {
	B           *BSimpleService `inject:"B"`
	initialized bool
}

func (a *ASimpleService) DargoInitialize() error {
	if !a.B.initialized {
		return fmt.Errorf("Injected service B MUST have been initialized before this is called")
	}

	a.initialized = true

	return nil
}

type CSimpleService struct {
	B           *BSimpleService `inject:"some/user/namespace#BService@Red@Green"`
	initialized bool
}

func (c *CSimpleService) DargoInitialize() error {
	if !c.B.initialized {
		return fmt.Errorf("Injected service B must have been initialized in CSimpleService")
	}

	c.initialized = true

	return nil
}
