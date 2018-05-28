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

package api

import (
	"errors"
	"reflect"
)

// Binder A fluent interface for creating descriptors
type Binder interface {
	To(reflect.Type) Binder
	Named(string) Binder
	Build() (WriteableDescriptor, error)
}

type binder struct {
	creator   func(ServiceLocator) (interface{}, error)
	contracts []reflect.Type
	name      string
}

// Bind the descriptor to the interface type  toMe must be an interface
func Bind(creatorFunc func(ServiceLocator) (interface{}, error)) Binder {
	return &binder{
		creator: creatorFunc,
	}
}

func (binder *binder) To(t reflect.Type) Binder {
	binder.contracts = append(binder.contracts, t)
	return binder
}

func (binder *binder) Named(userName string) Binder {
	binder.name = userName
	return binder
}

func (binder *binder) Build() (WriteableDescriptor, error) {
	return nil, errors.New("not yet implemented")
}
