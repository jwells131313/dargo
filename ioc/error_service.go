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
	"reflect"
)

// ErrorInformation is passed to the ErrorService to provide information about
// the error that has occurred
type ErrorInformation interface {
	// GetType returns the type of error condition that has occurred
	// The valid values are:<UL>
	// <LI>DYNAMIC_CONFIGURATION_FAILURE</LI>
	// <LI>SERVICE_CREATION_FAILURE</LI>
	// </UL>
	GetType() string
	// GetDescriptor returns the Descriptor associated with the failure
	GetDescriptor() Descriptor
	// GetInjectee returns the injectee for which the SERVICE_CREATION_FAILURE
	// occurred, if it is known, and nil otherwise
	GetInjectee() reflect.Type
	// GetAssociatedError returns the underlying error that occurred to cause
	// the failure, or nil if the underlying error is not known
	GetAssociatedError() error
}

// ErrorService when users implement and bind this service into the
// namespace user/services with name ErrorService.  The OnFailure method
// will be invoked when certain failures happen when performing operations
// in dargo
//
// An implementation of ErrorService must be in the Singleton scope.
// Implementations of ErrorService will be instantiated as soon as
// they are added to Dargo in order to avoid deadlocks and circular references
type ErrorService interface {
	// OnFailure is invoked by the system when certain failures happen
	// during processing.  Any panic or error from this method is
	// currently ignored
	OnFailure(ErrorInformation) error
}

type errorInformationData struct {
	typ      string
	desc     Descriptor
	injectee reflect.Type
	err      error
}

func newErrorImformation(typ string, desc Descriptor, injectee reflect.Type, err error) ErrorInformation {
	return &errorInformationData{
		typ:      typ,
		desc:     desc,
		injectee: injectee,
		err:      err,
	}
}

func (eid *errorInformationData) GetType() string {
	return eid.typ
}

func (eid *errorInformationData) GetDescriptor() Descriptor {
	return eid.desc
}

func (eid *errorInformationData) GetInjectee() reflect.Type {
	return eid.injectee
}

func (eid *errorInformationData) GetAssociatedError() error {
	return eid.err
}

func (eid *errorInformationData) String() string {
	errS := ""
	if eid.err != nil {
		errS = eid.err.Error()
	}

	return fmt.Sprintf("ErrorInformation(%s,%v,%s)", eid.typ, eid.desc, errS)
}
