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
	"github.com/jwells131313/goethe/cache"
)

type idKey struct {
	desc Descriptor
}

type singletonContextualData struct {
	locator ServiceLocator
	cache   cache.Computable
}

func newSingletonScope(locator *serviceLocatorData) (ContextualScope, error) {
	retVal := &singletonContextualData{
		locator: locator,
	}

	c, err := cache.NewCache(retVal, func(in interface{}) error {
		return fmt.Errorf("cycle detected in singleton scope involving %v", in)
	})
	if err != nil {
		return nil, err
	}

	retVal.cache = c

	return retVal, nil
}

func (single *singletonContextualData) GetScope() string {
	return Singleton
}

func (single *singletonContextualData) FindOrCreate(locator ServiceLocator, desc Descriptor) (interface{}, error) {
	return single.cache.Compute(idKey{desc: desc})
}

func (single *singletonContextualData) ContainsKey(locator ServiceLocator, desc Descriptor) bool {
	panic("implement me")
}

func (single *singletonContextualData) DestroyOne(locator ServiceLocator, desc Descriptor) error {
	panic("implement me")
}

func (single *singletonContextualData) GetSupportsNilCreation(locator ServiceLocator) bool {
	return false
}

func (single *singletonContextualData) IsActive(locator ServiceLocator) bool {
	return true
}

func (single *singletonContextualData) Shutdown(locator ServiceLocator) {
	panic("implement me")
}

func (single *singletonContextualData) Compute(in interface{}) (interface{}, error) {
	key, ok := in.(idKey)
	if !ok {
		return nil, fmt.Errorf("incomding key not the expected type %v", in)
	}

	return single.locator.CreateServiceFromDescriptor(key.desc)
}
