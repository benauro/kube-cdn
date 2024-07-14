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
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

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

//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete

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

	// Handle ingress
	if err := r.reconcileIngress(ctx, &cdn); err != nil {
		logger.Error(err, "Failed to reconcile ingress")
		return ctrl.Result{}, err
	}

	// Handle service
	if err := r.reconcileService(ctx, &cdn); err != nil {
		logger.Error(err, "Failed to reconcile service")
		return ctrl.Result{}, err
	}

	// Auto scaling
	if err := r.autoScale(ctx, &cdn); err != nil {
		logger.Error(err, "Failed to auto-scale CDN nodes")
		return ctrl.Result{}, err
	}

	// Handle networking
	if err := r.reconcileNetworking(ctx, &cdn); err != nil {
		logger.Error(err, "Failed to reconcile networking")
		return ctrl.Result{}, err
	}

	// Handle storage
	if err := r.reconcileStorage(ctx, &cdn); err != nil {
		logger.Error(err, "Failed to reconcile storage")
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

// SetupWithManager sets up the controller with the Manager.
func (r *ContentDeliveryNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cdnv3.ContentDeliveryNetwork{}).
		Owns(&networkingv1.Ingress{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *ContentDeliveryNetworkReconciler) applyCacheRules(cdn *cdnv3.ContentDeliveryNetwork) error {
	// TODO:
	_ = cdn
	return nil
}

func (r *ContentDeliveryNetworkReconciler) reconcileService(ctx context.Context, cdn *cdnv3.ContentDeliveryNetwork) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cdn.Name + "-service",
			Namespace: cdn.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": cdn.Name},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

	err := r.Create(ctx, service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if errors.IsAlreadyExists(err) {
		return r.Update(ctx, service)
	}

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

func (r *ContentDeliveryNetworkReconciler) reconcileIngress(ctx context.Context, cdn *cdnv3.ContentDeliveryNetwork) error {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cdn.Name + "-ingress",
			Namespace: cdn.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/$1",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: cdn.Spec.DomainName,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									PathType: func() *networkingv1.PathType {
										t := networkingv1.PathTypePrefix
										return &t
									}(),
									Path: "/",
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: cdn.Name + "-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	err := r.Create(ctx, ingress)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if errors.IsAlreadyExists(err) {
		return r.Update(ctx, ingress)
	}

	return nil
}

func (r *ContentDeliveryNetworkReconciler) reconcileNetworking(ctx context.Context, cdn *cdnv3.ContentDeliveryNetwork) error {
	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cdn.Name + "-network-policy",
			Namespace: cdn.Namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"app": cdn.Name},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 80}},
					},
				},
			},
		},
	}

	if err := r.Create(ctx, networkPolicy); err != nil {
		return err
	}

	return nil
}

func (r *ContentDeliveryNetworkReconciler) reconcileStorage(ctx context.Context, cdn *cdnv3.ContentDeliveryNetwork) error {
	// Create or update DataVolume
	dataVolume := &cdiv1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cdn.Name + "-data",
			Namespace: cdn.Namespace,
		},
		Spec: cdiv1.DataVolumeSpec{
			Source: &cdiv1.DataVolumeSource{
				Blank: &cdiv1.DataVolumeBlankImage{},
			},
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			},
		},
	}

	err := r.Create(ctx, dataVolume)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if errors.IsAlreadyExists(err) {
		return r.Update(ctx, dataVolume)
	}

	return nil
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
							Image: "benauro/kube-cdn:latest",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "content",
									MountPath: "/data",
								},
							},
							ImagePullPolicy: cdn.Spec.ImagePullPolicy, // Use imagePullPolicy from CDN spec
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "content",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: cdn.Name + "-storage",
								},
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
