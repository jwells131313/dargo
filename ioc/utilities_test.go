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

func TestStack(t *testing.T) {
	stack := newStack()
	assert.NotNil(t, stack, "new returned nil")

	retNil, found := stack.Peek()
	assert.False(t, found, "new stack peek should return false")
	assert.Nil(t, retNil, "empty stack must return nil value")

	retNil, found = stack.Pop()
	assert.False(t, found, "new stack pop should return false")
	assert.Nil(t, retNil, "empty stack must return nil value")

	err := stack.Push(1)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = stack.Push(2)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = stack.Push(3)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	raw, found := stack.Peek()
	assert.True(t, found, "should have seen last value with peek")
	assert.Equal(t, 3, raw.(int), "should have last value pushed")

	raw, found = stack.Pop()
	assert.True(t, found, "should have seen last value with pop")
	assert.Equal(t, 3, raw.(int), "should have last value pushed")

	raw, found = stack.Peek()
	assert.True(t, found, "should have seen middle value with peek")
	assert.Equal(t, 2, raw.(int), "should have middle value pushed")

	raw, found = stack.Pop()
	assert.True(t, found, "should have seen middle value with pop")
	assert.Equal(t, 2, raw.(int), "should have middle value pushed")

	raw, found = stack.Peek()
	assert.True(t, found, "should have seen first value with peek")
	assert.Equal(t, 1, raw.(int), "should have first value pushed")

	raw, found = stack.Pop()
	assert.True(t, found, "should have seen first value with pop")
	assert.Equal(t, 1, raw.(int), "should have first value pushed")

	retNil, found = stack.Peek()
	assert.False(t, found, "empty stack peek should return false")
	assert.Nil(t, retNil, "empty stack must return nil value")

	retNil, found = stack.Pop()
	assert.False(t, found, "empty stack pop should return false")
	assert.Nil(t, retNil, "empty stack must return nil value")
}

func TestParseSimpleInjectString(t *testing.T) {
	pd, err := parseInjectString("foo")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, DefaultNamespace, "foo", nil, false) {
		return
	}
}

func TestParseInjectStringWithNamespace(t *testing.T) {
	pd, err := parseInjectString("space#foo")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, "space", "foo", nil, false) {
		return
	}
}

func TestParseInjectStringWithNamespaceAndOneQualifier(t *testing.T) {
	pd, err := parseInjectString("space#foo@bar")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, "space", "foo", []string{"bar"}, false) {
		return
	}
}

func TestParseInjectStringWithNamespaceAndThreeQualifiers(t *testing.T) {
	pd, err := parseInjectString("space#foo@bar@baz@qux")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, "space", "foo", []string{"bar", "baz", "qux"}, false) {
		return
	}
}

func TestParseInjectStringWithThreeQualifiers(t *testing.T) {
	pd, err := parseInjectString("foo@bar@baz@qux")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, DefaultNamespace, "foo", []string{"bar", "baz", "qux"}, false) {
		return
	}
}

func TestParseOptional(t *testing.T) {
	pd, err := parseInjectString("foo,optional")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, DefaultNamespace, "foo", nil, true) {
		return
	}
}

func TestParseOptionalNamespace(t *testing.T) {
	pd, err := parseInjectString("space#foo,optional")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, "space", "foo", nil, true) {
		return
	}
}

func TestParseOptionalNamespaceOneQualifier(t *testing.T) {
	pd, err := parseInjectString("space#foo@bar,optional")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, "space", "foo", []string{"bar"}, true) {
		return
	}
}

func TestParseOptionalNamespaceQualifiers(t *testing.T) {
	pd, err := parseInjectString("space#foo@bar@baz@qux@quux,optional")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, "space", "foo", []string{"bar", "baz", "qux", "quux"}, true) {
		return
	}
}

func TestParseOptionalQualifiers(t *testing.T) {
	pd, err := parseInjectString("foo@bar@baz,optional")
	if !assert.Nil(t, err) {
		return
	}
	if !checkParseData(t, pd, DefaultNamespace, "foo", []string{"bar", "baz"}, true) {
		return
	}
}

func TestParseInvalidOption(t *testing.T) {
	_, err := parseInjectString("foo@bar@baz,required")
	if !assert.NotNil(t, err) {
		return
	}
}

func checkParseData(t *testing.T, pd *parseData, space, name string, qualifiers []string, optional bool) bool {
	if !assert.NotNil(t, pd) {
		return false
	}
	if !assert.NotNil(t, pd.serviceKey) {
		return false
	}

	retVal1 := assert.Equal(t, space, pd.serviceKey.GetNamespace(), "namespace does not match")
	retVal2 := assert.Equal(t, name, pd.serviceKey.GetName(), "names did not match")

	retVal3 := true
	if qualifiers != nil {
		skQualifiers := pd.serviceKey.GetQualifiers()

		if !assert.Equal(t, len(skQualifiers), len(qualifiers), "number of qualifiers differs") {
			retVal3 = false
		} else {
			for index, qualifier := range qualifiers {
				if !assert.Equal(t, qualifier, skQualifiers[index], "qualifier at index %d was not equal", index) {
					retVal3 = false
				}
			}
		}
	}

	retVal4 := assert.Equal(t, optional, pd.isOptional, "optional value does not match")

	return retVal1 && retVal2 && retVal3 && retVal4
}
