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

// ValidationInformation is passed into the Validator.Validate method
// to provide information about the service being validated
type ValidationInformation interface {
	// GetOperation returns the operation to be performed
	// Valid values are:
	// BIND - The candidate descriptor is being added to the system
	// UNBIND - The candidate descriptor is being removed from the system
	// LOOKUP - The candidate descriptor is being looked up
	GetOperation() string

	// GetCandidate returns the descriptor that is being looked up or
	// bound or unbound in this operation
	GetCandidate() Descriptor

	// GetInjecteeDescriptor returns the descriptor of the service which
	// will be injected if known.  Returns nil if there is no Injectee (for
	// example on a direct lookup) or if this is a bind/unbind operation
	GetInjecteeDescriptor() Descriptor

	// GetFilter returns the Filter being used to lookup the service
	// or nil if this is a struct injection or a bind/unbind operation
	GetFilter() Filter
}

// Validator is returned by the ValidationService for Filters that
// match.  The Validate method is called at specific points in the
// flow of control of Dargo
type Validator interface {
	// Validate is called whenever it has been determined that a validating
	// service is to be injected into an injection point, or when a descriptor
	// is being looked up explicitly with the API, or a descriptor is being
	// bound or unbound into the registry.
	//
	// The operation will determine what operation is being performed.  In the
	// BIND or UNBIND cases the Injectee will be nil.  In the LOOKUP case
	// the Injectee will be non-nil if this is being done as part of an
	// injection point.  In the LOOKUP case the Injectee will be nil if this
	// is being looked up directly
	Validate(ValidationInformation) error
}

// ValidationService can be used to add validation points in the flow
// of control in Dargo
//
// An implementation of ValidationService must be in the Singleton scope.
// Implementations of ValidationService will be instantiated as soon as
// they are added to Dargo in order to avoid deadlocks and circular references
type ValidationService interface {
	// GetFilter will be run at least once per descriptor at the point that the descriptor
	// is being looked up, either with the lookup API or due to
	// an injection resolution.  The decision made by this filter will be cached and
	// used every time that Descriptor is subsequently looked up.  No validation checks
	// should be done in the returned filter, it is purely meant to limit the
	// Descriptors that are passed into the validator.
	// The filter may be run more than once on a descriptor if some condition caused
	// the cache of results per descriptor to become invalidated
	GetFilter() Filter

	// GetValidator returns the Validator that will be run whenever
	// a Descriptor that passed the filter is to be looked up with the API
	// or injected into an injection point, or on any bind or unbind operation.
	// If this method returns a non-nil error then the operation will not proceed.
	// By default this error is NOT thrown up the stack, but will be passed to the
	// ErrorService
	GetValidator() Validator
}

type validationInformationData struct {
	operation  string
	descriptor Descriptor
	injectee   Descriptor
	filter     Filter
}

func newValidationInformation(operation string,
	desc Descriptor, injectee Descriptor, filter Filter) ValidationInformation {
	return &validationInformationData{
		operation:  operation,
		descriptor: desc,
		injectee:   injectee,
		filter:     filter,
	}
}

func (vid *validationInformationData) GetOperation() string {
	return vid.operation
}

func (vid *validationInformationData) GetCandidate() Descriptor {
	return vid.descriptor
}

func (vid *validationInformationData) GetInjecteeDescriptor() Descriptor {
	return vid.injectee
}

func (vid *validationInformationData) GetFilter() Filter {
	return vid.filter
}

func (vid *validationInformationData) String() string {
	retVal := fmt.Sprintf("ValidationInformation(%s,%v,%v,%v)", vid.operation,
		vid.descriptor, vid.injectee, vid.filter)
	return retVal
}
