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
	"sync"
)

type contextScopeData struct {
	lock          sync.Mutex
	contextCaches map[int32]cache.Cache
	locator       ServiceLocator
}

func newContextScope(locator ServiceLocator) (ContextualScope, error) {
	return &contextScopeData{
		contextCaches: make(map[int32]cache.Cache),
		locator:       locator,
	}, nil
}

func (cs *contextScopeData) addContext(dc *dargoContext) error {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	_, found := cs.contextCaches[dc.ID]
	if found {
		return fmt.Errorf("there is already a context for dc %v", dc)
	}

	cache, err := cache.NewCache(cs, nil)
	if err != nil {
		return err
	}

	cs.contextCaches[dc.ID] = cache

	return nil
}

func (cs *contextScopeData) removeContext(dc *dargoContext) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cache, found := cs.contextCaches[dc.ID]
	if !found {
		return
	}

	// Remove cache
	delete(cs.contextCaches, dc.ID)

	// Destroy all services in this cache
	cache.Remove(func(key interface{}, value interface{}) bool {
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

	rawID, found := stack.Peek()
	if !found {
		return nil, fmt.Errorf("must be called from inside a context")
	}

	contextID, ok := rawID.(int32)
	if !ok {
		return nil, fmt.Errorf("unknown type from peek")
	}

	cache, found := cs.contextCaches[contextID]
	if !found {
		return nil, fmt.Errorf("the context was either closed or was never created")
	}

	return cache, nil
}

func (cs *contextScopeData) FindOrCreate(locator ServiceLocator, desc Descriptor) (interface{}, error) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

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
	panic("implement me")
}

func (cs *contextScopeData) DestroyOne(locator ServiceLocator, desc Descriptor) error {
	panic("implement me")
}

func (cs *contextScopeData) GetSupportsNilCreation(locator ServiceLocator) bool {
	return false
}

func (cs *contextScopeData) IsActive(locator ServiceLocator) bool {
	return true
}

func (cs *contextScopeData) Shutdown(locator ServiceLocator) {
	panic("implement me")
}

func (cs *contextScopeData) Compute(in interface{}) (interface{}, error) {
	key, ok := in.(idKey)
	if !ok {
		return nil, fmt.Errorf("incomding key not the expected type %v", in)
	}

	return cs.locator.CreateServiceFromDescriptor(key.desc)
}
