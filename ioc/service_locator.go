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
	"github.com/jwells131313/goethe"
	"github.com/pkg/errors"
	"reflect"
	"sort"
)

// ServiceLocator The main registry for dargo.  Use it to get context sensitive lookups
// for application services
type ServiceLocator interface {
	fmt.Stringer

	// GetService gets the service that is correct for the current context with the given key.
	// It returns the best implementation of the interface
	GetService(toMe ServiceKey) (interface{}, error)

	// GetDService gets the service with the given name in the default namespace and
	// with the provided qualifiers
	// It returns the best implementation of the interface
	GetDService(name string, qualifiers ...string) (interface{}, error)

	// GetAllServices returns all the services matching the service key
	GetAllServices(toMe ServiceKey) ([]interface{}, error)

	// GetService gets the service that is correct for the current context with the given
	// descriptor and any other error if there was an error creating the interface
	GetServiceFromDescriptor(desc Descriptor) (interface{}, error)

	// CreateServiceFromDescriptor creates the service without using contextual data,
	// simply properly invoking the creation method
	CreateServiceFromDescriptor(desc Descriptor) (interface{}, error)

	// GetDescriptors Returns all descriptors that return true when passed through the input function
	// will not return nil, but may return an empty list
	GetDescriptors(Filter) ([]Descriptor, error)

	// GetBestDescriptor returns the best descriptor found returning true through the input function
	// The best descriptor is the one with the highest rank, or if rank is equal the one with the
	// lowest serviceId or if the serviceId are the same the one with the highest locatorId
	GetBestDescriptor(Filter) (Descriptor, error)

	// Inject takes a pointer to a structure and injects into it any
	// injection points from the services in this ServiceLocator.  The struct
	// passed in will not be managed by this ServiceLocator after this method
	// returns.  If an injection point cannot be resolved by this ServiceLocator
	// an error is returned.  If the input is not a pointer to a structure this
	// will return an error.  The locators error handlers will not be run in error
	// cases but validators will be run
	Inject(interface{}) error

	// GetName gets the name of this ServiceLocator
	GetName() string

	// GetID Gets the id of this ServiceLocator
	GetID() int64

	// Will shut down all services associated with this ServiceLocator
	Shutdown()

	// GetState Returns LocatorStateRunning or LocatorStateShutdown depending on if this
	// locator is currently running or it has been shut down
	GetState() string
}

// Provider is used as an injection point in a service for a few different reasons
// 1.  A service might be injected with Provider if there is a desire for the service
//     to be instantiated lazily.  This is useful for services that are expensive to
//     initialize and would benefit from late initialization
// 2.  A service might be interested in all of the registered implementations with
//     a certain name (usually all implementing the same interface).  The GetAll
//     method can be used to iterate over all of the services registered with the
//     same name
// 3.  A Provider can be used to select qualified services at run-time
//     rather than at compile time using the QualifiedBy method
type Provider interface {
	// Get gets the best implementation of the service associated with
	// this injected service
	Get() (interface{}, error)

	// GetAll gets all the implementions of the service associated with
	// this injected service
	GetAll() ([]interface{}, error)

	// QualifiedBy allows the user to select a particularly qualified
	// service at runtime rather than make the selection at compile
	// time.  This returns a further specified Provider for which the
	// Get or GetAll methods will include the qualifier or qualifiers provided
	QualifiedBy(string) Provider
}

type serviceLocatorData struct {
	glock              goethe.Lock
	name               string
	ID                 int64
	allDescriptors     []Descriptor
	nextServiceID      int64
	perLookupContext   ContextualScope
	singletonContext   ContextualScope
	generation         uint64
	state              string
	errorServices      []ErrorService
	validationServices []ValidationService
}

