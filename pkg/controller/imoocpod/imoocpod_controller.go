package imoocpod

import (
	"context"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"

	k8sv1alpha1 "github.com/imooc-com/imooc/imoocpod-operator/pkg/apis/k8s/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_imoocpod")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ImoocPod Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileImoocPod{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("imoocpod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ImoocPod
	err = c.Watch(&source.Kind{Type: &k8sv1alpha1.ImoocPod{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ImoocPod
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &k8sv1alpha1.ImoocPod{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileImoocPod implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileImoocPod{}

// ReconcileImoocPod reconciles a ImoocPod object
type ReconcileImoocPod struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ImoocPod object and makes changes based on the state read
// and what is in the ImoocPod.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileImoocPod) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ImoocPod")

	// Fetch the ImoocPod instance
	instance := &k8sv1alpha1.ImoocPod{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//1 获取name对应的所有pod
	lbls := labels.Set{
		"app":instance.Name,//资源的label 来查找对应的pod
	}
	existringPods := &corev1.PodList{}
	err = r.client.List(context.TODO(),existringPods,&client.ListOptions{
		Namespace: request.Namespace,
		LabelSelector: labels.SelectorFromSet(lbls),
	})
	if err != nil{
		reqLogger.Error(err,"取已经存在的pod失败")
		return reconcile.Result{},err
	}
	//2 获取pod列表中的所有pod name
	var existringPodNames []string
	for _,pod := range existringPods.Items{
		//如果删除时间戳不为0  则跳过添加操作
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil{
			continue
		}

		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning{
			existringPodNames = append(existringPodNames,pod.GetObjectMeta().GetName())
		}


	}
	//3 :update pod.status == 运行中的status
	// 比较DeepEqual
	status := k8sv1alpha1.ImoocPodStatus{
		PodNames: existringPodNames,
		Replicas: len(existringPodNames),
	}
	if !reflect.DeepEqual(instance.Status,status){
		instance.Status = status //把期望状态给运行状态
		err := r.client.Status().Update(context.TODO(),instance)
		if err!= nil{
			reqLogger.Error(err,"更新pod失败")
		}
	}




	//4 len(pod) > 运行中的len(pod.replices) 期望值小 需要scale down
	if len(existringPodNames) > instance.Spec.Replicas{
		//delete
		reqLogger.Info("正在删除pod，当前的existringPodNames和期望的replicas",existringPodNames,instance.Spec.Replicas)
		pod := existringPods.Items[0]
		err := r.client.Delete(context.TODO(),&pod)
		if err != nil{
			reqLogger.Error(err,"删除pod失败")
			return reconcile.Result{},err
		}
	}


	//5 len(pod) < 运行中的len(pod.replices) 期望值大，需要scale up create
	if len(existringPodNames) < instance.Spec.Replicas{
		reqLogger.Info("正在删除pod，当前的existringPodNames和期望的replicas",existringPodNames,instance.Spec.Replicas)
		// Define a new Pod object
		pod := newPodForCR(instance)
		// Set ImoocPod instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		err = r.client.Create(context.TODO(),pod)
		if err != nil{
			reqLogger.Error(err,"创建pod失败")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{Requeue: true},nil










}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *k8sv1alpha1.ImoocPod) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
