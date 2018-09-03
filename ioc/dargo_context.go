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
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sync/atomic"
	"time"
)

// DargoContextCreationService a service in the DargoContext scope that
// has a method for getting the DargoContext that was used to create
// the service itself.  It is useful to inject into other services
// in the DargoContext that need to know the originating context.Context
type DargoContextCreationService interface {
	// GetDargoCreationContext gets the context.Context that was used to originate
	// this service, which is useful for other services in the DargoContext
	// scope that need to know the originating context.Context
	GetDargoCreationContext() context.Context
}

type dargoContext struct {
	ID           int32
	parent       context.Context
	locator      ServiceLocator
	dargoContext *contextScopeData
	doneChannel  chan struct{}
}

var (
	nextContextID int32
)

// NewDargoContext creates a new dargo context with the given parent context
func NewDargoContext(parent context.Context, locator ServiceLocator) (context.Context, error) {
	id := atomic.AddInt32(&nextContextID, 1)

	retVal := &dargoContext{
		ID:          id,
		parent:      parent,
		locator:     locator,
		doneChannel: make(chan struct{}),
	}

	contextImplRaw, err := locator.GetService(CSK(ContextScope))
	if err != nil {
		if isServiceNotFound(err) {
			return nil, fmt.Errorf("there is no ContextScope.  You need to call EnableContextScope in the ioc package to install a ContextScope handler")
		}

		return nil, errors.Wrap(err, "error getting ContextScope while creating new DargoContext")
	}

	contextImpl, ok := contextImplRaw.(*contextScopeData)
	if !ok {
		return nil, fmt.Errorf("The context implementation was not the expected type while creating new DargoContext")
	}

	retVal.dargoContext = contextImpl

	contextImpl.addContext(retVal)

	threadManager.Go(retVal.killMe)

	return retVal, nil
}

type doneStruct struct{}

func (dgo *dargoContext) killMe() {
	// Wait for parent to be done
	<-dgo.parent.Done()

	dgo.dargoContext.removeContext(dgo)

	dgo.doneChannel <- doneStruct{}
}

func (dgo *dargoContext) Deadline() (deadline time.Time, ok bool) {
	return dgo.parent.Deadline()
}

func (dgo *dargoContext) Done() <-chan struct{} {
	return dgo.doneChannel
}

func (dgo *dargoContext) Err() error {
	return dgo.parent.Err()
}

func (dgo *dargoContext) Value(key interface{}) interface{} {

	switch key.(type) {
	case ServiceKey:
		retVal, err := dgo.getValue(key.(ServiceKey))
		if err != nil {
			return dgo.parent.Value(key)
		}

		return retVal
	case string:
		dsk := DSK(key.(string))
		retVal, err := dgo.getValue(dsk)
		if retVal == nil || err != nil {
			return dgo.parent.Value(key)
		}

		return retVal
	default:
		return dgo.parent.Value(key)
	}

}

type valReply struct {
	val interface{}
	err error
}

func (dgo *dargoContext) getValue(key ServiceKey) (interface{}, error) {
	tid := threadManager.GetThreadID()

	if tid < 0 {
		c := make(chan (*valReply))

		threadManager.Go(dgo.getChannelDargoValue, key, c)

		rply := <-c

		return rply.val, rply.err
	}

	return dgo.getGoetheDargoValue(key)
}

func (dgo *dargoContext) getChannelDargoValue(key ServiceKey, ch chan *valReply) {
	retVal, err := dgo.getGoetheDargoValue(key)

	ret := &valReply{
		val: retVal,
		err: err,
	}

	ch <- ret
}

func (dgo *dargoContext) getGoetheDargoValue(key ServiceKey) (interface{}, error) {
	tl, err := threadManager.GetThreadLocal(dargoContextThreadLocal)
	if err != nil {
		return nil, err
	}

	rawStack, err := tl.Get()
	if err != nil {
		return nil, err
	}

	stack, ok := rawStack.(stack)
	if !ok {
		return nil, fmt.Errorf("unexpected type from thread local")
	}

	err = stack.Push(dgo)
	if err != nil {
		return nil, err
	}
	defer stack.Pop()

	return dgo.locator.GetService(key)
}

type dargoContextCreationServiceData struct {
	context context.Context
}

func (dccsd *dargoContextCreationServiceData) GetDargoCreationContext() context.Context {
	return dccsd.context
}

func (dccsd *dargoContextCreationServiceData) DargoInitialize() error {
	tid := threadManager.GetThreadID()
	if tid < 0 {
		return fmt.Errorf("DargoCreationContextService not initialized on goethe thread")
	}

	tl, err := threadManager.GetThreadLocal(dargoContextThreadLocal)
	if err != nil {
		return err
	}

	rawStack, err := tl.Get()
	if err != nil {
		return err
	}

	stack, ok := rawStack.(stack)
	if !ok {
		return fmt.Errorf("unknown type of thread local when creating DargoCreationContextService")
	}

	rawContext, found := stack.Peek()
	if !found {
		return fmt.Errorf("nothing pushed onto stack improperly when creating DargoCreationContextService")
	}

	dccsd.context, ok = rawContext.(*dargoContext)
	if !ok {
		return fmt.Errorf("unknown type of thread local value when creating DargoCreationContextService")
	}

	return nil
}
