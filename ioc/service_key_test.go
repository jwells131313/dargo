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
	foo = "foo"
	bar = "bar"

	red   = "red"
	blue  = "blue"
	green = "green"
)

func TestNewSK(t *testing.T) {
	sk := NewServiceKey(foo, bar)

	assert.Equal(t, sk.GetNamespace(), foo, "namespace should be equal")
	assert.Equal(t, sk.GetName(), bar, "name should be equal")
	assert.Equal(t, len(sk.GetQualifiers()), 0, "should be no qualifiers")

	sk1 := NewServiceKey(foo, bar, red, blue, green)

	assert.Equal(t, sk1.GetNamespace(), foo, "namespace should be equal (2)")
	assert.Equal(t, sk1.GetName(), bar, "name should be equal (2)")

	assert.Equal(t, len(sk1.GetQualifiers()), 3, "should be 3 qualifiers")
	assert.Equal(t, red, sk1.GetQualifiers()[0], "should be red at 0 index")
	assert.Equal(t, blue, sk1.GetQualifiers()[1], "should be blue at 1 index")
	assert.Equal(t, green, sk1.GetQualifiers()[2], "should be green at 2 index")
}

func TestDSK(t *testing.T) {
	sk := DSK(bar)

	assert.Equal(t, sk.GetNamespace(), DefaultNamespace, "default namespace should be equal")
	assert.Equal(t, sk.GetName(), bar, "default name should be equal")
	assert.Equal(t, len(sk.GetQualifiers()), 0, "should be no qualifiers")

	sk1 := DSK(bar, red, blue, green)

	assert.Equal(t, sk1.GetNamespace(), DefaultNamespace, "default namespace should be equal (2)")
	assert.Equal(t, sk1.GetName(), bar, "default name should be equal (2)")

	assert.Equal(t, len(sk1.GetQualifiers()), 3, "default should be 3 qualifiers")
	assert.Equal(t, red, sk1.GetQualifiers()[0], "default should be red at 0 index")
	assert.Equal(t, blue, sk1.GetQualifiers()[1], "default should be blue at 1 index")
	assert.Equal(t, green, sk1.GetQualifiers()[2], "default should be green at 2 index")
}

func TestSSK(t *testing.T) {
	sk := SSK(bar)

	assert.Equal(t, sk.GetNamespace(), SystemNamespace, "system namespace should be equal")
	assert.Equal(t, sk.GetName(), bar, "system name should be equal")
	assert.Equal(t, len(sk.GetQualifiers()), 0, "should be no qualifiers")

	sk1 := SSK(bar, red, blue, green)

	assert.Equal(t, sk1.GetNamespace(), SystemNamespace, "system namespace should be equal (2)")
	assert.Equal(t, sk1.GetName(), bar, "system name should be equal (2)")

	assert.Equal(t, len(sk1.GetQualifiers()), 3, "system should be 3 qualifiers")
	assert.Equal(t, red, sk1.GetQualifiers()[0], "system should be red at 0 index")
	assert.Equal(t, blue, sk1.GetQualifiers()[1], "system should be blue at 1 index")
	assert.Equal(t, green, sk1.GetQualifiers()[2], "system should be green at 2 index")
}
