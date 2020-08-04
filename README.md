# Kubernetes Operator 快速入门教程
个人介绍
方玉龙
湖北大学 大三学弟
兴趣k8s 欢迎交流


 联系方式
qq   3095329264
邮箱  3095329264@qq.com
tel  15527142537






## 环境需求

Go 语言的依赖管理工具包：[dep](https://github.com/golang/dep)

operator-sdk

operator sdk 安装方法非常多，我们可以直接在 github 上面下载需要使用的版本，然后放置到 PATH 环境下面即可，当然也可以将源码 clone 到本地手动编译安装即可，如果你是 Mac，当然还可以使用常用的 brew 工具进行安装：

```powershell
$ brew install operator-sdk
......
$ operator-sdk version
operator-sdk version: v0.7.0
$ go version
go version go1.11.4 darwin/amd64
```

## 开始创建

### 创建新项目

operator-sdk new imoocpod-operator --skip-validation=true --repo=github.com/imooc-com/imooc/imoocpod-operator

```shell
operator-sdk new opdemo
```

```shell
INFO[0000] Creating new Go operator 'opdemo'.           
INFO[0000] Created go.mod                               
INFO[0000] Created tools.go                             
INFO[0000] Created cmd/manager/main.go                  
INFO[0000] Created build/Dockerfile                     
INFO[0000] Created build/bin/entrypoint                 
INFO[0000] Created build/bin/user_setup                 
INFO[0000] Created deploy/service_account.yaml          
INFO[0000] Created deploy/role.yaml                     
INFO[0000] Created deploy/role_binding.yaml             
INFO[0000] Created deploy/operator.yaml                 
INFO[0000] Created pkg/apis/apis.go                     
INFO[0000] Created pkg/controller/controller.go         
INFO[0000] Created version/version.go                   
INFO[0000] Created .gitignore                           
INFO[0000] Validating project                           
INFO[0031] Project validation successful.               
INFO[0031] Project creation complete. 
```

```shell
cd opdemo && tree -L 2
```

```shell
.
├── Gopkg.lock
├── Gopkg.toml
├── build
│   ├── Dockerfile
│   ├── _output
│   └── bin
├── cmd
│   └── manager
├── deploy
│   ├── crds
│   ├── operator.yaml
│   ├── role.yaml
│   ├── role_binding.yaml
│   └── service_account.yaml
├── pkg
│   ├── apis
│   └── controller
├── vendor
│   ├── cloud.google.com
│   ├── contrib.go.opencensus.io
│   ├── github.com
│   ├── go.opencensus.io
│   ├── go.uber.org
│   ├── golang.org
│   ├── google.golang.org
│   ├── gopkg.in
│   ├── k8s.io
│   └── sigs.k8s.io
└── version
    └── version.go

23 directories, 8 files
```

- **Gopkg.toml Gopkg.lock** — Go Dep 清单，用来描述当前 Operator 的依赖包。
- **cmd** - 包含 main.go 文件，使用 operator-sdk API 初始化和启动当前 Operator 的入口。
- **deploy** - 包含一组用于在 Kubernetes 集群上进行部署的通用的 Kubernetes 资源清单文件。
- **pkg/apis** - 包含定义的 API 和自定义资源（CRD）的目录树，这些文件允许 sdk 为 CRD 生成代码并注册对应的类型，以便正确解码自定义资源对象。
- **pkg/controller** - 用于编写所有的操作业务逻辑的地方
- **vendor** - golang vendor 文件夹，其中包含满足当前项目的所有外部依赖包，通过 go dep 管理该目录。

### 添加API

 operator-sdk add api --api-version=k8s.imooc.com/v1alpha1 --kind=ImoocPod

```shell
operator-sdk add api --api-version=app.example.com/v1 --kind=AppService
```

```powershell
INFO[0000] Generating api version app.example.com/v1 for kind AppService. 
INFO[0000] Created pkg/apis/app/group.go                
INFO[0031] Created pkg/apis/app/v1/appservice_types.go  
INFO[0062] Created pkg/apis/addtoscheme_app_v1.go       
INFO[0062] Created pkg/apis/app/v1/register.go          
INFO[0062] Created pkg/apis/app/v1/doc.go               
INFO[0062] Created deploy/crds/app.example.com_v1_appservice_cr.yaml 
INFO[0062] Running deepcopy code-generation for Custom Resource group versions: [app:[v1], ] 
INFO[0079] Code-generation complete.                    
INFO[0079] Running CRD generator.                       
INFO[0080] CRD generation complete.                     
INFO[0080] API generation complete.                     
INFO[0080] API generation complete. 
```

pkg/apis 下多出app文件夹

deploy    下多出crds文件夹

### 添加控制器

operator-sdk add controller --api-version=k8s.imcco.com/v1alpha1 --kind=ImoocPod

```shell
 operator-sdk add controller --api-version=app.example.com/v1 --kind=AppService
```

多出pkg/controller/appservice/appservice_controller.go 

多出pkg/controller/add_appservice.go





### 部署文件

 operator-sdk build mock.com:5000/imoocpod-operator

docker push mock.com:5000/imoocpod-operator

kubectl apply -f deploy/service_account.yaml 

 kubectl apply -f deploy/role.yaml 

 kubectl apply -f deploy/role_binding.yaml 

 kubectl apply -f deploy/crds/k8s.imooc.com_imoocpods_crd.yaml

替换operator.yaml 中的image

 kubectl apply -f deploy/operator.yaml

kubectl apply -f deploy/crds/k8s.imooc.com_v1alpha1_imoocpod_cr.yaml





### 自定义字段

apis/v1alpha1/types.go

```go
type ImoocPodSpec struct {
	Replicas int `json:"replicas"`
}

type ImoocPodStatus struct {
	Replicas int      `json:"replicas"`
	PodNames []string `json:"podNames"`
}
```



 operator-sdk generate k8s

operator-sdk generate crds







### 自定义控制器

Controller.go

```go
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

```



```go
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

```

 operator-sdk generate k8s



### 重新部署

operator-sdk build mock.com:5000/imoocpod-operator

docker push mock.com:5000/imoocpod-operator



删除cr 和crd文件operator文件 重新部署

```shell
fangyulong@BDSZYF000146577 imoocpod-operator % kubectl get pod
NAME                                 READY   STATUS    RESTARTS   AGE
example-imoocpod-pod2ghzt            1/1     Running   0          63s
example-imoocpod-pod6fwc5            1/1     Running   0          63s
example-imoocpod-pods7wbp            1/1     Running   0          63s
imoocpod-operator-7959bdcb7d-f9hgf   1/1     Running   0          69s
```







### 