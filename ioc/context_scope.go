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

type contextScopeData struct {
	// contextCaches is a cache from int32->cache.Cache
	// The cache.Cache returned is a cache from idKey->interface{}
	contextCaches cache.Cache
	locator       ServiceLocator
}

func newContextScope(locator ServiceLocator) (ContextualScope, error) {
	retVal := &contextScopeData{
		locator: locator,
	}

	cache, err := cache.NewComputeFunctionCache(func(key interface{}) (interface{}, error) {
		return cache.NewCache(retVal, func(cycler interface{}) error {
			return fmt.Errorf("A cycle was detected in ContextScope services involving %v", cycler)
		})
	})
	if err != nil {
		return nil, err
	}

	retVal.contextCaches = cache

	return retVal, nil
}

func (cs *contextScopeData) addContext(dc *dargoContext) error {
	if cs.contextCaches.HasKey(dc.ID) {
		return fmt.Errorf("there is already a context for dc %v", dc)
	}

	cs.contextCaches.Compute(dc.ID)

	return nil
}

func (cs *contextScopeData) removeContext(dc *dargoContext) {
	cs.contextCaches.Remove(func(key interface{}, value interface{}) bool {
		if key != dc.ID {
			return false
		}

		// Clear the inner cache
		innerCache, ok := value.(cache.Cache)
		if !ok {
			// TODO: weird
			return true
		}

		// Destroy all services in this cache
		innerCache.Remove(func(key interface{}, value interface{}) bool {
			idKey, ok := key.(idKey)
			if !ok {
				return true
			}

			destroyer := idKey.desc.GetDestroyFunction()
			if destroyer == nil {
				return true
			}

			destroyer(cs.locator, idKey.desc, value)

			return true
		})

		return true
	})
}

func (cs *contextScopeData) GetScope() string {
	return ContextScope
}

func (cs *contextScopeData) getCache() (cache.Cache, error) {
	tl, err := threadManager.GetThreadLocal(dargoContextThreadLocal)
	if err != nil {
		return nil, err
	}

	raw, err := tl.Get()
	if err != nil {
		return nil, err
	}

	stack, ok := raw.(stack)
	if !ok {
		return nil, fmt.Errorf("unknown type from thread local")
	}

	rawContext, found := stack.Peek()
	if !found {
		return nil, fmt.Errorf("must be called from inside a context")
	}

	context, ok := rawContext.(*dargoContext)
	if !ok {
		return nil, fmt.Errorf("unknown type from peek")
	}

	raw2, err := cs.contextCaches.Compute(context.ID)
	if err != nil {
		return nil, err
	}

	cache, ok := raw2.(cache.Cache)
	if !ok {
		return nil, fmt.Errorf("unknown type of inner cache")
	}

	return cache, nil
}

func (cs *contextScopeData) FindOrCreate(locator ServiceLocator, desc Descriptor) (interface{}, error) {
	cache, err := cs.getCache()
	if err != nil {
		return nil, err
	}

	idKey := idKey{
		desc: desc,
	}

	return cache.Compute(idKey)
}

func (cs *contextScopeData) ContainsKey(locator ServiceLocator, desc Descriptor) bool {
	cache, err := cs.getCache()
	if err != nil {
		return false
	}

	idKey := idKey{
		desc: desc,
	}

	return cache.HasKey(idKey)
}

func (cs *contextScopeData) DestroyOne(locator ServiceLocator, desc Descriptor) error {
	cache, err := cs.getCache()
	if err != nil {
		return err
	}

	idKey := idKey{
		desc: desc,
	}

	cache.Remove(func(key interface{}, value interface{}) bool {
		if idKey == key {
			destroyer := desc.GetDestroyFunction()

			if destroyer != nil {
				destroyer(locator, desc, value)
			}

			return true
		}

		return false
	})

	return nil
}

func (cs *contextScopeData) GetSupportsNilCreation(locator ServiceLocator) bool {
	return false
}

func (cs *contextScopeData) IsActive(locator ServiceLocator) bool {
	return true
}

func (cs *contextScopeData) Shutdown(locator ServiceLocator) {
	cs.contextCaches.Remove(func(key interface{}, value interface{}) bool {
		innerCache, ok := value.(cache.Cache)
		if !ok {
			return true
		}

		innerCache.Remove(func(key interface{}, v2 interface{}) bool {
			idKey, ok := key.(idKey)
			if !ok {
				return true
			}

			destroyer := idKey.desc.GetDestroyFunction()
			if destroyer == nil {
				return true
			}

			destroyer(locator, idKey.desc, v2)

			return true
		})

		return true
	})
}

func (cs *contextScopeData) Compute(in interface{}) (interface{}, error) {
	key, ok := in.(idKey)
	if !ok {
		return nil, fmt.Errorf("incomding key not the expected type %v", in)
	}

	return cs.locator.CreateServiceFromDescriptor(key.desc)
}
