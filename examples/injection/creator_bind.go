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

package injection

import (
	"github.com/jwells131313/dargo/ioc"
	"github.com/sirupsen/logrus"
)

const (
	// Example2LocatorName is the name given to the locators defined in example2
	Example2LocatorName = "Example2Locator"
	// EchoServiceName is the name of the echo service
	EchoServiceName = "EchoService_Name"
	// LoggerServiceName is the name of the logger service
	LoggerServiceName = "LoggerService_Name"
)

// EchoService is a service that logs the incoming string and
// then returns the string it was given (echo!)
type EchoService interface {
	Echo(string) string
}

type echoServiceData struct {
	logger *logrus.Logger
}

// CreateEchoLocator returns a ServiceLocator with the EchoService bound
// into it as well as a PerLookup logger service
func CreateEchoLocator() (ioc.ServiceLocator, error) {
	return ioc.CreateAndBind(Example2LocatorName, func(binder ioc.Binder) error {
		// binds the echo service into the locator in Singleton scope
		binder.BindWithCreator(EchoServiceName, newEchoService)

		// binds the logger service into the locator in PerLookup scope
		binder.BindWithCreator(LoggerServiceName, newLogger).InScope(ioc.PerLookup)

		return nil
	})
}

func newEchoService(locator ioc.ServiceLocator, key ioc.Descriptor) (interface{}, error) {
	logger, err := locator.GetDService(LoggerServiceName)
	if err != nil {
		return nil, err
	}

	return &echoServiceData{
		logger: logger.(*logrus.Logger),
	}, nil

}

func newLogger(ioc.ServiceLocator, ioc.Descriptor) (interface{}, error) {
	return logrus.New(), nil
}

func (echo *echoServiceData) Echo(in string) string {
	echo.logger.Printf("Echo got a string to log: %s", in)

	return in
}
