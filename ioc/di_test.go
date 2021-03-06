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
	"strings"
	"testing"
)

const (
	DILocator1 = "DITestLocator1"
	DILocator2 = "DITestLocator2"
	DILocator3 = "DITestLocator3"
	DILocator4 = "DITestLocator4"
	DILocator5 = "DITestLocator5"
	DILocator6 = "DITestLocator6"
	DILocator7 = "DITestLocator7"
	DILocator8 = "DITestLocator8"
	DILocator9 = "DITestLocator9"

	AServiceName = "A"
	BServiceName = "B"

	BNamespaceName2 = "some/user/namespace"
	BServiceName2   = "BService"
	BRed            = "Red"
	BBlue           = "Blue"
	BGreen          = "Green"

	CServiceName = "C"
	DServiceName = "D"
	EServiceName = "E"

	RainbowName      = "RainbowService"
	ColorServiceName = "ColorService"

	ExpectedPanicMessage = "expected panic message"
	PanicService         = "PanicService"
)

func TestInitializerSuccess(t *testing.T) {
	locator, err := CreateAndBind(DILocator1, func(binder Binder) error {
		binder.Bind(AServiceName, ASimpleService{})
		binder.Bind(BServiceName, BSimpleService{})

		return nil
	})
	if err != nil {
		t.Error("", err)
		return
	}

	aRaw, err := locator.GetDService(AServiceName)
	if err != nil {
		t.Error("", err)
		return
	}

	a, ok := aRaw.(*ASimpleService)
	if !ok {
		assert.True(t, ok, "invalid type")
		return
	}

	assert.True(t, a.initialized, "initializer not called")
}

