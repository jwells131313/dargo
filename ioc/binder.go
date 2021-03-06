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

import "reflect"

// BinderMethod is the method signature for binding services into the ServiceLocator
type BinderMethod func(Binder) error

// Binder A fluent interface for creating descriptors
type Binder interface {
	// Bind binds the given name to the structure.  prototype can be the struct to be created
	// or can be a pointer to the struct to be created.  prototype is NOT the structure
	// that will be used as a service by Dargo.  If prototype implements DargoInitializer
	// then the DargoInitialize method will be called on it prior to being given to other
	// services
	Bind(name string, prototype interface{}) Binder
	// BindWithCreator binds the given name to a creation function
	BindWithCreator(name string, bindMethod func(ServiceLocator, Descriptor) (interface{}, error)) Binder
	// BindConstant binds the exact constant as-is into the ServiceLocator
	BindConstant(name string, constant interface{}) Binder
	// InScope changes the scope to the given scope.  The default scope is Singleton
	InScope(string) Binder
	// InNamespace changes the namespace to the given value.  The default namespace is default
	InNamespace(string) Binder
	// QualifiedBy adds the given qualifier name
	QualifiedBy(string) Binder
	// Ranked changes the rank to the given rank.  Higher ranks are preferred over lower ranks
	Ranked(int32) Binder
	// AndDestroyWith sets the destroyer function to the given function
	AndDestroyWith(func(ServiceLocator, Descriptor, interface{}) error) Binder
}

// DargoInitializer is used when using Binder.Bind and need
// to be able to do further initialization after the services have
// been injected into the structure and before it is given to the injectee
// or lookup user
type DargoInitializer interface {
	// DargoInitialize is a method that will be called after all the
	// injected fields have been filled in.  If this method returns
	// a non-nil error then the creation of the service will fail
	// The descriptor passed in is the descriptor being used to create
	// the service
	DargoInitialize(Descriptor) error
}

type binder struct {
	parent      *serviceLocatorData
	descriptors []Descriptor
	current     WriteableDescriptor
	qualifiers  []string
}

func newBinder(parent *serviceLocatorData) *binder {
	return &binder{
		parent:      parent,
		descriptors: make([]Descriptor, 0),
	}
}

func (binder *binder) BindWithCreator(name string, cf func(ServiceLocator, Descriptor) (interface{}, error)) Binder {
	if binder.current != nil {
		if len(binder.qualifiers) > 0 {
			binder.current.SetQualifiers(binder.qualifiers)
		}

		binder.descriptors = append(binder.descriptors, binder.current)

		binder.current = nil
		binder.qualifiers = nil
	}

	binder.current = NewWriteableDescriptor()
	binder.current.SetCreateFunction(cf)
	binder.current.SetName(name)

	binder.qualifiers = make([]string, 0)

	return binder
}

func (binder *binder) BindConstant(name string, constant interface{}) Binder {
	return binder.BindWithCreator(name, func(ServiceLocator, Descriptor) (interface{}, error) {
		return constant, nil
	})
}

func (binder *binder) Bind(name string, str interface{}) Binder {
	if binder.current != nil {
		if len(binder.qualifiers) > 0 {
			binder.current.SetQualifiers(binder.qualifiers)
		}

		binder.descriptors = append(binder.descriptors, binder.current)

		binder.current = nil
		binder.qualifiers = nil
	}

	ty := reflect.TypeOf(str)
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}
	if ty.Kind() != reflect.Struct {
		panic("Bind must be passed a struct")
	}

	cf := newCreatorFunc(ty, binder.parent)

	binder.current = NewWriteableDescriptor()
	binder.current.SetCreateFunction(cf)
	binder.current.SetName(name)

	binder.qualifiers = make([]string, 0)

	return binder
}

func (binder *binder) InScope(scope string) Binder {
	if binder.current == nil {
		panic("must call bind before this method")
	}

	binder.current.SetScope(scope)

	return binder
}

func (binder *binder) InNamespace(namespace string) Binder {
	if binder.current == nil {
		panic("must call bind before this method")
	}

	binder.current.SetNamespace(namespace)

	return binder
}

func (binder *binder) QualifiedBy(qualifier string) Binder {
	if binder.current == nil {
		panic("must call bind before this method")
	}

	binder.qualifiers = append(binder.qualifiers, qualifier)

	return binder
}

func (binder *binder) Ranked(rank int32) Binder {
	if binder.current == nil {
		panic("must call bind before this method")
	}

	binder.current.SetRank(rank)

	return binder

}

func (binder *binder) AndDestroyWith(destroyer func(ServiceLocator, Descriptor, interface{}) error) Binder {
	if binder.current == nil {
		panic("must call bind before this method")
	}

	binder.current.SetDestroyFunction(destroyer)

	return binder
}

func (binder *binder) finish() []Descriptor {
	if binder.current != nil {
		if len(binder.qualifiers) > 0 {
			binder.current.SetQualifiers(binder.qualifiers)
		}

		binder.descriptors = append(binder.descriptors, binder.current)
	}

	return binder.descriptors
}