// NewServiceLocator this will find or create a service locator with the given name, and
// return errors based on the value of qos
func NewServiceLocator(name string, qos int) (ServiceLocator, error) {
	locatorsLock.Lock()
	defer locatorsLock.Unlock()

	err := checkNameCharacters(name)
	if err != nil {
		return nil, err
	}

	if qos != FailIfPresent && qos != FailIfNotPresent && qos != ReturnExistingOrCreateNew {
		return nil, fmt.Errorf("Unkonwn quality of service %d", qos)
	}

	retVal, found := locators[name]
	if found {
		if qos == FailIfPresent {
			return nil, fmt.Errorf("Quality of service is FailIfPresent and there is a locator with name %s", name)
		}

		return retVal, nil
	}

	// Not found
	if qos == FailIfNotPresent {
		return nil, fmt.Errorf("Quality of service is FailIfNotPresent and there is no locator named %s", name)
	}

	ID := currentID
	currentID = currentID + 1

	retVal = &serviceLocatorData{
		glock:              threadManager.NewGoetheLock(),
		name:               name,
		ID:                 ID,
		allDescriptors:     make([]Descriptor, 0),
		perLookupContext:   newPerLookupContext(),
		state:              LocatorStateRunning,
		errorServices:      make([]ErrorService, 0),
		validationServices: make([]ValidationService, 0),
	}

	retVal.singletonContext, err = newSingletonScope(retVal)
	if err != nil {
		return nil, err
	}

	serviceLocatorDescriptor := NewConstantDescriptor(SSK(ServiceLocatorName), retVal)
	serviceLocatorSystemDescriptor, err := NewDescriptor(serviceLocatorDescriptor, 0, ID)
	if err != nil {
		return nil, err
	}

	dcs := newDynamicConfigurationService(retVal)
	dynamicConfigurationDescriptor := NewConstantDescriptor(SSK(DynamicConfigurationServiceName), dcs)
	dcsSystemDescriptor, err := NewDescriptor(dynamicConfigurationDescriptor, 1, ID)
	if err != nil {
		return nil, err
	}

	retVal.allDescriptors = append(retVal.allDescriptors, serviceLocatorSystemDescriptor)
	retVal.allDescriptors = append(retVal.allDescriptors, dcsSystemDescriptor)

	retVal.nextServiceID = 2

	locators[name] = retVal

	return retVal, nil
}

func (locator *serviceLocatorData) checkState() error {
	if locator.state != LocatorStateRunning {
		return ErrLocatorIsShutdown
	}

	return nil
}

func (locator *serviceLocatorData) GetService(toMe ServiceKey) (interface{}, error) {
	return locator.getServiceFor(toMe, nil)
}

func (locator *serviceLocatorData) getServiceFor(toMe ServiceKey, forMe Descriptor) (interface{}, error) {
	err := locator.checkState()
	if err != nil {
		return nil, err
	}

	f := NewServiceKeyFilter(toMe)

	desc, err := locator.getBestDescriptorFor(f, forMe)
	if err != nil {
		return nil, err
	}

	if desc == nil {
		return nil, fmt.Errorf(ServiceWithNameNotFoundExceptionString, toMe)
	}

	return locator.createService(desc)
}

func (locator *serviceLocatorData) GetDService(name string, qualifiers ...string) (interface{}, error) {
	err := locator.checkState()
	if err != nil {
		return nil, err
	}

	return locator.GetService(DSK(name, qualifiers...))
}

func (locator *serviceLocatorData) GetAllServices(toMe ServiceKey) ([]interface{}, error) {
	return locator.getAllServicesFor(toMe, nil)
}

func (locator *serviceLocatorData) getAllServicesFor(toMe ServiceKey, forMe Descriptor) ([]interface{}, error) {
	err := locator.checkState()
	if err != nil {
		return nil, err
	}

	f := NewServiceKeyFilter(toMe)

	descs, err := locator.getDescriptorsFor(f, forMe)
	if err != nil {
		return nil, err
	}

	retVal := make([]interface{}, 0)
	retErr := NewMultiError()

	for _, desc := range descs {
		us, err := locator.createService(desc)
		if err != nil {
			retErr.AddError(err)
		} else {
			retVal = append(retVal, us)
		}
	}

	return retVal, retErr.GetFinalError()
}

func (locator *serviceLocatorData) GetServiceFromDescriptor(desc Descriptor) (interface{}, error) {
	err := locator.checkState()
	if err != nil {
		return nil, err
	}

	return locator.createService(desc)
}

func (locator *serviceLocatorData) GetDescriptors(filter Filter) ([]Descriptor, error) {
	return locator.getDescriptorsFor(filter, nil)
}

