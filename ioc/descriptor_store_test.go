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
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	FOO = "foo"
	BAR = "bar"
	BAZ = "baz"
	QUX = "qux"

	NS1 = "ns1"
	NS2 = "ns2"
	NS3 = "ns3"
)

func TestAddRemoveDescriptorStore(t *testing.T) {
	cache := newNameCache()
	assert.NotNil(t, cache, "should not be nil")

	d1 := createDescriptor(NS1, FOO, 0, 0)
	d2 := createDescriptor(NS2, BAR, 0, 1)
	d3 := createDescriptor(NS2, BAZ, 0, 2)

	cache.add(d1)
	cache.add(d2)
	cache.add(d3)

	assert.True(t, cache.remove(d2))
	assert.True(t, cache.remove(d3))
	assert.True(t, cache.remove(d1))

	assert.False(t, cache.remove(d2))
	assert.False(t, cache.remove(d1))
	assert.False(t, cache.remove(d3))
}

func TestLookupAllDescriptorStore(t *testing.T) {
	cache := newNameCache()
	assert.NotNil(t, cache, "should not be nil")

	d1 := createDescriptor(NS1, FOO, 0, 0)
	d2 := createDescriptor(NS2, BAR, 0, 1)
	d3 := createDescriptor(NS2, BAZ, 0, 2)

	cache.add(d1)
	cache.add(d2)
	cache.add(d3)

	checkMe := cache.lookup(AllFilter)

	assert.Equal(t, 3, len(checkMe), "incorrect size")

	assert.Equal(t, d1, checkMe[0])
	assert.Equal(t, d2, checkMe[1])
	assert.Equal(t, d3, checkMe[2])
}

func TestLookupSpecific(t *testing.T) {
	cache := newNameCache()
	assert.NotNil(t, cache, "should not be nil")

	d1 := createDescriptor(NS1, FOO, 0, 0)
	d2 := createDescriptor(NS2, BAR, 0, 1)
	d3 := createDescriptor(NS2, BAZ, 0, 2)

	cache.add(d1)
	cache.add(d2)
	cache.add(d3)

	filter1 := NewSingleFilter(NS1, FOO)
	filter2 := NewSingleFilter(NS2, BAR)
	filter3 := NewSingleFilter(NS2, BAZ)

	rv := cache.lookup(filter1)
	assert.Equal(t, 1, len(rv))
	assert.Equal(t, d1, rv[0])

	rv = cache.lookup(filter2)
	assert.Equal(t, 1, len(rv))
	assert.Equal(t, d2, rv[0])

	rv = cache.lookup(filter3)
	assert.Equal(t, 1, len(rv))
	assert.Equal(t, d3, rv[0])

}

func TestClone(t *testing.T) {
	cache := newNameCache()
	assert.NotNil(t, cache, "should not be nil")

	d1 := createDescriptor(NS1, FOO, 0, 0)
	d2 := createDescriptor(NS2, BAR, 0, 1)
	d3 := createDescriptor(NS2, BAZ, 0, 2)

	cache.add(d1)
	cache.add(d2)
	cache.add(d3)

	clone := cache.clone()

	filter1 := NewSingleFilter(NS1, FOO)
	filter2 := NewSingleFilter(NS2, BAR)
	filter3 := NewSingleFilter(NS2, BAZ)

	rv := clone.lookup(filter1)
	assert.Equal(t, 1, len(rv))
	assert.Equal(t, d1, rv[0])

	rv = clone.lookup(filter2)
	assert.Equal(t, 1, len(rv))
	assert.Equal(t, d2, rv[0])

	rv = clone.lookup(filter3)
	assert.Equal(t, 1, len(rv))
	assert.Equal(t, d3, rv[0])

	// Check that removal from clone does not affect original
	assert.True(t, clone.remove(d1))

	rv = cache.lookup(filter1)
	assert.Equal(t, 1, len(rv))
	assert.Equal(t, d1, rv[0])

	// Check that removal from original does not affect clone
	assert.True(t, cache.remove(d2))

	rv = clone.lookup(filter2)
	assert.Equal(t, 1, len(rv))
	assert.Equal(t, d2, rv[0])
}

func TestFilterNotCalled(t *testing.T) {
	cache := newNameCache()
	assert.NotNil(t, cache, "should not be nil")

	d1 := createDescriptor(NS1, FOO, 0, 0)
	d2 := createDescriptor(NS2, BAR, 0, 1)
	d3 := createDescriptor(NS2, BAZ, 0, 2)

	cache.add(d1)
	cache.add(d2)
	cache.add(d3)

	filter := &panicyFilter{space: NS3, name: FOO}

	dNone := cache.lookup(filter)
	if !assert.Equal(t, 0, len(dNone)) {
		return
	}

	filter = &panicyFilter{space: NS2, name: QUX}

	dNone = cache.lookup(filter)
	if !assert.Equal(t, 0, len(dNone)) {
		return
	}
}

func createDescriptor(space, name string, lid, sid int64) Descriptor {
	key, err := NewServiceKey(space, name)
	if err != nil {
		panic(err)
	}

	tmp := NewConstantDescriptor(key, 0)
	retVal, err := NewDescriptor(tmp, sid, lid)
	if err != nil {
		panic(err)
	}

	return retVal
}

type panicyFilter struct {
	space, name string
}

func (pf *panicyFilter) GetNamespace() string {
	return pf.space
}

func (pf *panicyFilter) GetName() string {
	return pf.name
}

func (pf *panicyFilter) Filter(Descriptor) bool {
	panic("should not be called")
}
