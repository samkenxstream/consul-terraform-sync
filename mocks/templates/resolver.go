// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	hcat "github.com/hashicorp/hcat"
	mock "github.com/stretchr/testify/mock"
)

// Resolver is an autogenerated mock type for the Resolver type
type Resolver struct {
	mock.Mock
}

// Run provides a mock function with given fields: tmpl, w
func (_m *Resolver) Run(tmpl hcat.Templater, w hcat.Watcherer) (hcat.ResolveEvent, error) {
	ret := _m.Called(tmpl, w)

	var r0 hcat.ResolveEvent
	if rf, ok := ret.Get(0).(func(hcat.Templater, hcat.Watcherer) hcat.ResolveEvent); ok {
		r0 = rf(tmpl, w)
	} else {
		r0 = ret.Get(0).(hcat.ResolveEvent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(hcat.Templater, hcat.Watcherer) error); ok {
		r1 = rf(tmpl, w)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