func (locator *serviceLocatorData) getDescriptorsFor(filter Filter, forMe Descriptor) ([]Descriptor, error) {
	err := locator.checkState()
	if err != nil {
		return nil, err
	}

	tid := threadManager.GetThreadID()
	if tid < 0 {
		c := make(chan *igsRet)

		threadManager.Go(locator.channelGetDescriptors, filter, forMe, c)

		ret := <-c

		return ret.descriptors, ret.err
	}

	return locator.internalGetDescriptors(filter, forMe)
}

func (locator *serviceLocatorData) GetBestDescriptor(filter Filter) (Descriptor, error) {
	return locator.getBestDescriptorFor(filter, nil)
}

func (locator *serviceLocatorData) getBestDescriptorFor(filter Filter, forMe Descriptor) (Descriptor, error) {
	err := locator.checkState()
	if err != nil {
		return nil, err
	}

	all, err := locator.getDescriptorsFor(filter, forMe)
	if err != nil {
		return nil, err
	}

	if len(all) == 0 {
		return nil, nil
	}

	return all[0], nil
}

func (locator *serviceLocatorData) GetName() string {
	return locator.name
}

func (locator *serviceLocatorData) GetID() int64 {
	return locator.ID
}

func (locator *serviceLocatorData) Inject(input interface{}) error {
	ty := reflect.TypeOf(input)
	if ty.Kind() != reflect.Ptr {
		return fmt.Errorf("input to inject must be a pointer")
	}
	val := reflect.ValueOf(input)

	iType := ty.Elem()
	if iType.Kind() != reflect.Struct {
		return fmt.Errorf("input to inject must be a pointer to a struct")
	}

	_, err := createAndInject(locator, nil, iType, &val)

	return err
}

func (locator *serviceLocatorData) Shutdown() {
	c := make(chan bool)

	threadManager.Go(func() {
		locator.singletonContext.Shutdown(locator)

		locator.state = LocatorStateShutdown

		c <- true
	})

	<-c
}

func (locator *serviceLocatorData) createService(desc Descriptor) (interface{}, error) {
	scope := desc.GetScope()

	var cs ContextualScope
	if scope == PerLookup {
		cs = locator.perLookupContext
	} else if scope == Singleton {
		cs = locator.singletonContext
	} else {
		csk := CSK(scope)
		raw, err := locator.GetService(csk)
		if err != nil {
			return nil, errors.Wrapf(err, "could not get context %s for service %s", scope, desc)
		}

		var ok bool
		cs, ok = raw.(ContextualScope)
		if !ok {
			return nil, fmt.Errorf("implementation of %s was not a ContextualScope", scope)
		}
	}

	if cs == nil {
		return nil, fmt.Errorf("Could not find scope named %s in catchall", scope)
	}

	userService, err := cs.FindOrCreate(locator, desc)
	if err != nil {
		return nil, err
	}

	return userService, nil
}

type igsRet struct {
	descriptors []Descriptor
	err         error
}

func (locator *serviceLocatorData) channelGetDescriptors(filter Filter, forMe Descriptor,
	retChan chan *igsRet) {
	descs, err := locator.internalGetDescriptors(filter, forMe)

	retVal := &igsRet{
		descriptors: descs,
		err:         err,
	}

	retChan <- retVal
}

// TODO: This will one day need to keep caches
func (locator *serviceLocatorData) internalGetDescriptors(filter Filter, forMe Descriptor) ([]Descriptor, error) {
	locator.glock.ReadLock()
	defer locator.glock.ReadUnlock()

	retVal := make([]Descriptor, 0)
	for _, desc := range locator.allDescriptors {
		if filter.Filter(desc) {
			passedValidation := true

			vi := newValidationInformation(LookupOperation, desc, forMe, filter)

			for _, validationService := range locator.validationServices {
				errRet := &errorReturn{}

				validationFilter := safeGetFilter(validationService, errRet)
				if errRet.err != nil {
					return nil, errRet.err
				}

				if validationFilter.Filter(desc) {
					errRet := &errorReturn{}

					validator := safeGetValidator(validationService, errRet)
					if errRet.err != nil {
						return nil, errRet.err
					}

					safeValidate(validator, vi, errRet)
					valError := errRet.err
					if valError != nil {
						_, ok := valError.(MultiError)
						if !ok {
							valError = NewMultiError(valError)
						}

						locator.runErrorHandlers(LookupValidationFailure, desc, nil, forMe, valError)

						passedValidation = false
					}
				}

			}

			if passedValidation {
				retVal = append(retVal, desc)
			}
		}
	}

	sort.Slice(retVal, func(i, j int) bool {
		if retVal[i].GetRank() > retVal[j].GetRank() {
			return true
		} else if retVal[i].GetRank() < retVal[j].GetRank() {
			return false
		}

		if retVal[i].GetLocatorID() > retVal[j].GetLocatorID() {
			return true
		} else if retVal[i].GetLocatorID() < retVal[j].GetLocatorID() {
			return false
		}

		if retVal[i].GetServiceID() < retVal[j].GetServiceID() {
			return true
		}

		return false
	})

	return retVal, nil
}

