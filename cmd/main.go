package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/symcn/hparecord/pkg/controller"
	"github.com/symcn/hparecord/pkg/kube"
	"github.com/symcn/pkg/clustermanager/client"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	mcc := client.NewMultiClientConfig()

	cmd := &cobra.Command{
		Use:          "hparecord",
		Short:        "hpa event collect",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			printFlags(cmd.Flags())

			ctx := signals.SetupSignalHandler()

			if err := kube.InitManagerPlaneClusterClient(ctx); err != nil {
				return err
			}

			ctrl, err := controller.New(ctx, mcc)
			if err != nil {
				return err
			}

			return ctrl.Start()
		},
	}
	klog.InitFlags(flag.CommandLine)
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().DurationVarP(&mcc.RebuildInterval, "rebuild_interval", "", mcc.RebuildInterval, "Auto invoke multiclusterconfiguration find new cluster or delete old cluster time interval.")
	cmd.PersistentFlags().DurationVarP(&mcc.Options.ExecTimeout, "exec_timeout", "", mcc.Options.ExecTimeout, "Set mingle client exec timeout, if less than default timeout, use default.")
	cmd.PersistentFlags().DurationVarP(&mcc.Options.HealthCheckInterval, "health_check_interval", "", mcc.Options.HealthCheckInterval, "Set mingle clinet check kubernetes connected interval.")
	cmd.PersistentFlags().DurationVarP(&mcc.Options.SyncPeriod, "sync_period", "", mcc.Options.SyncPeriod, "Set informer sync period time interval.")
	cmd.PersistentFlags().StringVarP(&mcc.Options.UserAgent, "user_agent", "", mcc.Options.UserAgent, "client-go connected user-agent.")
	cmd.PersistentFlags().IntVarP(&mcc.Options.QPS, "qps", "", mcc.Options.QPS, "Set mingle client qps for each cluster")
	cmd.PersistentFlags().IntVarP(&mcc.Options.Burst, "burst", "", mcc.Options.Burst, "Set mingle client burst for each cluster")
	cmd.PersistentFlags().BoolVarP(&mcc.Options.LeaderElection, "leader_election", "", mcc.Options.LeaderElection, "Enabled leader election, if true, should set --leader_election_id both.")
	cmd.PersistentFlags().StringVarP(&mcc.Options.LeaderElectionID, "leader_election_id", "", mcc.Options.LeaderElectionID, "Set leader election id.")
	cmd.PersistentFlags().StringVarP(&mcc.Options.LeaderElectionNamespace, "leader_election_ns", "", mcc.Options.LeaderElectionNamespace, "Set leader election namespace.")
	cmd.PersistentFlags().StringVarP(&controller.FilterLabels, "filter_labels", "", controller.FilterLabels, "filter hpa Labels to metrics labels")

	if err := cmd.Execute(); err != nil {
		klog.Errorf("Execute event exporter failed.")
	}
}

func printFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
