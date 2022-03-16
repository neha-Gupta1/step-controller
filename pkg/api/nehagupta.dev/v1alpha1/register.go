package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   "nehagupta.dev",
	Version: "v1alpha1",
}

var (
	SchemeBuilder runtime.SchemeBuilder
	AddToScheme   = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(addKnownType)
}

func addKnownType(scheme *runtime.Scheme) (err error) {
	scheme.AddKnownTypes(SchemeGroupVersion, &Step{}, &StepList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return err
}

func Resource(resource string) (qualifiedResource schema.GroupResource) {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
