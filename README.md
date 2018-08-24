[//]: # " DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS HEADER. "
[//]: # "  "
[//]: # " Copyright (c) 2018 Oracle and/or its affiliates. All rights reserved. "
[//]: # "  "
[//]: # " The contents of this file are subject to the terms of either the GNU "
[//]: # " General Public License Version 2 only (''GPL'') or the Common Development "
[//]: # " and Distribution License(''CDDL'') (collectively, the ''License'').  You "
[//]: # " may not use this file except in compliance with the License.  You can "
[//]: # " obtain a copy of the License at "
[//]: # " https://oss.oracle.com/licenses/CDDL+GPL-1.1 "
[//]: # " or LICENSE.txt.  See the License for the specific "
[//]: # " language governing permissions and limitations under the License. "
[//]: # "  "
[//]: # " When distributing the software, include this License Header Notice in each "
[//]: # " file and include the License file at LICENSE.txt. "
[//]: # "  "
[//]: # " GPL Classpath Exception: "
[//]: # " Oracle designates this particular file as subject to the ''Classpath'' "
[//]: # " exception as provided by Oracle in the GPL Version 2 section of the License "
[//]: # " file that accompanied this code. "
[//]: # "  "
[//]: # " Modifications: "
[//]: # " If applicable, add the following below the License Header, with the fields "
[//]: # " enclosed by brackets [] replaced by your own identifying information: "
[//]: # " ''Portions Copyright [year] [name of copyright owner]'' "
[//]: # "  "
[//]: # " Contributor(s): "
[//]: # " If you wish your version of this file to be governed by only the CDDL or "
[//]: # " only the GPL Version 2, indicate your decision by adding ''[Contributor] "
[//]: # " elects to include this software in this distribution under the [CDDL or GPL "
[//]: # " Version 2] license.''  If you don't indicate a single choice of license, a "
[//]: # " recipient has the option to distribute your version of this file under "
[//]: # " either the CDDL, the GPL Version 2 or to extend the choice of license to "
[//]: # " its licensees as provided above.  However, if you add GPL Version 2 code "
[//]: # " and therefore, elected the GPL Version 2 license, then the option applies "
[//]: # " only if the new code is made subject to such option by the copyright "
[//]: # " holder. "

# dargo [![GoDoc](https://godoc.org/github.com/jwells131313/dargo/ioc?status.svg)](https://godoc.org/github.com/jwells131313/dargo/ioc) [![wercker status](https://app.wercker.com/status/24379824ff4ec7e885f37323e261a36b/s/master "wercker status")](https://app.wercker.com/project/byKey/24379824ff4ec7e885f37323e261a36b)

Dynamic Service Registry and Inversion of Control for GO

## Service Registry

Dargo is an in-memory service registry for GO.  It also introduces inversion of control, in that once the
service descriptions are bound into the registry they are created in response to registry lookups.  Services
are scoped by context, and so are created based on the lifecycle defined by the scope.  For
example a service in the Singleton scope are only every created once.  A service in the PerLookup
scope are created every time they are looked up.

NOTE:  The current version of this API is 0.1.0.  This means that the API has
not settled completely and may change in future revisions.  Once the dargo
team has decided the API is good as it is we will make the 1.0 version which
will have some backward compatibility guarantees.  In the meantime, if you
have questions or comments please open issues.  Thank you.

The general flow of an application that uses dargo is to:

1.  Create a ServiceLocator
2.  Bind services into the ServiceLocator
3.  Use the ServiceLocator in your code to find services

Services can depend on other services.  When a service is created first all of its dependencies can be created with
the same ServiceLocator.

There can be multiple implementations of the same service, and there are specific rules
for choosing the best service amongst all of the possible choices.  In some cases services can be differentiated
by qualifiers.  In other cases services can be given ranks, with higher ranks being chosen over lower ranks.

Using dargo helps unit test your code as it becomes easy to replace services served by the locator with mocks.
If you ensure that your test mocks have a higher rank than the service bound by your normal code then
all of your internal code will use the mock from the ServiceLocator rather than the original service.

### An Example

In the following example we will bind two services.  One provides an EchoService and is in the Singleton
scope, while the other is a logger service and is in the PerLookup scope.  First, here is the definition
and implementation of the EchoService:

```go
// EchoService is a service that logs the incoming string and
// then returns the string it was given (echo!)
type EchoService interface {
	Echo(string) string
}

type echoServiceData struct {
	logger *logrus.Logger
}

func (echo *echoServiceData) Echo(in string) string {
	echo.logger.Printf("Echo got a string to log: %s", in)

	return in
}
```

To allow dargo to create the EchoService the user must supply a creation function.  The creation
function is passed a ServiceLocator to be used to find other services it may depend on and the
ServiceKey that describes the service further.  This is the creation function for the EchoService:

```go
func newEchoService(locator ioc.ServiceLocator, key ioc.ServiceKey) (interface{}, error) {
	logger, err := locator.GetDService(LoggerServiceName)
	if err != nil {
		return nil, err
	}

	return &echoServiceData{
		logger: logger.(*logrus.Logger),
	}, nil

}
```

The code above used the ServiceLocator method GetDService to get the LoggerService.  The
method GetDService returns services in the default service namespace (more about service names later).
It then gives that service to the echo data structure that is returned.

Here is the creation function for the logger service:

```go
import "github.com/sirupsen/logrus"

func newLogger(ioc.ServiceLocator, ioc.ServiceKey) (interface{}, error) {
	return logrus.New(), nil
}
```

Now that we have our services defined and our creator functions written we can create a
ServiceLocator and bind those services.  This is the method that does that:

```go
// CreateEchoLocator returns a ServiceLocator with the EchoService bound
// into it as well as a PerLookup logger service
func CreateEchoLocator() (ioc.ServiceLocator, error) {
	
	// Use CreateAndBind to create and bind services all at once!
	return ioc.CreateAndBind(Example2LocatorName, func(binder ioc.Binder) error {
		
		// binds the echo service into the locator in Singleton scope
		binder.Bind(EchoServiceName, newEchoService)

		// binds the logger service into the locator in PerLookup scope
		binder.Bind(LoggerServiceName, newLogger).InScope(ioc.PerLookup)

		return nil
	})
}
```

The CreateAndBind method both creates a ServiceLocator and takes a binder function into which a
Binder is passed for use in binding services.  It is important to note that the services are **not**
created at this time, rather a description of the service is put into the ServiceLocator.  Services
are normally created when they are requested depending on the rules of the scope.  Singleton services
are created the first time they are asked for, while PerLookup services are created every time someone
looks the service up.

You can now look up and use the echo service, as shown in the following test code:

```go
func TestExample2(t *testing.T) {
	locator, err := CreateEchoLocator()
	if err != nil {
		t.Error("could not create locator")
		return
	}

	rawService, err := locator.GetDService(EchoServiceName)
	if err != nil {
		t.Errorf("could not find echo service %v", err)
		return
	}

	echoService, ok := rawService.(EchoService)
	if !ok {
		t.Errorf("raw echo service was not the correct type %v", rawService)
		return
	}

	ret := echoService.Echo("hi")
	if ret != "hi" {
		t.Errorf("did not get expected reply: %s", ret)
	}
}
``` 

When the test code does "locator.GetDService(EchoServiceName)" the create method for the EchoService will be
invoked, which will in turn lookup the logger service, which, since it is in the PerLookup scope, will always
return a new one.  Subsequent lookups of the EchoService, however, will return the **same** EchoService, since
the EchoService is in the Singleton scope.

### Service Names

Every service bound into the ServiceLocator has a name.  The names are scoped by a namespace.  There is
a default namespace which is sufficient for most use cases.  However, there are
other special name spaces such as, "system", used for system services, and "sys/scope", used for special
ContextualScope services

The allowed characters for a name are alphanumeric and _.  The allowed characters for a namespace
are alphanumeric, _, and ":".  Qualifiers have the same restrictions as the name.

The ServiceKey interface represents a full service key:

```go
// ServiceKey the key to a dargo managed service
type ServiceKey interface {
	GetNamespace() string
	GetName() string
	GetQualifiers() []string
}
```

There are helper methods for generating ServiceKeys from simple strings.  Also the ServiceLocator
has a method GetDService which always uses the default namespace to find services.  Here
are the helper method signatures for creating ServiceKeys:

```go
// DSK creates a service key in the default namespace with the given name
func DSK(name string, qualifiers ...string) ServiceKey {...}

// SSK creates a service key in the system namespace with the given name
func SSK(name string, qualifiers ...string) ServiceKey {...}

// CSK creates a service key in the contextual scope namespace with the given name
func CSK(name string, qualifiers ...string) ServiceKey {...}
```

