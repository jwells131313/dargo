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

# dargo [![GoDoc](https://godoc.org/github.com/jwells131313/dargo/ioc?status.svg)](https://godoc.org/github.com/jwells131313/dargo/ioc) [![wercker status](https://app.wercker.com/status/24379824ff4ec7e885f37323e261a36b/s/master "wercker status")](https://app.wercker.com/project/byKey/24379824ff4ec7e885f37323e261a36b) [![Go Report Card](https://goreportcard.com/badge/github.com/jwells131313/dargo)](https://goreportcard.com/report/github.com/jwells131313/dargo)

Dependency Injector for GO

## Depenency Injector

Dargo is an depenency injection system for GO.

Dargo services are scoped and are created and destroyed based on the defined lifecycle of the
scope.  For example services in the Singleton scope are only created once.  Services in the PerLookup
scope are created every time they are injected.

NOTE:  The current version of this API is 0.3.0.  This means that the API has
not settled completely and may change in future revisions.  Once the dargo
team has decided the API is good as it is we will make the 1.0 version which
will have some backward compatibility guarantees.  In the meantime, if you
have questions or comments please open issues.  Thank you.

## Table of Contents

1.  [Basic Usage](#basic-usage)
2.  [Testing](#testing)
3.  [Service Names](#service-names)
4.  [Context Scope](#context-scope)
5.  [Provider](#provider)
6.  [Error Service](#error-service)
7.  [Security](#validation-service)

## Basic Usage

The general flow of an application that uses dargo is to:

1.  Create a ServiceLocator
2.  Bind services into the ServiceLocator
3.  Use the ServiceLocator in your code to find services
4.  Any dependent services of the found service are automaticially injected

Services can depend on other services.  When a service is created first all of its dependencies are
created.  A service binding can either provide a method with which to create
the service, or it can use the automatic injection capability of dargo.

There can be multiple implementations of the same service, and there are specific rules
for choosing the best service amongst all of the possible choices.  In some cases services can be differentiated
by qualifiers.  In other cases services can be given ranks, with higher ranks being chosen over lower ranks.

Using dargo helps unit test your code as it becomes easy to replace services served by the locator with mocks.
If you ensure that your test mocks have a higher rank than the service bound by your normal code then
all of your internal code will use the mock from the ServiceLocator rather than the original service.

### Injection Example

In this example a service called SimpleService will inject a logger.  The logger itself is a dargo
service that is bound with a creation method.  That creation method looks like this:

```go
func newLogger(ioc.ServiceLocator, ioc.Descriptor) (interface{}, error) {
	return logrus.New(), nil
}
```

The binding of SimpleService will provide the struct that should be used to implement the interface.  The
struct has a field annotated with _inject_ followed by the name of the service to inject.  This
is the interface and the struct used to implement it:

```go
type SimpleService interface {
	// CallMe logs a message to the logger!
	CallMe()
}

// SimpleServiceData is a struct implementing SimpleService
type SimpleServiceData struct {
	Log *logrus.Logger `inject:"LoggerService_Name"`
}

// CallMe implements the SimpleService method
func (ssd *SimpleServiceData) CallMe() {
	ssd.Log.Info("This logger was injected!")
}
```

Both the logger service and the SimpleService are bound into the ServiceLocator.  This is normally done near
the start of your program:

```go
locator, err := ioc.CreateAndBind("InjectionExampleLocator", func(binder ioc.Binder) error {
	    // Binds SimpleService by providing the structure
	    binder.Bind("SimpleService", SimpleServiceData{})

	    // Binds the logger service by providing the creation function 
	    binder.BindWithCreator("LoggerService_Name", newLogger).InScope(ioc.PerLookup)
	    return nil
    })
```

The returned locator can be used to lookup the SimpleService service.  The SimpleService is bound
into the Singleton scope (the default scope), which means that it will only be created the first
time it is looked up or injected, and never again.  The LoggerService, on the other hand is in the
PerLookup scope, which means that every time it is injected or looked up a new one will be created.

This is the code that uses the looked up service:

```go
    raw, err := locator.GetDService("SimpleService")
	if err != nil {
		return err
	}

	ss, ok := raw.(SimpleService)
	if !ok {
		return fmt.Errorf("Invalid type for simple service %v", ss)
	}

	ss.CallMe()
```

Any depth of injection is supported (ServiceA can depend on ServiceB which depends on ServiceC and so on).
A service can also depend on as many services as it would like (ServiceA can depend on service D, E and F etc).
Howerver, services cannot have circular dependencies.

### Another Example

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

To allow Dargo to create the EchoService the user must supply a creation function.  The creation
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
		binder.BindWithCreator(EchoServiceName, newEchoService)

		// binds the logger service into the locator in PerLookup scope
		binder.BindWithCreator(LoggerServiceName, newLogger).InScope(ioc.PerLookup)

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

## Testing

Unit testing becomes easier with Dargo services due to the dynamic nature of Dargo services and the fact
that the choice of service used can be affected with the Rank of the service.  You can create mock versions
of any of the services bound into a ServiceLocator and then bind them into the ServiceLocator your system uses
with a higher rank.  When you then run your code in your unit tests the mock services will be chosen instead
of the services that would normally be injected.

### Testing Example

In this example we have a service that has an expensive operation.

```go
type AnExpensiveService interface {
	DoExpensiveThing(string) (string, error)
}
```

We then have a normal version of that service that is implemented in the normal user code.  In this example
the expensive thing merely sleeps and returns "Normal"

```go
type NormalExpensiveServiceData struct {
	// whatever stuff is in here
}

func (nesd *NormalExpensiveServiceData) DoExpensiveThing(thingToDo string) (string, error) {
	// Do something expensive
	time.Sleep(5 * time.Second)

	return "Normal", nil
}
```

This struct injects an instance of AnExpensiveService.  A method on it uses the expensive service and returns
the result.

```go
type SomeOtherServiceData struct {
	ExpensiveService AnExpensiveService `inject:"AnExpensiveService"`
}

func (other *SomeOtherServiceData) DoSomeUserCode() (string, error) {
	// In user code this will be the real service, in test code this will be the mock
	return other.ExpensiveService.DoExpensiveThing("foo")
}
```

In the users code other.ExpensiveService will be injected as the normal, truly expensive service.  The binding of
these normal services happen in the following initialization block, which is where most Dargo ServiceLocators
are created and wired.

```go
var globalLocator ioc.ServiceLocator

func init() {
	myLocator, err := ioc.CreateAndBind("TestingExampleLocator", func(binder ioc.Binder) error {
		binder.Bind("UserService", SomeOtherServiceData{})
		// Bound with default rank of 0
		binder.Bind("AnExpensiveService", NormalExpensiveServiceData{})

		return nil
	})
	if err != nil {
		panic(err)
	}

	globalLocator = myLocator
}
```

The ExpensiveService is bound with the default rank of 0.  Ranks can have positive or negative values.  Higher
ranks are preferred above lower ranks.  Ranking order is even honored when getting all instances of a service,
so higher ranked services will appear first in the slice and lower ranked services will appear later in the slice.

Now we want to test UserService.  But UserService normally injects the ExpensiveService.  This is not appropriate
for this unit test.  Maybe the ExpensiveService contacts a database, or maybe the ExpensiveService requires
manual input normally.  In the test code we want to mock this service.  Luckily, in the test code we can
bind a service with rank 1 or higher, and then that Mock service will be preferred over the normal code.

Here is the full Mock code for AnExpensiveService from the test file:

```go
type MockExpensiveService struct {
}

func (mock *MockExpensiveService) DoExpensiveThing(thingToDo string) (string, error) {
	// This service doesn't really do anything, but does return a different answer that can be checked
	return "Mock", nil
}
```

Here is the full test code, including the code that binds the mock service into the ServiceLocator with
a Rank of 1, which will cause the mock to get injected in favor of the normal service:

```go
func putMocksIn() error {
	return ioc.BindIntoLocator(globalLocator, func(binder ioc.Binder) error {
		binder.Bind("AnExpensiveService", MockExpensiveService{}).Ranked(1)

		return nil
	})
}

func TestWithAMock(t *testing.T) {
	err := putMocksIn()
	if err != nil {
		t.Error(err.Error())
		return
	}

	raw, err := globalLocator.GetDService("UserService")
	if err != nil {
		t.Error(err.Error())
		return
	}

	userService := raw.(*SomeOtherServiceData)

	result, err := userService.DoSomeUserCode()
	if err != nil {
		t.Error(err.Error())
		return
	}

	if result != "Mock" {
		t.Errorf("Was expecting mock service but got %s", result)
		return
	}
}
```

Using a dependency injection framework like Dargo means having a lot of flexibility when unit testing and
can therefor lead to higher code coverage of your tests.

## Service Names

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

You can also use complex names in the inject key of structures.  The general format is:

```
namespace#name@qualifier1@qualifier2
```

Only the name part is required.  For example, if you wanted to inject a service
named ColorService in the visible/light namespace with qualifier Green, you would do
something like this:

```go
type Service struct {
	Green ColorService `inject:"visible/light#ColorService@Green"`
}
```

## Context Scope

Many go programs use context.Context to get their services.  Dargo provides an optional
Context scope called DargoContext which can associate a ServiceLocator with a context.Context.  With the
DargoContext scope programs can continue to use context.Context and be getting all the dependency-injection
goodness from Dargo.

The definition of the lifecycle of the DargoContext scope is that of the underlying parent context.Context.  
When the parent context.Context is finished all of the Dargo services associated with that context.Context
will be destroyed.  For example, if you have a per-request context.Context, you can use that as the parent
for the DargoContext scope.  Every service that is bound into the DargoContext scope will be unique per request
and will be destroyed when the request has been finished.

To enable the DargoContext scope the method ioc.EnableDargoContextScope must be called.  This method
will add in the DargoContext ContextualScope implementation.   It also adds a DargoContext scoped
service named _DargoContextCreationService_ (ioc.DargoContextCreationServiceName) to the ServiceLocator.
The DargoContextCreationService is a convenient service that returns the DargoContext context.Context
under which the service was created.

#### Context Scope Example

This example creates a  Per-Request context.Context that carries the name of the user in
the value.  That Per-Request context.Context is wrapped by a DargoContext which has a
Per-Context AuthorizationService.  The AuthorizationService uses the context.Context with
which it was created to get the username, and uses that username to decide if the user
can proceed.

We will not go into the details of creating the Per-Request context, but the code for this
example can be found in the context_example.go file in the examples subdirectory.

First lets see the definition of the AuthorizationService and a corresponding structure that
imlements the interface:

```go
// AuthorizationService is an example authorization service
type AuthorizationService interface {
	// MotherMayI asks the auth service for permission to do something
	MotherMayI() bool
}

// AuthorizationServiceData is the struct implementing AuthorizationService
// It injects the DargoContextCreationService to get the context under
// which this service was created
type AuthorizationServiceData struct {
	ContextService ioc.DargoContextCreationService `inject:"DargoContextCreationService"`
}
```

The implementation of AuthorizationService just lets anyone do anything, except for Mallory:

```go
// MotherMayI allows everyone to do everything except Mallory, who isn't allowed to do anything
func (asd *AuthorizationServiceData) MotherMayI() bool {
	context := asd.ContextService.GetDargoCreationContext()

	userRaw := context.Value(userNameKey)
	if userRaw == nil {
		return false
	}

	user := userRaw.(string)

	if user == "Mallory" {
		return false
	}

	return true
}
```

Now lets see how our initial creation of this ServiceLocator would look:

```go
locator, _ := ioc.CreateAndBind("ContextExample", func(binder ioc.Binder) error {
    binder.Bind("AuthService", AuthorizationServiceData{}).InScope(ioc.ContextScope)

    return nil
})

ioc.EnableDargoContextScope(locator)
```

Services in the DargoContext scope must be looked up from the context.Context, not through
the ServiceLocator.  So in order to get an instance of the AuthorizationService the context must
be used.  Here is some example code of creating a few requests with different users and then
using the AuthorizationService to grant them access:

```go
func createContext(locator ioc.ServiceLocator, user string) (context.Context, context.CancelFunc, error) {
	parent, canceller := context.WithCancel(context.Background())

	requestScoped := NewRequestContext(parent, user)

	dargoContext, err := ioc.NewDargoContext(requestScoped, locator)

	return dargoContext, canceller, err
}

func getAuthorizeService(context context.Context) AuthorizationService {
	raw := context.Value("AuthService")
	if raw == nil {
		return nil
	}

	return raw.(AuthorizationService)
}

func runExample() error {
	// other code
	
    aliceContext, aliceCanceller, _ := createContext(locator, "Alice")
    defer aliceCanceller()
    
    bobContext, bobCanceller, _ := createContext(locator, "Bob")
    defer bobCanceller()
    
    malloryContext, malloryCanceller, _ := createContext(locator, "Mallory")
    defer malloryCanceller()
    
    aliceAuthorizer := getAuthorizeService(aliceContext)
    bobAuthorizer := getAuthorizeService(bobContext)
    malloryAuthorizer := getAuthorizeService(malloryContext)
    
    canI := aliceAuthorizer.MotherMayI()
    if !canI {
        return fmt.Errorf("Alice should have been able to go")
    }
    
    canI = bobAuthorizer.MotherMayI()
    if !canI {
        return fmt.Errorf("Alice should have been able to go")
    }
    
    canI = malloryAuthorizer.MotherMayI()
    if canI {
        // Mallory should have NOT been allowed
        return fmt.Errorf("Mallory is a bad person, and should not be allowed to do anything")
    }
    
    return nil
}
```

One thing not shown in this example but which is very useful for DargoContext scoped service is the
use of the destructor function.  Whenever a context is cancelled all services created for that
context.Context will have their destructor function called, which is a good way to clean up any
resources that the service might have acquired.

## Provider

Rather than injecting an explicit structure it is sometimes useful to inject a Provider.
The benefits of injecting a Provider are:

1.  Lazy creation of the associated service
2.  Getting ALL of the services associated with the name rather than just one
3.  Selecting a particularly qualified service at runtime

You use a provider by making the type of your injection point a Provider, like this:

```go
type RainbowServiceData struct {
	ColorProvider ioc.Provider `inject:"ColorService"`
}
```

The type of ColorProvider is Provider.  When the ColorProvider Get method is used it will return
a service named ColorService.  

## Error Service

The user can supply an implementation of the ioc.ErrorService interface to be notified about certain errors
that happen during the lookup and creation of services.  This is useful for centralized logging
or for other tracing applications.

These are the types of errors that are sent to the ErrorService.  They are:

1.  Service creation failure
2.  Dynamic configuration error
3.  Validation lookup failure

Implementations of ErrorService must be named _ErrorService_ (ioc.ErrorServiceName) in the
namespace _user/services_ (ioc.UserServicesNamespace).  Implementations of ErrorService
**must** be in the Singleton scope.  Implementations of ErrorService will be created by
the system as soon as they are bound into the ServiceLocator.  Any failure during creation
of the ErrorService will cause the configuration commit to fail.  Care should be taken
with the services used by an ErrorService since they will also be created as soon as
the ErrorService is bound into the locator.

### Service Creation Errors

When a service fails during creation the ErrorService OnFailure method will be called with:

1.  The type will be _DYNAMIC_CONFIGURATION_FAILURE_ (ioc.ServiceCreationFailure)
2.  The error that occurred (possibly wrapped in a MultiError)
3.  The descriptor of the service that failed during creation
4.  The injectee struct into which this service was to be injected if appropriate
5.  A nil injectee descriptor

### Dynamic Configuration Error

When a dynamic configuration of the locator fails the ErrorService OnFailure method will be
called with:

1.  The type will be _SERVICE_CREATION_FAILURE_ (ioc.DynamicConfigurationFailure)
2.  The error that occurred (possibly wrapped in a MultiError)
3.  A nil descriptor
4.  A nil injectee
5.  A nil injectee descriptor

### Validation Lookupg Error

1.  The type will be _LOOKUP_VALIDATION_FAILURE_ (ioc.LookupValidationFailure)
2.  The error that occurred (possibly wrapped in a MultiError)
3.  The descriptor that failed validation
4.  A nil injectee
5.  The descriptor of the parent of the service to be injected, or nil if this is a direct lookup

### Error Service Example

This is an example of an ErrorService that logs the error with fields from the information
passed to the OnFailure method.  Not all the code in the example is in the README, please see
the examples/error_service_example.go for the rest of the code.

Here is an implementation of the ErrorService:

```go
type ErrorService struct {
	Logger *logrus.Logger `inject:"Logger"`
}

func (es *ErrorService) OnFailure(info ioc.ErrorInformation) error {
	es.Logger.WithField("FailureType", info.GetType()).
		WithField("ErrorString", info.GetAssociatedError().Error()).
		WithField("ErrorInjectee", info.GetInjectee()).
		Errorf("Descriptor %v failed", info.GetDescriptor())
	return nil
}
```

This is how to bind this service (along with the other services):

```go
locator, err := ioc.CreateAndBind("ErrorServiceExample", func(binder ioc.Binder) error {
		binder.BindWithCreator("Logger", loggerServiceCreator)
		binder.BindWithCreator("WonkyService", wonkyServiceCreator)
		binder.Bind(ioc.ErrorServiceName, ErrorService{}).InNamespace(ioc.UserServicesNamespace)
```

WonkyService always returns an error in its creation method.  When it does, the error service
is called, creating a log that looks something like this:

```
time="2018-09-08T13:49:55-04:00"
level=error
msg="Descriptor default#WonkyService.5.3 failed"
ErrorInjectee="<nil>"
ErrorString="wonky service error"
FailureType=SERVICE_CREATION_FAILURE
```

## Validation Service

The user can supply an implementation of the ioc.ValidationService interface which will be used
when the user attempts to do the following actions in Dargo:

1.  Bind a service into the ServiceLocator
2.  Unbind a service from the ServiceLocator
3.  Lookup a service directly from the ServiceLocator
4.  Inject a service into another service

Implementations of ValidationService must be named _ValidationService_ (ioc.ValidationServiceName) in the
namespace _user/services_ (ioc.UserServicesNamespace).  Implementations of ValidationService
**must** be in the Singleton scope.  Implementations of ValidationService will be created by
the system as soon as they are bound into the ServiceLocator.  Any failure during creation
of the ValidationService will cause the configuration commit to fail.  Care should be taken
with the services used by an ValidationService since they will also be created as soon as
the ValidationService is bound into the locator.


