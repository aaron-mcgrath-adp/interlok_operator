/*
Copyright 2021.

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

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intergrationsv1alpha1 "github.com/aaron-mcgrath-adp/interlok_operator/api/v1alpha1"
)

const (
	ProfilerEnv = "-javaagent:lib/aspectjweaver.jar -Dorg.aspectj.weaver.loadtime.configuration=META-INF/profiler-aop.xml"
	TimeToLive  = "-Dsun.net.inetaddr.ttl=30"
)

// InterlokReconciler reconciles a Interlok object
type InterlokReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=intergrations.proagrica.com,resources=interloks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intergrations.proagrica.com,resources=interloks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intergrations.proagrica.com,resources=interloks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Interlok object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *InterlokReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)
	logr.Info("Interlok operator invoked after event", "name", req.NamespacedName)

	interlok := &intergrationsv1alpha1.Interlok{}

	if err := r.Get(ctx, req.NamespacedName, interlok); err != nil {
		logr.Info("Interlok object not found.  Ignoring event.", "name", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the deployment already exists, if not create a new one
	interlokDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: interlok.Name, Namespace: interlok.Namespace}, interlokDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForInterlok(interlok)
		logr.Info("Creating a new Interlok Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			logr.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logr.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := interlok.Spec.Instances
	if *interlokDeployment.Spec.Replicas != size {
		interlokDeployment.Spec.Replicas = &size
		err = r.Update(ctx, interlokDeployment)
		if err != nil {
			logr.Error(err, "Failed to update Deployment", "Deployment.Namespace", interlokDeployment.Namespace, "Deployment.Name", interlokDeployment.Name)
			return ctrl.Result{}, err
		}
		// Ask to requeue after 1 minute in order to give enough time for the
		// pods be created on the cluster side and the operand be able
		// to do the next update step accurately.
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	return ctrl.Result{}, nil
}

// deploymentForInterlok returns an Interlok Deployment object
func (r *InterlokReconciler) deploymentForInterlok(m *intergrationsv1alpha1.Interlok) *appsv1.Deployment {
	ls := labelsForInterlok(m.Name)

	replicas := m.Spec.Instances
	image := m.Spec.Image
	port := m.Spec.JettyPort
	javaopts := TimeToLive

	if m.Spec.WithProfiler {
		javaopts = javaopts + " " + ProfilerEnv
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: image,
						Name:  "interlok",
						Env: []corev1.EnvVar{
							{
								Name:  "PORT",
								Value: strconv.Itoa(int(port)),
							},
							{
								Name:  "JAVA_OPTS",
								Value: javaopts,
							},
						},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: port,
								Name:          "interlok",
							},
						},
					}},
				},
			},
		},
	}
	// Set Interlok instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsForInterlok returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForInterlok(name string) map[string]string {
	return map[string]string{"app": "interlok", "interlok_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InterlokReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&intergrationsv1alpha1.Interlok{}).
		Complete(r)
}
