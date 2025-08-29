// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
)

func (i *IntentReconciler) HandleKubernetesObjects(ctx context.Context, intent *flarev1alpha1.Intent) error {
	condition := meta.FindStatusCondition(intent.Status.Conditions, flarev1alpha1.IntentStatusTypeDeploy).DeepCopy()

	condition.Status = metav1.ConditionTrue
	condition.Reason = "KubernetesObjectsHandled"
	condition.Message = ""

	switch intent.Spec.Workload.Type {
	case flarev1alpha1.IntentWorkloadTypeService:
		if err := i.kubernetesService(ctx, intent); err != nil {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "ServiceCreationFailed"
			condition.Message = err.Error()

			break
		}

		if err := i.kubernetesIngress(ctx, intent); err != nil {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "IngressCreationFailed"
			condition.Message = err.Error()

			break
		}

		if err := i.kubernetesDeployment(ctx, intent); err != nil {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "DeploymentCreationFailed"
			condition.Message = err.Error()

			break
		}
	case flarev1alpha1.IntentWorkloadTypeBatch:
		if err := i.kubernetesJob(ctx, intent); err != nil {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "DeploymentCreationFailed"
			condition.Message = err.Error()

			break
		}
	}

	return UpdateStatusCondition(ctx, i.Client, intent, *condition)
}

//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=create;get;list;watch;update

func (i *IntentReconciler) kubernetesIngress(ctx context.Context, intent *flarev1alpha1.Intent) error {

	var exposed []flarev1alpha1.IntentWorkloadPort
	for _, port := range intent.Spec.Workload.Ports {
		if port.Expose {
			exposed = append(exposed, port)
		}
	}

	if len(exposed) == 0 {
		return nil
	}

	var ingress networkingv1.Ingress
	ingress.Name = intent.Namespace
	ingress.Namespace = intent.Namespace

	_, err := controllerutil.CreateOrUpdate(ctx, i.Client, &ingress, func() error {
		if len(ingress.Spec.Rules) != len(exposed) {
			ingress.Spec.Rules = make([]networkingv1.IngressRule, len(exposed))
		}

		for index, port := range exposed {
			ingress.Spec.Rules[index].Host = port.Domain

			if ingress.Spec.Rules[index].HTTP == nil {
				ingress.Spec.Rules[index].HTTP = &networkingv1.HTTPIngressRuleValue{}
			}

			ingress.Spec.Rules[index].HTTP.Paths = []networkingv1.HTTPIngressPath{
				{
					Path:     "/",
					PathType: ptr.To(networkingv1.PathTypePrefix),
					Backend: networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: intent.Namespace,
							Port: networkingv1.ServiceBackendPort{
								Number: port.Port,
							},
						},
						Resource: nil,
					},
				},
			}
		}

		return controllerutil.SetOwnerReference(intent, &ingress, i.Client.Scheme())
	})
	if err != nil {
		return err
	}

	return nil
}

//+kubebuilder:rbac:groups="batch",resources=jobs,verbs=create;get;list;watch;update

func (i *IntentReconciler) kubernetesJob(ctx context.Context, intent *flarev1alpha1.Intent) error {
	var job v1.Job
	job.Name = intent.Namespace
	job.Namespace = intent.Namespace

	_, err := controllerutil.CreateOrUpdate(ctx, i.Client, &job, func() error {
		job.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"intent": intent.Name,
			},
		}

		job.Spec.BackoffLimit = ptr.To(int32(intent.Spec.Workload.Batch.MaxRetries))
		job.Spec.Parallelism = ptr.To(int32(intent.Spec.Workload.Batch.ParallelTasks))
		job.Spec.ActiveDeadlineSeconds = ptr.To(int64(intent.Spec.Workload.Batch.Timeout.Duration.Seconds()))

		if intent.Spec.Workload.Batch.CompletionPolicy == "All" {
			job.Spec.Completions = ptr.To(int32(intent.Spec.Workload.Batch.ParallelTasks))
		}

		if err := i.kubernetesPodTemplate(&job.Spec.Template, intent); err != nil {
			return err
		}

		return controllerutil.SetOwnerReference(intent, &job, i.Client.Scheme())
	})
	if err != nil {
		return err
	}

	return nil
}

//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=create;get;list;watch;update

func (i *IntentReconciler) kubernetesDeployment(ctx context.Context, intent *flarev1alpha1.Intent) error {
	var deployment appsv1.Deployment
	deployment.Name = intent.Namespace
	deployment.Namespace = intent.Namespace

	_, err := controllerutil.CreateOrUpdate(ctx, i.Client, &deployment, func() error {
		deployment.Spec.Replicas = ptr.To(int32(1))
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"intent": intent.Name,
			},
		}
		if err := i.kubernetesPodTemplate(&deployment.Spec.Template, intent); err != nil {
			return err
		}

		return controllerutil.SetOwnerReference(intent, &deployment, i.Client.Scheme())
	})

	return err
}

