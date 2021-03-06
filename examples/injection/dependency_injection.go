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
	"fmt"
	"github.com/jwells131313/dargo/ioc"
	"github.com/sirupsen/logrus"
)

// SimpleService is a test service
type SimpleService interface {
	// CallMe logs a message to the logger!
	CallMe()
}

// SimpleServiceData is a struct implementing SimpleService
// and which injects its logger
type SimpleServiceData struct {
	Log *logrus.Logger `inject:"LoggerService_Name"`
}

// CallMe implements the SimpleService method
func (ss *SimpleServiceData) CallMe() {
	ss.Log.Info("This logger was injected!")
}

func runExample() error {
	locator, err := ioc.CreateAndBind("example", func(binder ioc.Binder) error {
		binder.Bind("SimpleService", SimpleServiceData{})
		binder.BindWithCreator(LoggerServiceName, newLogger).InScope(ioc.PerLookup)
		return nil
	})
	if err != nil {
		return err
	}

	raw, err := locator.GetDService("SimpleService")
	if err != nil {
		return err
	}

	ss, ok := raw.(SimpleService)
	if !ok {
		return fmt.Errorf("Invalid type for simple service %v", ss)
	}

	ss.CallMe()

	return nil
}
