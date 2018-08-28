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
	"regexp"
)

// ServiceKey the key to a dargo managed service
type ServiceKey interface {
	GetNamespace() string
	GetName() string
	GetQualifiers() []string
}

// NewServiceKey Creates a new darget managed service key
func NewServiceKey(namespace, name string, qualifiers ...string) (ServiceKey, error) {
	err := checkNamespaceCharacters(namespace)
	if err != nil {
		return nil, err
	}

	err = checkNameCharacters(name)
	if err != nil {
		return nil, err
	}

	qs := make([]string, 0)
	for _, q := range qualifiers {
		err = checkNameCharacters(q)
		if err != nil {
			return nil, err
		}

		qs = append(qs, q)
	}

	return &serviceKeyData{
		namespace:  namespace,
		name:       name,
		qualifiers: qs,
	}, nil
}

type serviceKeyData struct {
	namespace, name string
	qualifiers      []string
}

func (key *serviceKeyData) GetNamespace() string {
	return key.namespace
}

func (key *serviceKeyData) GetName() string {
	return key.name
}

func (key *serviceKeyData) GetQualifiers() []string {
	return key.qualifiers
}

func (key *serviceKeyData) String() string {
	qPart := ""
	if len(key.qualifiers) > 0 {
		for _, qualifier := range key.qualifiers {
			qPart = qPart + "@" + qualifier
		}
	}

	return fmt.Sprintf("%s/%s%s", key.namespace, key.name, qPart)

}

// DSK creates a service key in the default namespace with the given name
func DSK(name string, qualifiers ...string) ServiceKey {
	retVal, err := NewServiceKey(DefaultNamespace, name, qualifiers...)
	if err != nil {
		panic(err)
	}

	return retVal
}

// SSK creates a service key in the system namespace with the given name
func SSK(name string, qualifiers ...string) ServiceKey {
	retVal, err := NewServiceKey(SystemNamespace, name, qualifiers...)
	if err != nil {
		panic(err)
	}

	return retVal
}

// CSK creates a service key in the contextual scope namespace with the given name
func CSK(name string, qualifiers ...string) ServiceKey {
	retVal, err := NewServiceKey(ContextualScopeNamespace, name, qualifiers...)
	if err != nil {
		panic(err)
	}

	return retVal
}

func newServiceKeyFromDescriptor(desc Descriptor) (ServiceKey, error) {
	return NewServiceKey(desc.GetNamespace(), desc.GetName(), desc.GetQualifiers()...)
}

func checkNamespaceCharacters(input string) error {
	if input == "" {
		return fmt.Errorf("The namespace may not be empty")
	}

	re := regexp.MustCompile("^[A-Za-z0-9_:/]*$")
	good := re.MatchString(input)
	if !good {
		return fmt.Errorf("The namespace may only have alphanumeric charaters and underscore, colon and slash (%s)",
			input)
	}

	return nil
}

func checkNameCharacters(input string) error {
	if input == "" {
		return fmt.Errorf("The name may not be empty")
	}

	re := regexp.MustCompile("^[A-Za-z0-9_]*$")
	good := re.MatchString(input)
	if !good {
		return fmt.Errorf("The name may only have alphanumeric charaters and underscore (%s)",
			input)
	}

	return nil
}
