package controller

import (
	"github.com/symcn/api"
	corev1 "k8s.io/api/core/v1"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func filterHpaEvent() api.Predicate {
	return &filterHpaEventHandler{}
}

type filterHpaEventHandler struct{}

func (f *filterHpaEventHandler) Create(obj rtclient.Object) bool {
	return filterHpaKind(obj)
}

func (f *filterHpaEventHandler) Update(oldObj, newObj rtclient.Object) bool {
	return filterHpaKind(newObj)
}

func (f *filterHpaEventHandler) Delete(obj rtclient.Object) bool {
	return filterHpaKind(obj)
}

func (f *filterHpaEventHandler) Generic(obj rtclient.Object) bool {
	return filterHpaKind(obj)
}

func filterHpaKind(obj rtclient.Object) bool {
	e, ok := obj.(*corev1.Event)
	if !ok {
		return false
	}
	if e.InvolvedObject.Kind != "HorizontalPodAutoscaler" {
		return false
	}
	if e.Type != corev1.EventTypeNormal {
		return false
	}
	return true
}
