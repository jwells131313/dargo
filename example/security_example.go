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

package example

import (
	"fmt"
	"github.com/jwells131313/dargo/ioc"
	"github.com/pkg/errors"
)

const (
	// ProtectedNamespace is a namespace that is allowed to inject protected services
	ProtectedNamespace = "user/protected"
	// SecureQualifier is the qualifer put onto services that should be protected
	SecureQualifier = "Secure"
)

type secureValidationService struct {
}

func (svs *secureValidationService) GetFilter() ioc.Filter {
	return ioc.AllFilter
}

func (svs *secureValidationService) GetValidator() ioc.Validator {
	return svs
}

func (svs *secureValidationService) Validate(info ioc.ValidationInformation) error {
	switch info.GetOperation() {
	case ioc.BindOperation:
		if info.GetCandidate().GetNamespace() == ProtectedNamespace {
			return fmt.Errorf("may not bind service into protected namespace")
		}
		break
	case ioc.UnbindOperation:
		break
	case ioc.LookupOperation:
		candidate := info.GetCandidate()
		if hasSecureQualifier(candidate) {
			// Those with Secure qualifier can only be injected into
			// services in the ProtectedNamespace and cannot be looked
			// up directly
			injectee := info.GetInjecteeDescriptor()
			if injectee == nil {
				return fmt.Errorf("Secure services cannot be looked up directly")
			} else if injectee.GetNamespace() != ProtectedNamespace {
				return fmt.Errorf("Secure service can only be injected into special services")
			}
		}
		break
	default:
	}

	return nil
}

func hasSecureQualifier(desc ioc.Descriptor) bool {
	for _, q := range desc.GetQualifiers() {
		if q == SecureQualifier {
			return true
		}
	}
	return false
}

// SuperSecretService is an example protected service.  It will be annotated with "Secure"
type SuperSecretService struct {
}

// ServiceData an example service injecting a protected service
type ServiceData struct {
	ProtectedService *SuperSecretService `inject:"SuperSecretService"`
}

func runSecurityExample() error {
	locator, err := ioc.CreateAndBind("SecurityExampleLocator", func(binder ioc.Binder) error {
		binder.Bind(ioc.ValidationServiceName, secureValidationService{}).InNamespace(ioc.UserServicesNamespace)
		binder.Bind("SuperSecretService", SuperSecretService{}).QualifiedBy(SecureQualifier)
		binder.Bind("SystemProtectedService", ServiceData{}).InNamespace(ProtectedNamespace)
		binder.Bind("NormalUserService", ServiceData{})

		return nil
	})
	if err != nil {
		return err
	}

	// A lookup of this service should appear as if the service isn't there
	_, err = locator.GetDService("SuperSecretService")
	if err == nil {
		return fmt.Errorf("service should have appeared to not exist")
	}

	// Should not be able to bind a service into protected namespace
	err = ioc.BindIntoLocator(locator, func(binder ioc.Binder) error {
		binder.Bind("NotAllowedToBindThis", ServiceData{}).InNamespace(ProtectedNamespace)
		return nil
	})
	if err == nil {
		return fmt.Errorf("should not have been able to bind service into protected namespace")
	}

	// The NormalService should not be able to inject the secret service since its not
	// in the ProtectedNamespace
	_, err = locator.GetDService("NormalUserService")
	if err == nil {
		return fmt.Errorf("should not have been able to create normal service with secure injection point")
	}

	// The service in the protected namespace injecting the secure service should work
	key, _ := ioc.NewServiceKey(ProtectedNamespace, "SystemProtectedService")

	serviceRaw, err := locator.GetService(key)
	if err != nil {
		return errors.Wrapf(err, "Did not find %v", key)
	}
	service := serviceRaw.(*ServiceData)

	if service.ProtectedService == nil {
		return fmt.Errorf("protected service should be available")
	}

	return nil
}
