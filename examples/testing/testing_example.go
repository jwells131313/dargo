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

package testing

import (
	"github.com/jwells131313/dargo/ioc"
	"time"
)

var globalLocator ioc.ServiceLocator

// AnExpensiveService is a service that is expensive in some way
type AnExpensiveService interface {
	// DoExpensiveThing does some expensive operation
	DoExpensiveThing(string) (string, error)
}

// NormalExpensiveServiceData is an implementation of AnExpensiveService in the normal code
type NormalExpensiveServiceData struct {
}

// SomeOtherServiceData is another user service which injects the expensive service
type SomeOtherServiceData struct {
	// ExpensiveService is the expensive service, injected by Dargo
	ExpensiveService AnExpensiveService `inject:"AnExpensiveService"`
}

// DoExpensiveThing implements the interface with a long sleep and returns "Normal"
func (nesd *NormalExpensiveServiceData) DoExpensiveThing(thingToDo string) (string, error) {
	time.Sleep(5 * time.Second)

	return "Normal", nil
}

func init() {
	myLocator, err := ioc.CreateAndBind("TestingExampleLocator", func(binder ioc.Binder) error {
		binder.Bind("UserService", SomeOtherServiceData{})
		binder.Bind("AnExpensiveService", NormalExpensiveServiceData{})

		return nil
	})
	if err != nil {
		panic(err)
	}

	globalLocator = myLocator
}

// DoSomeUserCode is the user code that uses the injected service
func (other *SomeOtherServiceData) DoSomeUserCode() (string, error) {
	return other.ExpensiveService.DoExpensiveThing("foo")
}
