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

	DoNotBindService   = "DoNotBindService"
	NeverUnbindService = "NeverUnbindService"
	SimpleServiceName  = "SimpleService"
)

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
	default:
	}

	return nil
}

type SimpleService struct {
}

type InvalidService struct {
}
