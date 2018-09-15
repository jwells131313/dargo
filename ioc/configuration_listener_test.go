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
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ConfigListenerLocator1 = "ConfigListenerLocator1"

	Q1 = "Q1"
	Q2 = "Q2"
)

func TestListenerCalledForAddAndRemove(t *testing.T) {
	locator, err := CreateAndBind(ConfigListenerLocator1, func(binder Binder) error {
		binder.Bind(SimpleServiceName, &SimpleService{}).QualifiedBy(Q1)
		binder.Bind(ConfigurationListenerName, &ConfigListenerServiceData{}).InNamespace(UserServicesNamespace)

		return nil
	})
	if !assert.Nil(t, err, "could not create locator") {
		return
	}

	key, err := NewServiceKey(UserServicesNamespace, ConfigurationListenerName)
	if !assert.Nil(t, err, "could not create key locator") {
		return
	}

	cListenerRaw, err := locator.GetService(key)
	if !assert.Nil(t, err, "could not get listener") {
		return
	}

	cListener, ok := cListenerRaw.(ConfigurationListener)
	if !assert.True(t, ok, "invalid type") {
		return
	}

	if !verifyListener(t, cListener, []string{}, []string{}) {
		return
	}

	err = BindIntoLocator(locator, func(binder Binder) error {
		binder.Bind(SimpleServiceName, &SimpleService{}).QualifiedBy(Q2)
		return nil
	})
	if !assert.Nil(t, err, "dynamic bind failure %v", err) {
		return
	}

	if !verifyListener(t, cListener, []string{Q2}, []string{}) {
		return
	}

	err = UnbindDServices(locator, SimpleServiceName)
	if !assert.Nil(t, err, "dynamic unbind failure %v", err) {
		return
	}

	if !verifyListener(t, cListener, []string{}, []string{Q1, Q2}) {
		return
	}

}

func verifyListener(t *testing.T, listener ConfigurationListener, added []string, removed []string) bool {
	specific, ok := listener.(*ConfigListenerServiceData)
	if !assert.True(t, ok, "unexpected listener") {
		return false
	}

	if !assert.Equal(t, len(added), len(specific.lastAdded), "incorrect addition size") {
		return false
	}

	for lcv := 0; lcv < len(added); lcv++ {
		_, found := specific.lastAdded[added[lcv]]
		if !assert.True(t, found, "Did not get expected addition %s", added[lcv]) {
			return false
		}
	}

	if !assert.Equal(t, len(removed), len(specific.lastRemoved), "incorrect removal size") {
		return false
	}

	for lcv := 0; lcv < len(removed); lcv++ {
		_, found := specific.lastRemoved[removed[lcv]]
		if !assert.True(t, found, "did not get expected removal %s", removed[lcv]) {
			return false
		}
	}

	return true
}

type ConfigListenerServiceData struct {
	Locator        ServiceLocator `inject:"system#ServiceLocator"`
	simpleServices map[Descriptor]string
	lastAdded      map[string]Descriptor
	lastRemoved    map[string]Descriptor
}

func (listener *ConfigListenerServiceData) DargoInitialize(desc Descriptor) error {
	listener.lastAdded = make(map[string]Descriptor)
	listener.lastRemoved = make(map[string]Descriptor)

	f := NewServiceKeyFilter(DSK(SimpleServiceName))

	simples, err := listener.Locator.GetDescriptors(f)
	if err != nil {
		return err
	}

	listener.simpleServices = make(map[Descriptor]string)
	for _, aSimple := range simples {
		listener.simpleServices[aSimple] = aSimple.GetFullName()
	}

	return nil
}

func (listener *ConfigListenerServiceData) ConfigurationChanged() {
	f := NewServiceKeyFilter(DSK(SimpleServiceName))

	removedDescriptors := make(map[Descriptor]string)
	for k, v := range listener.simpleServices {
		removedDescriptors[k] = v
	}

	currentSimples, err := listener.Locator.GetDescriptors(f)
	if err != nil {
		fmt.Printf("Error getting descriptors from ConfigurationChanged %v\n", err)
		return
	}

	currentAsMap := make(map[Descriptor]string)
	for _, current := range currentSimples {
		currentAsMap[current] = current.GetFullName()
	}

	added := make(map[string]Descriptor)
	for desc, _ := range currentAsMap {
		delete(removedDescriptors, desc)
		_, found := listener.simpleServices[desc]
		if !found {
			zero := getZeroQualifier(desc)
			if zero == "" {
				panic("Should not have empty qualifier for addition")
			}

			added[zero] = desc
		}
	}

	removed := make(map[string]Descriptor)
	for desc, _ := range removedDescriptors {
		zero := getZeroQualifier(desc)
		if zero == "" {
			panic("Should not have empty qualifier for removal")
		}

		removed[zero] = desc
	}

	listener.simpleServices = currentAsMap
	listener.lastAdded = added
	listener.lastRemoved = removed
}

func getZeroQualifier(desc Descriptor) string {
	if desc == nil {
		return ""
	}

	qualifiers := desc.GetQualifiers()
	if qualifiers == nil {
		return ""
	}

	if len(qualifiers) == 0 {
		return ""
	}

	return qualifiers[0]
}