func (locator *serviceLocatorData) getGeneration() uint64 {
	tid := threadManager.GetThreadID()
	if tid < 0 {
		c := make(chan uint64)

		threadManager.Go(func(ret chan uint64) {
			locator.glock.ReadLock()
			defer locator.glock.ReadUnlock()

			ret <- locator.generation
		}, c)

		return <-c
	}

	locator.glock.ReadLock()
	defer locator.glock.ReadUnlock()

	return locator.generation
}

func (locator *serviceLocatorData) getNextServiceID() int64 {
	tid := threadManager.GetThreadID()
	if tid < 0 {
		c := make(chan int64)

		threadManager.Go(func(ret chan int64) {
			locator.glock.WriteLock()
			defer locator.glock.WriteUnlock()

			retVal := locator.nextServiceID

			locator.nextServiceID = locator.nextServiceID + 1

			ret <- retVal
		}, c)

		return <-c
	}

	locator.glock.WriteLock()
	defer locator.glock.WriteUnlock()

	retVal := locator.nextServiceID

	locator.nextServiceID = locator.nextServiceID + 1

	return retVal
}

func (locator *serviceLocatorData) CreateServiceFromDescriptor(desc Descriptor) (interface{}, error) {
	err := locator.checkState()
	if err != nil {
		return nil, err
	}

	cf := desc.GetCreateFunction()

	retVal, err := cf(locator, desc)
	if err != nil {
		var hasRunHandlers bool

		hasRunError, isHasRunError := err.(hasRunErrorHandlersError)
		if isHasRunError {
			hasRunHandlers = hasRunError.GetHasRunErrorHandlers()

			err = hasRunError.GetUnderlyingError()
		} else {
			_, isMulti := err.(MultiError)

			if !isMulti {
				err = NewMultiError(err)
			}
		}

		if !hasRunHandlers {
			locator.runErrorHandlers(ServiceCreationFailure, desc, nil, nil, err)
		}
	}

	return retVal, err
}

