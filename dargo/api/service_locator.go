package api
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

import "reflect"

// ServiceLocator The main registry for dargo.  Use it to get context sensitive lookups
// for application services
type ServiceLocator interface {
	// GetService gets the service that is correct for the current context with the given interface
	GetService(toMe reflect.Type) (interface{}, error)
	
	// GetDescriptors Returns all descriptors that return true when passed through the input function
	// will not return nil, but may return an empty list
	GetDescriptors(func (Descriptor) bool) []Descriptor
	
	// GetBestDescriptor returns the best descriptor found returning true through the input function
	// The best descriptor is the one with the highest rank, or if rank is equal the one with the
	// lowest serviceId or if the serviceId are the same the one with the highest locatorId
	GetBestDescriptor(func (Descriptor) bool) (Descriptor, bool)
	
	// GetDescriptorsWithName Returns all descriptors that return true when passed through the input function
	// and which have the given name.  Can drastically reduce the number of descriptors passed to the method
	// will not return nil, but may return an empty list
	GetDescriptorsWithNameOrType(func (Descriptor) bool, reflect.Type, string) []Descriptor
	
	// GetBestDescriptor returns the best descriptor found returning true through the input function
	// and which have the given name
	// The best descriptor is the one with the highest rank, or if rank is equal the one with the
	// lowest serviceId or if the serviceId are the same the one with the highest locatorId
	GetBestDescriptorWithNameOrType(func (Descriptor) bool, reflect.Type, string) (Descriptor, bool)
	
	// GetName gets the name of this ServiceLocator
	GetName() string
	
	// GetID Gets the id of this ServiceLocator
	GetID() int64
	
	// Will shut down all services associated with this ServiceLocator
	Shutdown()
}
