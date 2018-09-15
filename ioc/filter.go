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

// Filter is used to filter descriptors for matching services
type Filter interface {
	Filter(Descriptor) bool
}

type serviceKeyFilter struct {
	keys []ServiceKey
}

// NewServiceKeyFilter creates a filter for the given service key
func NewServiceKeyFilter(keys ...ServiceKey) Filter {
	cpy := make([]ServiceKey, len(keys))
	copy(cpy, keys)

	return &serviceKeyFilter{
		cpy,
	}
}

func (filter *serviceKeyFilter) Filter(desc Descriptor) bool {
	for _, key := range filter.keys {
		if !filterOne(key, desc) {
			return false
		}
	}

	return true
}

func filterOne(key ServiceKey, desc Descriptor) bool {
	if desc.GetNamespace() != key.GetNamespace() {
		return false
	}

	if desc.GetName() != key.GetName() {
		return false
	}

	cache := make(map[string]string)
	first := true

	// Descriptor must have a superset of the descriptors asked for in the key
	for _, qualifier := range key.GetQualifiers() {
		found := false

		if first {
			first = false

			for _, dQualifier := range desc.GetQualifiers() {
				cache[dQualifier] = dQualifier

				if qualifier == dQualifier {
					found = true
				}
			}
		} else {
			_, found = cache[qualifier]
		}

		if !found {
			return false
		}
	}

	return true
}

type allFilterData struct {
}

func (allFilterData) Filter(Descriptor) bool {
	return true
}

type idFilterData struct {
	locatorID int64
	serviceID int64
}

// NewIDFilter is a filter specific to a descriptor with
// the exact locatorID and serviceID given
func NewIDFilter(locatorID, serviceID int64) Filter {
	return &idFilterData{
		locatorID: locatorID,
		serviceID: serviceID,
	}
}

func (idFilter *idFilterData) Filter(desc Descriptor) bool {
	return (desc.GetLocatorID() == idFilter.locatorID) && (desc.GetServiceID() == idFilter.serviceID)
}