func TestComplexInjectionName(t *testing.T) {
	locator, err := CreateAndBind(DILocator2, func(binder Binder) error {
		binder.Bind(CServiceName, CSimpleService{})
		binder.Bind(BServiceName2, BSimpleService{}).InNamespace(BNamespaceName2).
			QualifiedBy(BBlue).QualifiedBy(BRed).QualifiedBy(BGreen)

		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator2) {
		return
	}

	cRaw, err := locator.GetDService(CServiceName)
	if !assert.Nil(t, err, "couldn't create CService") {
		fmt.Println("", err)
		return
	}

	c, ok := cRaw.(*CSimpleService)
	if !assert.True(t, ok, "Invalid type for CService") {
		return
	}

	assert.True(t, c.initialized, "initializer not called")
}

func TestComplexNoNamespaceInjectionName(t *testing.T) {
	locator, err := CreateAndBind(DILocator3, func(binder Binder) error {
		binder.Bind(DServiceName, DSimpleService{})
		binder.Bind(BServiceName, BSimpleService{}).
			QualifiedBy(BBlue).QualifiedBy(BRed).QualifiedBy(BGreen)

		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator3) {
		return
	}

	dRaw, err := locator.GetDService(DServiceName)
	if !assert.Nil(t, err, "couldn't create DService") {
		fmt.Println("", err)
		return
	}

	d, ok := dRaw.(*DSimpleService)
	if !assert.True(t, ok, "Invalid type for DService") {
		return
	}

	assert.True(t, d.B.initialized, "initializer not called")
}

func TestProviderInjection(t *testing.T) {
	locator, err := CreateAndBind(DILocator4, func(binder Binder) error {
		binder.Bind(EServiceName, ESimpleService{})
		binder.Bind(BServiceName, BSimpleService{})

		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator4) {
		return
	}

	eRaw, err := locator.GetDService(EServiceName)
	if !assert.Nil(t, err, "couldn't create DService") {
		fmt.Println("", err)
		return
	}

	e, ok := eRaw.(*ESimpleService)
	if !assert.True(t, ok, "Invalid type for DService") {
		return
	}

	assert.True(t, e.initialized, "initializer not called")

	bServiceRaw, err := e.BProvider.Get()
	if !assert.Nil(t, err, "provider should not be nil") {
		return
	}

	bService, ok := bServiceRaw.(*BSimpleService)
	if !assert.True(t, ok, "invalid type for bServiceRaw") {
		return
	}

	assert.True(t, bService.initialized, "BService was not initialized")
}

func TestProviderGetAll(t *testing.T) {
	locator, err := CreateAndBind(DILocator6, func(binder Binder) error {
		binder.Bind(RainbowName, RainbowServiceData{})

		binder.Bind(ColorServiceName, colorServiceData{}).QualifiedBy(BRed)
		binder.Bind(ColorServiceName, colorServiceData{}).QualifiedBy(BBlue)
		binder.Bind(ColorServiceName, colorServiceData{}).QualifiedBy(BGreen)

		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator6) {
		return
	}

	rainbowRaw, err := locator.GetDService(RainbowName)
	if !assert.Nil(t, err, "couldn't create RainbowService") {
		fmt.Println("", err)
		return
	}

	rainbow, ok := rainbowRaw.(*RainbowServiceData)
	if !assert.True(t, ok, "Invalid type for RainbowService") {
		return
	}

	checkColors := []string{BRed, BBlue, BGreen}

	allColorServicesRaw, err := rainbow.ColorProvider.GetAll()
	if !assert.Nil(t, err, "Could not get all color services") {
		return
	}

	assert.Equal(t, 3, len(allColorServicesRaw), "unexpected number of services")

	for index, cc := range checkColors {
		colorServiceRaw := allColorServicesRaw[index]

		colorService, ok := colorServiceRaw.(ColorService)
		if !assert.True(t, ok, "service from GetAll array has incorrect type") {
			return
		}

		if !assert.Equal(t, cc, colorService.GetColor(), "invalid color at index %d", index) {
			// return
		}
	}
}

func TestProviderQualifiedBy(t *testing.T) {
	locator, err := CreateAndBind(DILocator5, func(binder Binder) error {
		binder.Bind(RainbowName, RainbowServiceData{})

		binder.Bind(ColorServiceName, colorServiceData{}).QualifiedBy(BRed)
		binder.Bind(ColorServiceName, colorServiceData{}).QualifiedBy(BBlue)
		binder.Bind(ColorServiceName, colorServiceData{}).QualifiedBy(BGreen)

		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator5) {
		return
	}

	rainbowRaw, err := locator.GetDService(RainbowName)
	if !assert.Nil(t, err, "couldn't create RainbowService") {
		fmt.Println("", err)
		return
	}

	rainbow, ok := rainbowRaw.(*RainbowServiceData)
	if !assert.True(t, ok, "Invalid type for RainbowService") {
		return
	}

	checkColors := []string{BGreen, BRed, BBlue}

	for _, cc := range checkColors {
		if !checkColor(t, rainbow, cc) {
			return
		}
	}
}

func TestNoCurrentServiceProvider(t *testing.T) {
	locator, err := CreateAndBind(DILocator8, func(binder Binder) error {
		binder.Bind(RainbowName, RainbowServiceData{})

		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator8) {
		return
	}

	rainbowRaw, err := locator.GetDService(RainbowName)
	if !assert.Nil(t, err, "couldn't create RainbowService") {
		fmt.Println("", err)
		return
	}

	rainbow, ok := rainbowRaw.(*RainbowServiceData)
	if !assert.True(t, ok, "Invalid type for RainbowService") {
		return
	}

	_, err = rainbow.ColorProvider.Get()
	if !assert.NotNil(t, err, "Get should have failed, there is no service in the locator yet") {
		return
	}

	err = BindIntoLocator(locator, func(binder Binder) error {
		binder.Bind(ColorServiceName, colorServiceData{}).QualifiedBy(BRed)

		return nil
	})
	if !assert.Nil(t, err, "error adding color service") {
		return
	}

	raw, err := rainbow.ColorProvider.Get()
	if !assert.Nil(t, err, "service should now be available") {
		return
	}

	_, ok = raw.(ColorService)
	assert.True(t, ok, "Incorrect type")
}

func TestInitializerPanics(t *testing.T) {
	locator, err := CreateAndBind(DILocator7, func(binder Binder) error {
		binder.Bind(PanicService, PanicyInitializerData{})
		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator7) {
		return
	}

	raw, err := locator.GetDService(PanicService)
	if !assert.Nil(t, raw, "should not have returned a service, got %v", raw) {
		return
	}

	if !assert.NotNil(t, err, "Should have gotten a failure") {
		return
	}

	assert.True(t, strings.Contains(err.Error(), ExpectedPanicMessage), "should have gotten panic message")
}

func checkColor(t *testing.T, rainbow *RainbowServiceData, color string) bool {
	raw, err := rainbow.ColorProvider.QualifiedBy(color).Get()
	if !assert.Nil(t, err, "Did not find service %s", color) {
		return false
	}

	colorService, ok := raw.(ColorService)
	if !assert.True(t, ok, "incorrect type for color service %v", colorService) {
		return false
	}

	return assert.Equal(t, color, colorService.GetColor(), "color did not match")
}

func TestOptionalInject(t *testing.T) {
	locator, err := CreateAndBind(DILocator9, func(binder Binder) error {
		binder.Bind("OptionalService", &ServiceWithOptionalAndRequiredInjections{}).InScope(PerLookup)

		// Bind A and C but not B
		binder.Bind("SimpleService", &SimpleService{}).QualifiedBy("A")
		binder.Bind("SimpleService", &SimpleService{}).QualifiedBy("C")
		return nil
	})
	if !assert.Nil(t, err, "couldn't create locator %s", DILocator9) {
		return
	}

	raw, err := locator.GetDService("OptionalService")
	if !assert.Nil(t, err, "couldn't get optional service %v", err) {
		return
	}

	optionalService := raw.(*ServiceWithOptionalAndRequiredInjections)

	assert.NotNil(t, optionalService.SSa)
	assert.Nil(t, optionalService.SSb)
	assert.NotNil(t, optionalService.SSc)

	// Now bind the missing service, make sure optional service do, like, actually show up when there
	err = BindIntoLocator(locator, func(binder Binder) error {
		binder.Bind("SimpleService", &SimpleService{}).QualifiedBy("B")
		return nil
	})

	raw, err = locator.GetDService("OptionalService")
	if !assert.Nil(t, err, "couldn't get optional service %v", err) {
		return
	}

	optionalService = raw.(*ServiceWithOptionalAndRequiredInjections)

	assert.NotNil(t, optionalService.SSa)
	assert.NotNil(t, optionalService.SSb)
	assert.NotNil(t, optionalService.SSc)

	// Now unbind one of the other services and make sure they aren't somehow optional
	err = UnbindServices(locator, DSK("SimpleService", "C"))
	if !assert.Nil(t, err, "couldn't unbind existing service") {
		return
	}

	_, err = locator.GetDService("OptionalService")
	if !assert.NotNil(t, err, "should not have been able to get service", err) {
		return
	}

	assert.True(t, IsServiceNotFound(err), "should have been ServiceNotFound %v", err)

}

type BSimpleService struct {
	initialized bool
}

func (b *BSimpleService) DargoInitialize(Descriptor) error {
	b.initialized = true
	return nil
}

type ASimpleService struct {
	B           *BSimpleService `inject:"B"`
	initialized bool
}

func (a *ASimpleService) DargoInitialize(Descriptor) error {
	if !a.B.initialized {
		return fmt.Errorf("Injected service B MUST have been initialized before this is called")
	}

	a.initialized = true

	return nil
}

type CSimpleService struct {
	B           *BSimpleService `inject:"some/user/namespace#BService@Red@Green"`
	initialized bool
}

func (c *CSimpleService) DargoInitialize(Descriptor) error {
	if !c.B.initialized {
		return fmt.Errorf("Injected service B must have been initialized in CSimpleService")
	}

	c.initialized = true

	return nil
}

type DSimpleService struct {
	B *BSimpleService `inject:"B@Green"`
}

type ESimpleService struct {
	BProvider   Provider `inject:"B"`
	initialized bool
}

func (e *ESimpleService) DargoInitialize(Descriptor) error {
	e.initialized = true

	return nil
}

type RainbowServiceData struct {
	ColorProvider Provider `inject:"ColorService"`
}

type ColorService interface {
	GetColor() string
}

type colorServiceData struct {
	color string
}

func (csd *colorServiceData) DargoInitialize(desc Descriptor) error {
	csd.color = desc.GetQualifiers()[0]
	return nil
}

func (csd *colorServiceData) GetColor() string {
	return csd.color
}

type PanicyInitializerData struct {
}

func (pid *PanicyInitializerData) DargoInitialize(desc Descriptor) error {
	panic(ExpectedPanicMessage)
}

type ServiceWithOptionalAndRequiredInjections struct {
	SSa *SimpleService `inject:"SimpleService@A"`
	SSb *SimpleService `inject:"SimpleService@B,optional"`
	SSc *SimpleService `inject:"SimpleService@C"`
}
