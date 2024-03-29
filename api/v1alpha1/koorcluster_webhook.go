/*
Copyright 2023 Koor Technologies, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/robfig/cron/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var koorclusterlog = logf.Log.WithName("koorcluster-resource")

func (r *KoorCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-storage-koor-tech-v1alpha1-koorcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=storage.koor.tech,resources=koorclusters,verbs=create;update,versions=v1alpha1,name=mkoorcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KoorCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KoorCluster) Default() {
	koorclusterlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-storage-koor-tech-v1alpha1-koorcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=storage.koor.tech,resources=koorclusters,verbs=create;update,versions=v1alpha1,name=vkoorcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KoorCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KoorCluster) ValidateCreate() (admission.Warnings, error) {
	koorclusterlog.Info("validate create", "name", r.Name)

	return r.validateKoorCluster()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KoorCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	koorclusterlog.Info("validate update", "name", r.Name)

	return r.validateKoorCluster()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KoorCluster) ValidateDelete() (admission.Warnings, error) {
	koorclusterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *KoorCluster) validateKoorCluster() (admission.Warnings, error) {
	var allErrs field.ErrorList
	if err := r.validateUpgradeSchedule(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: "storage.koor.tech", Kind: "KoorCluster"},
		r.Name, allErrs)
}

func (r *KoorCluster) validateUpgradeSchedule() *field.Error {
	if !r.Spec.UpgradeOptions.IsEnabled() {
		return nil
	}

	schedule := r.Spec.UpgradeOptions.Schedule
	if _, err := cron.ParseStandard(schedule); err != nil {
		return field.Invalid(field.NewPath("spec").Child("upgradeOptions").Child("schedule"), schedule, err.Error())
	}
	return nil
}
