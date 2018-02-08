package utilities
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

import (
    "testing"
    "strings"
)

func TestLocatorFactoryWithName(t *testing.T) {
	slFactory, _ := GetSystemLocatorFactory()
	
	locator, found := slFactory.FindOrCreateRootLocator("test1")
	
	lName := locator.GetName()
	
	if lName != "test1" {
		t.Errorf("Name returned from factory is incorrect, expected test1 got %s", lName)
	}
	
	if found == true {
		t.Errorf("This should have been a fresh locator, but created returned false")
	}
}

func TestLocatorFactoryWithNoName(t *testing.T) {
	slFactory, _ := GetSystemLocatorFactory()
	
	locator, found := slFactory.FindOrCreateRootLocator("")
	
	lName := locator.GetName()
	
	if !strings.Contains(lName, "__Generated_Service_Locator_Name_") {
		t.Errorf("Name returned from factory is incorrect, expected test1 got %s", lName)
	}
	
	if found == true {
		t.Errorf("This should have been a fresh locator, but created returned false")
	}
	
	locator2, found2 := slFactory.FindOrCreateRootLocator("")
	
	lName2 := locator2.GetName()
	
	if !strings.Contains(lName2, "__Generated_Service_Locator_Name_") {
		t.Errorf("Name returned from factory is incorrect, expected test1 got %s", lName2)
	}
	
	if found2 == true {
		t.Errorf("This should have been a fresh locator, but created returned false")
	}
	
	if lName == lName2 {
		t.Errorf("Generated names should not match (%s/%s)", lName, lName2)
	}
	
	id1 := locator.GetID()
	id2 := locator2.GetID()
	
	if id1 == id2 {
		t.Errorf("The id's of returned locators should never match (%v/%v)", id1, id2)
	}
}