func (i *IntentReconciler) kubernetesPodTemplate(podTemplate *corev1.PodTemplateSpec, intent *flarev1alpha1.Intent) error {
	podTemplate.Labels = map[string]string{
		"intent": intent.Name,
	}

	if len(podTemplate.Spec.Containers) != 1 {
		podTemplate.Spec.Containers = make([]corev1.Container, 1)
	}

	podTemplate.Spec.Containers[0].Name = intent.Spec.Workload.Name
	podTemplate.Spec.Containers[0].Image = intent.Spec.Workload.Image
	podTemplate.Spec.Containers[0].Command = intent.Spec.Workload.Commands
	podTemplate.Spec.Containers[0].Env = func() []corev1.EnvVar {
		out := make([]corev1.EnvVar, 0, len(intent.Spec.Workload.Env)+len(intent.Spec.Workload.Secrets))

		for _, env := range intent.Spec.Workload.Env {
			var envVar corev1.EnvVar

			if parts := strings.Split(env, "="); len(parts) == 2 {
				envVar = corev1.EnvVar{
					Name:  parts[0],
					Value: parts[1],
				}
			} else {
				envVar = corev1.EnvVar{
					Name:  parts[0],
					Value: "",
				}
			}

			out = append(out, envVar)
		}

		for _, secret := range intent.Spec.Workload.Secrets {
			out = append(out, corev1.EnvVar{
				Name: secret.Env,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: intent.Name,
						},
						Key: secret.Name,
					},
				},
			})
		}

		return out
	}()

	if len(podTemplate.Spec.Containers[0].VolumeMounts) != len(intent.Spec.Workload.Storage.Volumes) {
		podTemplate.Spec.Containers[0].VolumeMounts = make([]corev1.VolumeMount, 0, len(intent.Spec.Workload.Storage.Volumes))
	}

	if len(podTemplate.Spec.Volumes) != len(intent.Spec.Workload.Storage.Volumes) {
		podTemplate.Spec.Volumes = make([]corev1.Volume, len(intent.Spec.Workload.Storage.Volumes))
	}

	for index, volume := range intent.Spec.Workload.Storage.Volumes {
		podTemplate.Spec.Containers[0].VolumeMounts[index].Name = volume.Name
		podTemplate.Spec.Containers[0].VolumeMounts[index].MountPath = volume.Path

		podTemplate.Spec.Volumes[index].Name = volume.Name
		podTemplate.Spec.Volumes[index].VolumeSource = corev1.VolumeSource{
			EmptyDir: func() *corev1.EmptyDirVolumeSource {
				if volume.Type == "Temporary" {
					return &corev1.EmptyDirVolumeSource{
						SizeLimit: ptr.To(volume.Size),
					}
				}

				return nil
			}(),
			CSI: func() *corev1.CSIVolumeSource {
				if volume.Type == "Persistent" {
					return &corev1.CSIVolumeSource{
						Driver: volume.Source.Type,
						VolumeAttributes: map[string]string{
							"uri": volume.Source.Uri,
						},
						NodePublishSecretRef: &corev1.LocalObjectReference{
							Name: volume.Source.Credentials,
						},
					}
				}

				return nil
			}(),
		}
	}

	return nil
}

//+kubebuilder:rbac:groups="",resources=services,verbs=create;get;list;watch;update

func (i *IntentReconciler) kubernetesService(ctx context.Context, intent *flarev1alpha1.Intent) error {
	if len(intent.Spec.Workload.Ports) == 0 {
		return nil
	}

	var svc corev1.Service
	svc.Name = intent.Namespace
	svc.Namespace = intent.Namespace

	_, err := controllerutil.CreateOrUpdate(ctx, i.Client, &svc, func() error {
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = map[string]string{
			"intent": intent.Name,
		}

		if len(svc.Spec.Ports) != len(intent.Spec.Workload.Ports) {
			svc.Spec.Ports = make([]corev1.ServicePort, len(intent.Spec.Workload.Ports))
		}

		for k, port := range intent.Spec.Workload.Ports {
			svc.Spec.Ports[k].Name = strings.ToLower(port.Protocol) + "-" + strconv.FormatInt(int64(port.Port), 10)
			svc.Spec.Ports[k].Protocol = corev1.Protocol(port.Protocol)
			svc.Spec.Ports[k].Port = port.Port
			svc.Spec.Ports[k].TargetPort = intstr.FromInt32(port.Port)
		}

		return controllerutil.SetOwnerReference(intent, &svc, i.Client.Scheme())
	})
	if err != nil {
		return err
	}

	return nil
}
