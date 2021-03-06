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

## Dependency Injector

Dargo is an depenency injection system for GO.

Dargo services are scoped and are created and destroyed based on the defined lifecycle of the
scope.  For example services in the Singleton scope are only created once.  Services in the PerLookup
scope are created every time they are injected.

The similarities between this library and the java Dependency Injection framework [hk2](https://javaee.github.io/hk2/)
is intentional.  The plan is to port many of the features of hk2 to this library.

## Table of Contents

1.  [Basic Usage](#basic-usage)
2.  [Testing](#testing)
3.  [Automatic Service Creation](#automatic-service-creation)
4.  [Service Names](#service-names)
5.  [Optional Injection](#optional-injection)
6.  [Immediate Services](#immediate-scope)
7.  [Context Scope](#context-scope)
8.  [Provider](#provider)
9.  [Error Service](#error-service)
10.  [Security](#validation-service)
11.  [Configuration Listener](#configuration-listener)
12.  [Custom Injection](#custom-injection)

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

The entire Dargo API is thread-safe.  You can call Dargo API from within callbacks from the Dargo API.

### Injection Example

In this example a service called SimpleService will inject a logger.

```go
// SimpleServiceData is a struct implementing SimpleService
type SimpleServiceData struct {
	Log *logrus.Logger `inject:"LoggerService_Name"`
}

// SimpleService is a test service
type SimpleService interface {
	// CallMe logs a message to the logger!
	CallMe()
}

// CallMe implements the SimpleService method
func (ss *SimpleServiceData) CallMe() {
	ss.Log.Info("This logger was injected!")
}
```

SimpleServiceData has a field annotated with _inject_ followed by the name of the service to inject.

The logger is a dargo service that is bound with a creation method.  That creation method looks like this:

```go
func newLogger(ioc.ServiceLocator, ioc.Descriptor) (interface{}, error) {
	return logrus.New(), nil
}
```

The binding of SimpleServiceData will provide the struct to create.  The binding of the Logger
will provide the creation function.  Both the logger service and the SimpleService are bound into a
ServiceLocator.  This is normally done near the start of your program:

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

## Automatic Service Creation

A service that is bound with the Bind method provides an instance of the struct to create.  The struct
passed in is NOT the one that will be created and injected into, it is only used to determine the items
that need to be injected and the type to create.  If that structure implements DargoInitializer (see below)
then the DargoInitialize method will be called after all the service dependencies have been injected.  This
provides an opportunity to do other initialization to the structure, or to return an error should there be
some issue that can't be resolved.

```go
// DargoInitializer is used when using Binder.Bind and need
// to be able to do further initialization after the services have
// been injected into the structure and before it is given to the injectee
// or lookup user
type DargoInitializer interface {
	// DargoInitialize is a method that will be called after all the
	// injected fields have been filled in.  If this method returns
	// a non-nil error then the creation of the service will fail
	// The descriptor passed in is the descriptor being used to create
	// the service
	DargoInitialize(Descriptor) error
}
```

A service that is bound with a Creator function expects the entire initialization of that service to be
done by the Creator function.  Even if that service implements DargoInitializer it will **not** have the
DargoInitialize method called on it by the system.  The Creator function is passed the ServiceLocator that
was used to create the service and the descriptor representing the description of the service.

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

You can also use complex names in the injection description of structures.  The general format is:

```
[namespace#]name[@qualifier]*[,directive]*
```

A valid example of an injection description could be something like this:

```go
my/user/namespace#LoggerService@Red@Green,optional
```

In the above example the namespace is _my/user/namespace_, the name is _LoggerService_, there are two qualifiers
_Red_ and _Green_ and one directive, _optional_.  There is currently only one legal directive, which is "optional".

Only the name part is required.  For example, if you wanted to inject a service
named ColorService in the visible/light namespace with qualifier Green, you would do
something like this:

```go
type Service struct {
	Green ColorService `inject:"visible/light#ColorService@Green"`
}
```

## Optional Injection

Sometimes it is not certain whether an injection point will be satisfyable at the time a service
is created.  For cases like this optional injection may be appropriate.  An injection point may specify
that the injection is optional by adding the __optional__ directive to the injection string.  When an
injection point is optional and no matching service is found it will not cause an error and instead will
not inject anything into the field.  The following structure has two required injection points and one
optional injection point:

```go
type ServiceWithOptionalAndRequiredInjections struct {
	Foo *Foo `inject:"Foo"`
	Bar *Bar `inject:"Bar,optional"`
	Baz *Baz `inject:"Baz"`
}
```

The fields **Foo** and **Baz** are required, but the field **Bar** can either be available or not.  When a required
injection point cannot be satisified it will cause an error, but when an optional injection point
cannot be satisfied it will simply be left alone and the structure can still be created normally.

## Immediate Scope

Services bound into _ImmediateScope_ (ioc.ImmediateScope) will be started immediately.  These are not
lazy services, but instead services that will be started as soon as the system sees that they have been
bound into the system.  The Immediate scope is enabled by calling the method ioc.EnableImmediateScope.
Services in the ImmediateScope will only be started once the immediate scope has been enabled.

When service descriptions in the ImmediateScope are unbound from the ServiceLocator the services corresponding
to the unbound descriptor will be destroyed.  If a service in the ImmediateScope fails during creation the
[ErrorService](#error-service) can be used to catch the error and do remediation.

Care should be taken with the services injected into an Immediate service, since they will also become
immediate.  Instead consider injecting [Providers](#provider) into immediate scoped services which enable
those injected services to remain lazy.

### Immediate Scope Example

In this example there is a service that prints out a "Hello, World!" banner without having to be
explicitly looked up.  To do so it prints the banner in it's DargoInitialize method:

```go
var immediateExampleStarted = false

type IShoutImmediately struct{}

func (shouter *IShoutImmediately) DargoInitialize(ioc.Descriptor) error {
	fmt.Println("*********************************")
	fmt.Println("*                               *")
	fmt.Println("*       Hello, World            *")
	fmt.Println("*                               *")
	fmt.Println("*********************************")

	immediateExampleStarted = true

	return nil
}
```

This service is bound into the locator at the start.  However, it will not be run until the ImmediateContext
has been enabled.  Here is the binding of the service:

```go
locator, _ := ioc.CreateAndBind("ImmediateLocator", func(binder ioc.Binder) error {
    binder.Bind("Shouter", &IShoutImmediately{}).InScope(ioc.ImmediateScope)
    return nil
})
```

Since the immediate scope isn't there yet, the service will not be started yet.  You must call this:

```go
ioc.EnableImmediateScope(locator)
```

At that point all services in the immediate scope will be started.  Further, any services bound into
the locator AFTER EnableImmediateScope is enabled will also be started immediately.

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

### Security (Validation) Service Example

This example shows how to create services that can only be injected into services in a special namespace
_user/protected_.  Further, once the initial set of services in the special
namespace have been created that namespace will be locked down so that no-one can later add services into it.

Any service qualified with _Secure_ will not be able to be directly looked up, and
can only be injected into services in the _user/protected_ namespace because the
Validator checks these conditions and disallows the operation otherwise.  Here is the implementation of the
Validator (returned from the ValidationService):

```go
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
```

The following code shows the binding of the validation service and a service in the
_user/protected_ namespace along with some other services.  The interesting thing to note about
it is that since the validation service is bound at the same time as the service in
the protected namespace the validator is NOT run, and therefor the service is allowed in.
However, after that the Validator is run for all Bind/Unbind operations.

```go
    ioc.CreateAndBind("SecurityExampleLocator", func(binder ioc.Binder) error {
	    // The validator is not run against the services bound in this binder
		binder.Bind(ioc.ValidationServiceName, secureValidationService{}).InNamespace(ioc.UserServicesNamespace)
		
		// This service is marked "Secret" and can not be looked up or injected into a normal service
		binder.Bind("SuperSecretService", SuperSecretService{}).QualifiedBy(SecureQualifier)
		
		// This service is in the protected namespace and therefore CAN have Secure services injected
		binder.Bind("SystemProtectedService", ServiceData{}).InNamespace(ProtectedNamespace)
		
		// This is a normal user service, which should NOT be able to inject Secure services
		binder.Bind("NormalUserService", ServiceData{})

		return nil
	})
```

The rest of the security example is found in the examples/security_example.go file.  It is an exercise left to
the reader to modify the implementation of the Validator to also disallow people from Unbinding the
ValidationService itself, since if someone could do that they could disable the security checks!

## Configuration Listener

A user may register an implementation of ConfigurationListener to be notified whenever the set of
Descriptors in a ServiceLocator has changed. The ConfigurationListener must be in the Singleton scope,
be named _ConfigurationListener_ (ioc.ConfigurationListenerName) and be in the _user/services_
(ioc.UserServicesNamespace) namespace.

## Custom Injection

Dargo allows users to choose their own injection scheme.  The default scheme
provided by the system uses the `inject:"whatever"` annotation on structures.
However, Dargo allows the use of any other injection scheme.  For example,
instead of having tag _inject_ mean something instead you could use
_alternate_, or you can have some external file providing information about
injection points.

### Custom Injection Example

The following example does magic injection by simply using the name of the struct or
interface as the name of the service to inject.  If the example resolver finds a service in
the default namespace with that name it uses it.  Otherwise it simply returns false
so other resolvers can take a look.

[embedmd]:# (examples/resolution/custom_resolution_example.go /^package.*/ $)
```go
package resolution

import (
	"fmt"
	"github.com/jwells131313/dargo/ioc"
	"reflect"
)

// AutomaticResolver will resolve any field of a struct that has
// a type with a service with a name that equals that type
type AutomaticResolver struct {
}

// Resolve looks at the type of the field and if it is a pointer or an interface
// gets the simple name of that type and uses that as the name of the service
// to look up (in the default namespace).  Doing this creates a "magic" injector
// that works even without use of annotations in the structure being injected into
func (ar *AutomaticResolver) Resolve(locator ioc.ServiceLocator, injectee ioc.Injectee) (*reflect.Value, bool, error) {
	field := injectee.GetField()
	typ := field.Type

	var name string
	switch typ.Kind() {
	case reflect.Ptr:
		itype := typ.Elem()

		name = itype.Name()
		break
	case reflect.Interface:
		name = typ.Name()
		break
	default:
		return nil, false, nil
	}

	if name == "" {
		return nil, false, nil
	}

	svc, err := locator.GetDService(name)
	if err != nil {
		return nil, false, nil
	}

	rVal := reflect.ValueOf(svc)

	return &rVal, true, nil
}

// BService is a service injected into AService that just prints Hello, World
type BService struct {
}

func (b *BService) run() {
	fmt.Println("Hello, World")
}

// AService is injected using the custom injector
type AService struct {
	// BService will be magically injected, even without an indicator on the struct
	BService *BService
}

// CustomResolution is a method that will create a locator, binding in the
// custom resolver and the A and B Services.  It will then get the AService
// and use the injected BService.  This example shows how a custom resolver
// can use whatever resources it has available to choose injection points
// in a service
func CustomResolution() error {
	locator, err := ioc.CreateAndBind("AutomaticResolverLocator", func(binder ioc.Binder) error {
		binder.Bind(ioc.InjectionResolverName, &AutomaticResolver{}).InNamespace(ioc.UserServicesNamespace)
		binder.Bind("AService", &AService{})
		binder.Bind("BService", &BService{})

		return nil
	})
	if err != nil {
		return err
	}

	aServiceRaw, err := locator.GetDService("AService")
	if err != nil {
		return err
	}

	aService := aServiceRaw.(*AService)

	aService.BService.run()

	return nil
}
```
