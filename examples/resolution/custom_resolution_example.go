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

package resolution

import (
	"fmt"
	"github.com/jwells131313/dargo/ioc"
	"reflect"
)

// AutomaticResolver will resolve any field of a struct that has
// a type with a service with a name that equals that type
type AutomaticResolver struct {
}

func (ar *AutomaticResolver) Resolve(locator ioc.ServiceLocator, injectee ioc.Injectee) (*reflect.Value, bool, error) {
	field := injectee.GetField()
	typ := field.Type

	var name string
	switch typ.Kind() {
	case reflect.Ptr:
		itype := typ.Elem()

		name = itype.Name()
		break
	case reflect.Interface:
		name = typ.Name()
		break
	default:
		return nil, false, nil
	}

	if name == "" {
		return nil, false, nil
	}

	svc, err := locator.GetDService(name)
	if err != nil {
		return nil, false, nil
	}

	rVal := reflect.ValueOf(svc)

	return &rVal, true, nil
}

// BService is a service injected into AService that just prints Hello, World
type BService struct {
}

func (b *BService) run() {
	fmt.Println("Hello, World")
}

// AService is injected using the custom injector
type AService struct {
	// BService will be magically injected, even without an indicator on the struct
	BService *BService
}

// CustomResolution is a method that will create a locator, binding in the
// custom resolver and the A and B Services.  It will then get the AService
// and use the injected BService.  This example shows how a custom resolver
// can use whatever resources it has available to choose injection points
// in a service
func CustomResolution() error {
	locator, err := ioc.CreateAndBind("AutomaticResolverLocator", func(binder ioc.Binder) error {
		binder.Bind(ioc.InjectionResolverName, &AutomaticResolver{}).InNamespace(ioc.UserServicesNamespace)
		binder.Bind("AService", &AService{})
		binder.Bind("BService", &BService{})

		return nil
	})
	if err != nil {
		return err
	}

	aServiceRaw, err := locator.GetDService("AService")
	if err != nil {
		return err
	}

	aService := aServiceRaw.(*AService)

	aService.BService.run()

	return nil
}
