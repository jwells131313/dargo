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

package example

import (
	"context"
	"fmt"
	"github.com/jwells131313/dargo/ioc"
	"time"
)

const (
	userNameKey = "contextUser"
)

// AuthorizationService is an example authorization service
type AuthorizationService interface {
	// MotherMayI asks the auth service for permission to do something
	MotherMayI() bool
}

// RequestContext is an example request-scoped context
type RequestContext struct {
	parent context.Context
	user   string
}

// NewRequestContext creates the example request-scoped context
func NewRequestContext(parent context.Context, user string) context.Context {
	return &RequestContext{
		parent: parent,
		user:   user,
	}
}

func (rc *RequestContext) Deadline() (deadline time.Time, ok bool) {
	return rc.parent.Deadline()
}

func (rc *RequestContext) Done() <-chan struct{} {
	return rc.parent.Done()
}

func (rc *RequestContext) Err() error {
	return rc.parent.Err()
}

func (rc *RequestContext) Value(key interface{}) interface{} {
	switch key.(type) {
	case string:
		skey := key.(string)
		if skey == userNameKey {
			return rc.user
		}

		return rc.parent.Value(key)
	default:
		return rc.parent.Value(key)
	}
}

// AuthorizationServiceData is the struct implementing AuthorizationService
// It injects the DargoContextCreationService to get the context under
// which this service was created
type AuthorizationServiceData struct {
	ContextService ioc.DargoContextCreationService `inject:"DargoCreationContextService"`
}

func (asd *AuthorizationServiceData) MotherMayI() bool {
	context := asd.ContextService.GetDargoCreationContext()

	userRaw := context.Value(userNameKey)
	if userRaw == nil {
		return false
	}

	user := userRaw.(string)

	if user == "Mallory" {
		return false
	}

	return true
}

func runContextExample() error {
	locator, _ := ioc.CreateAndBind("ContextExample", func(binder ioc.Binder) error {
		binder.Bind("AuthService", AuthorizationServiceData{}).InScope(ioc.ContextScope)

		return nil
	})

	ioc.EnableContextScope(locator)

	aliceContext, aliceCanceller, _ := createContext(locator, "Alice")
	defer aliceCanceller()

	bobContext, bobCanceller, _ := createContext(locator, "Bob")
	defer bobCanceller()

	malloryContext, malloryCanceller, _ := createContext(locator, "Mallory")
	defer malloryCanceller()

	aliceAuthorizer := getAuthorizeService(aliceContext)
	bobAuthorizer := getAuthorizeService(bobContext)
	malloryAuthorizer := getAuthorizeService(malloryContext)

	canI := aliceAuthorizer.MotherMayI()
	if !canI {
		return fmt.Errorf("Alice should have been able to go")
	}

	canI = bobAuthorizer.MotherMayI()
	if !canI {
		return fmt.Errorf("Alice should have been able to go")
	}

	canI = malloryAuthorizer.MotherMayI()
	if canI {
		// Mallory should have NOT been allowed
		return fmt.Errorf("Mallory is a bad person, and should not be allowed to do anything")
	}

	return nil
}

func createContext(locator ioc.ServiceLocator, user string) (context.Context, context.CancelFunc, error) {
	parent, canceller := context.WithCancel(context.Background())

	requestScoped := NewRequestContext(parent, user)

	dargoContext, err := ioc.NewDargoContext(requestScoped, locator)

	return dargoContext, canceller, err
}

func getAuthorizeService(context context.Context) AuthorizationService {
	raw := context.Value("AuthService")
	if raw == nil {
		return nil
	}

	return raw.(AuthorizationService)
}
