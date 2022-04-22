package controller

import (
	"context"
	"github.com/symcn/hparecord/pkg/metrics"
	"k8s.io/api/autoscaling/v2beta2"
	"sync"
	"time"

	"github.com/symcn/api"
	"github.com/symcn/hparecord/pkg/kube"
	symcnclient "github.com/symcn/pkg/clustermanager/client"
	"github.com/symcn/pkg/clustermanager/configuration"
	"github.com/symcn/pkg/clustermanager/handler"
	"github.com/symcn/pkg/clustermanager/predicate"
	"github.com/symcn/pkg/clustermanager/workqueue"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"k8s.io/utils/trace"
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
	ctrl.registryBeforeAfterHandler()

	return ctrl, nil
}

func (ctrl *Controller) Start() error {
	return ctrl.MultiMingleClient.Start(ctrl.ctx)
}

func (ctrl *Controller) registryBeforeAfterHandler() {
	go metrics.Start(ctrl.ctx)

	ctrl.RegistryBeforAfterHandler(func(ctx context.Context, cli api.MingleClient) error {
		queue, err := workqueue.Complted(workqueue.NewWrapQueueConfig(cli.GetClusterCfgInfo().GetName(), ctrl)).NewQueue()
		if err != nil {
			return err
		}

		go queue.Start(ctx)

		cli.AddResourceEventHandler(
			&v2beta2.HorizontalPodAutoscaler{},
			handler.NewResourceEventHandler(
				queue,
				handler.NewDefaultTransformNamespacedNameEventHandler(),
				predicate.NamespacePredicate("*"),
			),
		)

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

	hpa := &v2beta2.HorizontalPodAutoscaler{}
	err = cli.Get(req.NamespacedName, hpa)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Warningf("Get cluster %s hpa %s failed: %+v", cli.GetClusterCfgInfo().GetName(), req.String(), err)
		}
		return api.Done, 0, nil
	}
	tr.Step("getHpa")

	if err := handleMetrics(req.QName, hpa); err != nil {
		return api.Requeue, time.Second * 5, err
	}
	tr.Step("handleMetrics")

	return api.Done, 0, nil
}
