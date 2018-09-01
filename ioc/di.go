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
	"reflect"
)

type diData struct {
	ty reflect.Type
}

func newCreatorFunc(ty reflect.Type) func(ServiceLocator, Descriptor) (interface{}, error) {
	diVal := &diData{
		ty: ty,
	}

	retVal := func(locator ServiceLocator, desc Descriptor) (interface{}, error) {
		return diVal.create(locator, desc)
	}

	return retVal
}

type indexAndValueOfDependency struct {
	index int
	value reflect.Value
}

func (di *diData) create(locator ServiceLocator, desc Descriptor) (interface{}, error) {
	numFields := di.ty.NumField()

	dependencies := make([]*indexAndValueOfDependency, 0)
	for lcv := 0; lcv < numFields; lcv++ {
		fieldVal := di.ty.Field(lcv)

		injectString := fieldVal.Tag.Get("inject")

		if injectString != "" {
			// TODO: Needs to be a whole, like, parsing discussion...
			dependency, err := locator.GetDService(injectString)
			if err != nil {
				return nil, err
			}

			dependencyAsValue := reflect.ValueOf(dependency)

			dependencies = append(dependencies, &indexAndValueOfDependency{
				index: lcv,
				value: dependencyAsValue,
			})
		}
	}

	retVal := reflect.New(di.ty)
	indirect := reflect.Indirect(retVal)

	for _, iav := range dependencies {
		index := iav.index
		value := iav.value

		fieldValue := indirect.Field(index)
		fieldValue.Set(value)
	}

	iFace := retVal.Interface()

	initializer, ok := iFace.(DargoInitializer)
	if ok {
		err := initializer.DargoInitialize()
		if err != nil {
			return nil, err
		}
	}

	return iFace, nil
}
