package controller

import (
	"context"
	"sync"
	"time"

	"github.com/symcn/api"
	"github.com/symcn/hparecord/pkg/kube"
	symcnclient "github.com/symcn/pkg/clustermanager/client"
	"github.com/symcn/pkg/clustermanager/configuration"
	"github.com/symcn/pkg/clustermanager/handler"
	"github.com/symcn/pkg/clustermanager/predicate"
	"github.com/symcn/pkg/clustermanager/workqueue"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/trace"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	configmapDataName = "hpaRecord"
)

type Controller struct {
	ctx context.Context

	api.MultiMingleClient
	sync.Mutex
}

func New(ctx context.Context, mcc *symcnclient.MultiClientConfig) (*Controller, error) {
	mcc.ClusterCfgManager = configuration.NewClusterCfgManagerWithCM(
		kube.ManagerPlaneClusterClient.GetKubeInterface(),
		"sym-admin",
		map[string]string{"ClusterOwner": "sym-admin"},
		"kubeconfig.yaml",
		"status",
	)

	cc, err := symcnclient.Complete(mcc)
	if err != nil {
		return nil, err
	}

	mc, err := cc.New()

	ctrl := &Controller{
		ctx:               ctx,
		MultiMingleClient: mc,
	}
	ctrl.registryBeforAfterHandler()

	return ctrl, nil
}

func (ctrl *Controller) Start() error {
	return ctrl.MultiMingleClient.Start(ctrl.ctx)
}

func (ctrl *Controller) registryBeforAfterHandler() {
	go startHTTPServer(ctrl.ctx)

	ctrl.RegistryBeforAfterHandler(func(ctx context.Context, cli api.MingleClient) error {
		queue, err := workqueue.Complted(workqueue.NewWrapQueueConfig(cli.GetClusterCfgInfo().GetName(), ctrl)).NewQueue()
		if err != nil {
			return err
		}

		go queue.Start(ctx)

		cli.AddResourceEventHandler(
			&corev1.Event{},
			handler.NewResourceEventHandler(
				queue,
				handler.NewDefaultTransformNamespacedNameEventHandler(),
				predicate.NamespacePredicate("*"),
				filterHpaEvent(),
			),
		)
		_, err = cli.GetInformer(&autoscalingv1.HorizontalPodAutoscaler{})
		if err != nil {
			return err
		}
		_, err = cli.GetInformer(&corev1.ConfigMap{})
		if err != nil {
			return err
		}

		return nil
	})
}

func (ctrl *Controller) Reconcile(req api.WrapNamespacedName) (requeue api.NeedRequeue, after time.Duration, err error) {
	tr := trace.New("hpa-event-collector",
		trace.Field{Key: "cluster", Value: req.QName},
		trace.Field{Key: "namespace", Value: req.Namespace},
		trace.Field{Key: "name", Value: req.Name},
	)
	defer tr.LogIfLong(time.Millisecond * 100)

	// get client
	cli, err := ctrl.GetWithName(req.QName)
	if err != nil {
		return api.Requeue, time.Second * 5, err
	}
	tr.Step("GetClientWithClusterName")

	// get event
	e := &corev1.Event{}
	err = cli.Get(req.NamespacedName, e)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Warningf("Get cluster %s event %s failed: %+v", cli.GetClusterCfgInfo().GetName(), req.NamespacedName.String(), err)
		}
		return api.Done, 0, nil
	}
	tr.Step("GetEventWithInformer")

	// get hpa
	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	nreq := types.NamespacedName{Namespace: e.InvolvedObject.Namespace, Name: e.InvolvedObject.Name}
	err = cli.Get(nreq, hpa)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Warningf("Get cluster %s hpa %s failed: %+v", cli.GetClusterCfgInfo().GetName(), nreq.String(), err)
		}
		return api.Done, 0, nil
	}
	tr.Step("GetHpaWithInformer")

	data, ok := hpa.Annotations["autoscaling.alpha.kubernetes.io/current-metrics"]
	if !ok {
		return api.Done, 0, nil
	}

	nreq.Name = nreq.Name + "-hpa-record"
	notFouncCm := false
	cm := &corev1.ConfigMap{}
	err = kube.ManagerPlaneClusterClient.Get(nreq, cm)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Warningf("Get cluster %s configmap %s failed: %+v", cli.GetClusterCfgInfo().GetName(), nreq.String(), err)
			return api.Done, 0, nil
		}
		notFouncCm = true
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nreq.Name,
				Namespace: nreq.Namespace,
			},
			Data: map[string]string{},
		}
	}
	tr.Step("GetConfigmap")

	recordData := kube.BuildAggragageData(cm.Data[configmapDataName], req.QName, data, e.Message)
	cm.Data[configmapDataName] = recordData
	if notFouncCm {
		klog.Infof("Create cluster %s configmap %s", req.QName, nreq.String())
		err = kube.ManagerPlaneClusterClient.Create(cm, &client.CreateOptions{})
		tr.Step("CreateConfigmp")
	} else {
		klog.Infof("Update cluster %s configmap %s", req.QName, nreq.String())
		err = kube.ManagerPlaneClusterClient.Update(cm, &client.UpdateOptions{})
		tr.Step("UpdateConfigmap")
	}
	if err != nil {
		return api.Requeue, time.Second * 5, err
	}
	return api.Done, 0, nil
}
