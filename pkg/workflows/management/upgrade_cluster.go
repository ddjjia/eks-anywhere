package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type upgradeCluster struct{}

// Run upgradeCluster performs actions needed to upgrade the management cluster.
func (s *upgradeCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	// TODO(g-gaston): move this to eks-a installer and eks-d installer
	err := commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Upgrading management cluster")
	if err := commandContext.ClusterUpgrader.Run(ctx, commandContext.ClusterSpec, *commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		// TODO(@pjshah): check if we need this or not
		// Take backup of bootstrap cluster capi components
		// if commandContext.BootstrapCluster != nil {
		// 	logger.Info("Backing up management components from bootstrap cluster")
		// 	err := commandContext.ClusterManager.BackupCAPIWaitForInfrastructure(ctx, commandContext.BootstrapCluster, fmt.Sprintf("bootstrap-%s", commandContext.ManagementClusterStateDir), commandContext.ManagementCluster.Name)
		// 	if err != nil {
		// 		logger.Info("Bootstrap management component backup failed, use existing workload cluster backup", "error", err)
		// 	}
		// }
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &reconcileGitOps{}
}

func (s *upgradeCluster) Name() string {
	return "upgrade-workload-cluster"
}

func (s *upgradeCluster) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *upgradeCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &reconcileGitOps{}, nil
}
