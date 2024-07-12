/*
Copyright 2024.

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

package controller

import (
	"context"
	"math"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cdnv3 "github.com/benauro/kube-cdn/api/v3"
)

// ContentDeliveryNetworkReconciler reconciles a ContentDeliveryNetwork object
type ContentDeliveryNetworkReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cdn.benauro.gg,resources=contentdeliverynetworks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cdn.benauro.gg,resources=contentdeliverynetworks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cdn.benauro.gg,resources=contentdeliverynetworks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ContentDeliveryNetwork object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ContentDeliveryNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var cdn cdnv3.ContentDeliveryNetwork
	if err := r.Get(ctx, req.NamespacedName, &cdn); err != nil {
		logger.Error(err, "Unable to fetch ContentDeliveryNetwork")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Apply caching rules
	if err := r.applyCacheRules(&cdn); err != nil {
		logger.Error(err, "Failed to apply cache rules")
		return ctrl.Result{}, err
	}

	// Auto scaling
	if err := r.autoScale(ctx, &cdn); err != nil {
		logger.Error(err, "Failed to auto-scale CDN nodes")
		return ctrl.Result{}, err
	}

	// Update metrics
	if err := r.updateMetrics(ctx, &cdn); err != nil {
		logger.Error(err, "Failed to update metrics")
		return ctrl.Result{}, err
	}

	// Updata status
	cdn.Status.State = "Ready"
	cdn.Status.LastUpdated = metav1.Now()

	if err := r.Status().Update(ctx, &cdn); err != nil {
		logger.Error(err, "Unable to update ContentDeliveryNetwork status")
		return ctrl.Result{}, err
	}

	// Requeue per minute
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *ContentDeliveryNetworkReconciler) applyCacheRules(cdn *cdnv3.ContentDeliveryNetwork) error {
	// TODO:
	_ = cdn
	return nil
}

func (r *ContentDeliveryNetworkReconciler) autoScale(ctx context.Context, cdn *cdnv3.ContentDeliveryNetwork) error {
	// Get current deployment
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Namespace: cdn.Namespace, Name: cdn.Name + "-deployment"}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Creat one if not found
			return r.createCDNDeployment(ctx, cdn)
		}
		return err
	}

	// Decide whether to auto-scale or not
	desiredReplicas := int32(calculateDesiredReplicas(cdn))

	if *deployment.Spec.Replicas != desiredReplicas {
		deployment.Spec.Replicas = &desiredReplicas
		if err := r.Update(ctx, deployment); err != nil {
			return err
		}
	}

	return nil
}

func (r *ContentDeliveryNetworkReconciler) updateMetrics(ctx context.Context, cdn *cdnv3.ContentDeliveryNetwork) error {
	// TODO:
	_, _ = ctx, cdn
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContentDeliveryNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cdnv3.ContentDeliveryNetwork{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func calculateDesiredReplicas(cdn *cdnv3.ContentDeliveryNetwork) int {
	requestsPerReplica := 100.0 // Assume each replica can handle 100 QPS
	requestsPerSecond, _ := strconv.ParseFloat(cdn.Status.Metrics.RequestsPerSecond, 64)
	desiredReplicas := int(math.Ceil(requestsPerSecond / requestsPerReplica))

	if desiredReplicas < cdn.Spec.MinReplicas {
		return cdn.Spec.MinReplicas
	}
	if desiredReplicas > cdn.Spec.MaxReplicas {
		return cdn.Spec.MaxReplicas
	}
	return desiredReplicas
}

func (r *ContentDeliveryNetworkReconciler) createCDNDeployment(ctx context.Context, cdn *cdnv3.ContentDeliveryNetwork) error {
	replicas := int32(cdn.Spec.MinReplicas)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cdn.Name + "-deployment",
			Namespace: cdn.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": cdn.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": cdn.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "cdn-node",
							Image: "your-cdn-node-image:latest",
							Env: []corev1.EnvVar{
								{Name: "ORIGIN", Value: cdn.Spec.Origin},
								{Name: "DOMAIN_NAME", Value: cdn.Spec.DomainName},
							},
						},
					},
				},
			},
		},
	}

	if err := r.Create(ctx, deployment); err != nil {
		return err
	}

	return nil
}
