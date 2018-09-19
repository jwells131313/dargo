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
	"sync/atomic"
	"testing"
	"time"
)

const (
	ImmediateTestLocator1 = "ImmediateTestLocator1"
	ImmediateTestLocator2 = "ImmediateTestLocator2"
	ImmediateTestLocator3 = "ImmediateTestLocator3"
	ImmediateTestLocator4 = "ImmediateTestLocator4"

	ImmediateServiceName = "ImmediateService"
)

var globalStarted bool
var howManyStarted int32

func TestExistingAreStarted(t *testing.T) {
	globalStarted = false

	locator, err := CreateAndBind(ImmediateTestLocator1, func(binder Binder) error {
		binder.Bind(ImmediateServiceName, &ImmediateService{}).InScope(ImmediateScope)
		return nil
	})
	if !assert.Nil(t, err, "could not create locator") {
		return
	}

	if !assert.False(t, globalStarted, "should not have started yet") {
		return
	}

	err = EnableImmediateScope(locator)
	if !assert.Nil(t, err, "could not enable immediate scope %v", err) {
		return
	}

	for lcv := 0; lcv < 20; lcv++ {
		if globalStarted {
			break
		}

		time.Sleep(1 * time.Second)
	}

	assert.True(t, globalStarted, "service was not started")
}

func TestAddedAreStartedAndStopped(t *testing.T) {
	globalStarted = false

	locator, err := CreateAndBind(ImmediateTestLocator2, func(binder Binder) error {
		return nil
	})
	if !assert.Nil(t, err, "could not create locator") {
		return
	}

	if !assert.False(t, globalStarted, "should not have started yet") {
		return
	}

	err = EnableImmediateScope(locator)
	if !assert.Nil(t, err, "could not enable immediate scope %v", err) {
		return
	}

	if !assert.False(t, globalStarted, "should not have started yet") {
		return
	}

	err = BindIntoLocator(locator, func(binder Binder) error {
		binder.Bind(ImmediateServiceName, &ImmediateService{}).InScope(ImmediateScope).
			AndDestroyWith(func(l ServiceLocator, d Descriptor, o interface{}) error {
				immediate, ok := o.(*ImmediateService)
				if !ok {
					return fmt.Errorf("incorrect type")
				}

				immediate.stopped = true

				return nil
			})
		return nil
	})

	for lcv := 0; lcv < 20; lcv++ {
		if globalStarted {
			break
		}

		time.Sleep(1 * time.Second)
	}

	assert.True(t, globalStarted, "service was not started")

	raw, err := locator.GetDService(ImmediateServiceName)
	if !assert.Nil(t, err, "didn't find immediate service?") {
		return
	}

	is, ok := raw.(*ImmediateService)
	if !assert.True(t, ok, "invalid type") {
		return
	}

	if !assert.False(t, is.stopped, "already stopped?") {
		return
	}

	err = UnbindDServices(locator, ImmediateServiceName)
	if !assert.Nil(t, err, "could not unbind immediate service") {
		return
	}

	for lcv := 0; lcv < 20; lcv++ {
		if is.stopped {
			break
		}

		time.Sleep(1 * time.Second)
	}
}

func TestManyStarted(t *testing.T) {
	howManyStarted = 0

	locator, err := CreateAndBind(ImmediateTestLocator3, func(binder Binder) error {
		for lcv := 0; lcv < 100; lcv++ {
			q := fmt.Sprintf("Qualifier%d", lcv)
			binder.Bind(ImmediateServiceName, &CountingImmediateService{}).InScope(ImmediateScope).
				QualifiedBy(q)
		}
		return nil
	})
	if !assert.Nil(t, err, "could not create locator") {
		return
	}

	if !assert.Equal(t, int32(0), howManyStarted, "should not be any started yet") {
		return
	}

	// Let the floodgates open
	err = EnableImmediateScope(locator)
	if !assert.Nil(t, err, "could not start immediate scope") {
		return
	}

	for lcv := 0; lcv < 200; lcv++ {
		value := atomic.AddInt32(&howManyStarted, 0)
		if value == 100 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	assert.Equal(t, int32(100), howManyStarted, "did not start expectee number")

}

func TestManyStartedAfterEnabled(t *testing.T) {
	howManyStarted = 0

	locator, err := CreateAndBind(ImmediateTestLocator4, func(binder Binder) error {
		return nil
	})
	if !assert.Nil(t, err, "could not create locator") {
		return
	}

	// Let the floodgates open
	err = EnableImmediateScope(locator)
	if !assert.Nil(t, err, "could not start immediate scope") {
		return
	}

	if !assert.Equal(t, int32(0), howManyStarted, "should not be any started yet") {
		return
	}

	err = BindIntoLocator(locator, func(binder Binder) error {
		for lcv := 0; lcv < 100; lcv++ {
			q := fmt.Sprintf("Qualifier%d", lcv)
			binder.Bind(ImmediateServiceName, &CountingImmediateService{}).InScope(ImmediateScope).
				QualifiedBy(q)
		}
		return nil
	})
	if !assert.Nil(t, err, "could not bind into locator") {
		return
	}

	for lcv := 0; lcv < 200; lcv++ {
		value := atomic.AddInt32(&howManyStarted, 0)
		if value == 100 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	assert.Equal(t, int32(100), howManyStarted, "did not start expectee number")

}

type ImmediateService struct {
	stopped bool
}

func (is *ImmediateService) DargoInitialize(desc Descriptor) error {
	globalStarted = true
	return nil
}

type CountingImmediateService struct {
}

func (cis *CountingImmediateService) DargoInitialize(Descriptor) error {
	atomic.AddInt32(&howManyStarted, 1)
	return nil
}
