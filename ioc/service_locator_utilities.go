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

import "fmt"

// CreateAndBind creates a ServiceLocator with the given name (failing
// if the locator already exists) and binding the methods described in
// the BinderMethod into the locator
func CreateAndBind(locatorName string, method BinderMethod) (ServiceLocator, error) {
	locator, err := NewServiceLocator(locatorName, FailIfPresent)
	if err != nil {
		return nil, err
	}

	err = BindIntoLocator(locator, method)
	if err != nil {
		return nil, err
	}

	return locator, nil
}

// BindIntoLocator uses the binder to add services into an existing ServiceLocator
func BindIntoLocator(locator ServiceLocator, method BinderMethod) error {
	dcs, err := getDCS(locator)
	if err != nil {
		return err
	}

	binder := newBinder()

	err = method(binder)
	if err != nil {
		return err
	}

	descs := binder.finish()

	config, err := dcs.CreateDynamicConfiguration()
	if err != nil {
		return err
	}

	for _, desc := range descs {
		_, err = config.Bind(desc)
		if err != nil {
			return err
		}
	}

	err = config.Commit()
	if err != nil {
		return err
	}

	return nil
}

func getDCS(locator ServiceLocator) (DynamicConfigurationService, error) {
	dcsRaw, err := locator.GetService(SSK(DynamicConfigurationServiceName))
	if err != nil {
		return nil, err
	}

	dcs, ok := dcsRaw.(DynamicConfigurationService)
	if !ok {
		return nil, fmt.Errorf("DynamicConfigurationService is an unexpected type")
	}

	return dcs, nil
}

// EnableContextScope enables the use of the DargoContext
func EnableContextScope(locator ServiceLocator) error {
	dargoKey := CSK(ContextScope)
	filter := NewServiceKeyFilter(dargoKey)

	// TODO: Need idempotent semantics
	_, err := locator.GetBestDescriptor(filter)
	if err != nil {
		if isServiceNotFound(err) {
			return nil
		}

		return err
	}

	return BindIntoLocator(locator, func(binder Binder) error {
		binder.Bind(ContextScope, contextCreator).InNamespace(ContextualScopeNamespace)

		return nil
	})
}

func contextCreator(locator ServiceLocator, key ServiceKey) (interface{}, error) {
	return newContextScope(locator)
}

func DumpAllDescriptors(locator ServiceLocator) {
	if locator == nil {
		return
	}

	all, err := locator.GetDescriptors(AllFilter)
	if err != nil {
		fmt.Printf("Could not find any descriptors %s", err.Error())
		return
	}

	for index, desc := range all {
		fmt.Printf("%d. %v\n", (index + 1), desc)
	}
	fmt.Printf("finished dumping all %d descriptors from locator %s\n", len(all), locator.GetName())

}