// update returns true if the error handlers have already been run
func (locator *serviceLocatorData) update(newDescs []Descriptor,
	removers []Filter, originalGeneration uint64) (bool, error) {
	locator.glock.WriteLock()
	defer locator.glock.WriteUnlock()

	if originalGeneration != locator.generation {
		return false, fmt.Errorf("there was an update to the ServiceLocator after this DynamicConfiguration was created")
	}

	newAllDescs := make([]Descriptor, 0)

	var errorServiceUpdate bool
	var validationServiceUpdate bool

	removedDescriptors := make([]Descriptor, 0)
	for _, myDesc := range locator.allDescriptors {
		removeMe := false

		for _, removeFilter := range removers {
			removeMe = removeMe || removeFilter.Filter(myDesc)
		}

		if !removeMe {
			newAllDescs = append(newAllDescs, myDesc)
		} else {
			errorServiceUpdate = errorServiceUpdate || isErrorService(myDesc)
			validationServiceUpdate = validationServiceUpdate || isValidationService(myDesc)

			removedDescriptors = append(removedDescriptors, myDesc)
		}
	}

	for _, removedDescriptor := range removedDescriptors {
		unbindValidationInformation := newValidationInformation(UnbindOperation,
			removedDescriptor, nil, nil)

		for _, validationService := range locator.validationServices {
			errRet := &errorReturn{}

			validator := safeGetValidator(validationService, errRet)
			if errRet.err != nil {
				return false, errRet.err
			}
			if validator == nil {
				continue
			}

			safeValidate(validator, unbindValidationInformation, errRet)
			err := errRet.err
			if err != nil {
				_, ok := err.(MultiError)
				if !ok {
					err = NewMultiError(err)
				}

				locator.runErrorHandlers(DynamicConfigurationFailure, removedDescriptor,
					nil, nil, err)

				return true, err
			}
		}
	}

	for _, newDesc := range newDescs {
		bindValidationInformation := newValidationInformation(BindOperation, newDesc, nil, nil)

		for _, validationService := range locator.validationServices {
			errRet := &errorReturn{}

			validator := safeGetValidator(validationService, errRet)
			if errRet.err != nil {
				return false, errRet.err
			}
			if validator == nil {
				continue
			}

			safeValidate(validator, bindValidationInformation, errRet)
			err := errRet.err
			if err != nil {
				_, ok := err.(MultiError)
				if !ok {
					err = NewMultiError(err)
				}

				locator.runErrorHandlers(DynamicConfigurationFailure, newDesc, nil, nil, err)

				return true, err
			}
		}

		if isErrorService(newDesc) || isValidationService(newDesc) || isConfigurationListener(newDesc) {
			if Singleton != newDesc.GetScope() {
				return false, fmt.Errorf("implementations of %s must be in the singleton scope",
					newDesc.GetName())
			}

			if isErrorService(newDesc) {
				errorServiceUpdate = true
			}
			if isValidationService(newDesc) {
				validationServiceUpdate = true
			}
		}

		newAllDescs = append(newAllDescs, newDesc)
	}

	var success bool

	// Get old services
	oldDescriptors := locator.allDescriptors
	oldErrorServices := locator.errorServices
	oldValidationServices := locator.validationServices

	locator.allDescriptors = newAllDescs

	defer func() {
		if success {
			locator.generation = locator.generation + 1
			return
		}

		locator.validationServices = oldValidationServices
		locator.errorServices = oldErrorServices
		locator.allDescriptors = oldDescriptors
	}()

	if errorServiceUpdate {
		// Must get all error services again
		errorServiceKey, err := NewServiceKey(UserServicesNamespace, ErrorServiceName)
		if err != nil {
			return false, errors.Wrap(err, "creation of error service key failed")
		}

		raws, err := locator.GetAllServices(errorServiceKey)
		if err != nil {
			return false, errors.Wrap(err, "creation of error services failed")
		}

		newErrorServices := make([]ErrorService, 0)
		for _, errorServiceRaw := range raws {
			errorService, ok := errorServiceRaw.(ErrorService)
			if !ok {
				return false, fmt.Errorf("a service %v with error service key does not implement error service",
					errorServiceRaw)
			}

			newErrorServices = append(newErrorServices, errorService)
		}

		locator.errorServices = newErrorServices
	}

	if validationServiceUpdate {
		// Must get all validation services again
		validationServiceKey, err := NewServiceKey(UserServicesNamespace, ValidationServiceName)
		if err != nil {
			return false, errors.Wrap(err, "creation of validation service key failed")
		}

		raws, err := locator.GetAllServices(validationServiceKey)
		if err != nil {
			return false, errors.Wrap(err, "creation of error services failed")
		}

		newValidationServices := make([]ValidationService, 0)
		for _, validationServiceRaw := range raws {
			validationService, ok := validationServiceRaw.(ValidationService)
			if !ok {
				return false, fmt.Errorf("a service %v with validation service key does not implement error service",
					validationServiceRaw)
			}

			newValidationServices = append(newValidationServices, validationService)
		}

		locator.validationServices = newValidationServices
	}

	success = true

	return false, nil
}

func (locator *serviceLocatorData) runErrorHandlers(typ string, desc Descriptor, injectee reflect.Type, forMe Descriptor, err error) {
	ei := newErrorImformation(typ, desc, injectee, forMe, err)

	for _, errorService := range locator.errorServices {
		safeCallUserErrorService(errorService, ei)
	}
}

func (locator *serviceLocatorData) GetState() string {
	return locator.state
}

func (locator *serviceLocatorData) String() string {
	return fmt.Sprintf("ServiceLocator(%s,%d)", locator.name, locator.ID)
}
