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

// nameCache.  Locking is done via caller for ALL api
type nameCache struct {
	all  []Descriptor
	data map[string]map[string][]Descriptor
}

func newNameCache() *nameCache {
	return &nameCache{
		all:  make([]Descriptor, 0),
		data: make(map[string]map[string][]Descriptor),
	}
}

// add is add or replace
func (nc *nameCache) add(desc Descriptor) {
	nc.all = append(nc.all, desc)

	space := desc.GetNamespace()
	name := desc.GetName()

	internal, found := nc.data[space]
	if !found {
		internal = make(map[string][]Descriptor)
		nc.data[space] = internal
	}

	ar, found := internal[name]
	if !found {
		ar = make([]Descriptor, 0)
		internal[name] = ar
	}

	ar = append(ar, desc)
	internal[name] = ar
}

func getIndex(ar []Descriptor, desc Descriptor) int {
	lid := desc.GetLocatorID()
	sid := desc.GetServiceID()

	for index, d := range ar {
		dLid := d.GetLocatorID()
		dSid := d.GetServiceID()

		if dLid == lid && dSid == sid {
			return index
		}
	}

	return -1
}

func removeOneFromDescArray(ar []Descriptor, index int) []Descriptor {
	oneLessSize := len(ar) - 1
	oneLessAll := make([]Descriptor, oneLessSize)

	for lcv := 0; lcv < oneLessSize; lcv++ {
		if lcv < index {
			oneLessAll[lcv] = ar[lcv]
		} else {
			oneLessAll[lcv] = ar[lcv+1]
		}
	}

	return oneLessAll
}

func (nc *nameCache) remove(desc Descriptor) bool {
	index := getIndex(nc.all, desc)
	if index >= 0 {
		nc.all = removeOneFromDescArray(nc.all, index)
	}

	space := desc.GetNamespace()
	name := desc.GetName()

	internal, found := nc.data[space]
	if !found {
		return false
	}

	ar, found := internal[name]
	if !found {
		return false
	}

	index = getIndex(ar, desc)
	if index < 0 {
		return false
	}

	ar = removeOneFromDescArray(ar, index)
	if len(ar) == 0 {
		delete(internal, name)

		if len(internal) == 0 {
			delete(nc.data, space)
		}
	}

	return true
}

func (nc *nameCache) clone() *nameCache {
	cloneAll := make([]Descriptor, len(nc.all))
	copy(cloneAll, nc.all)

	retVal := make(map[string]map[string][]Descriptor)

	for space, internal := range nc.data {
		cp := make(map[string][]Descriptor)

		for name, descArray := range internal {
			cloneDescs := make([]Descriptor, len(descArray))
			copy(cloneDescs, descArray)

			cp[name] = cloneDescs
		}

		retVal[space] = cp
	}

	return &nameCache{
		all:  cloneAll,
		data: retVal,
	}
}

func (nc *nameCache) lookup(filter Filter) []Descriptor {
	space := filter.GetNamespace()
	name := filter.GetName()

	candidates := nc.all

	if space != "" && name != "" {
		internal, found := nc.data[space]
		if found {
			dArray, found := internal[name]
			if found {
				candidates = dArray
			} else {
				candidates = []Descriptor{}
			}
		} else {
			candidates = []Descriptor{}
		}
	}

	retVal := make([]Descriptor, 0)
	for _, candidate := range candidates {
		if filter.Filter(candidate) {
			retVal = append(retVal, candidate)
		}
	}

	return retVal
}

func checkFilter(filter Filter, desc Descriptor) bool {
	filterNamespace := filter.GetNamespace()
	filterName := filter.GetName()

	if filterNamespace != "" && filterName != "" {
		if desc.GetNamespace() != filterNamespace || desc.GetName() != filterName {
			return false
		}
	}

	return filter.Filter(desc)
}
