package cli

import (
	"zero-backend/internal/cli/runner"

	"github.com/241x/zero-kit/logger"
	"github.com/spf13/cobra"
)

// WecomSyncCmd 企业微信数据同步命令
func WecomSyncCmd(log logger.Logger, svc runner.WecomSyncService) *cobra.Command {
	var storeId uint32
	var scope string

	cmd := &cobra.Command{
		Use:   "wecom-sync",
		Short: "企业微信数据同步",
		Long: `企业微信数据同步命令，从企微拉取客户群数据写入本地。

当前仅支持客户群同步（scope=group），通讯录/客户同步接入中。

示例:
  # 同步指定企业的客户群
  cli wecom-sync --store-id=5

  # 同步全部已接入企业的客户群
  cli wecom-sync`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.NewWecomSyncRunner(log, svc).Run(cmd.Context(), storeId, scope)
		},
	}
	cmd.Flags().Uint32Var(&storeId, "store-id", 0, "企业ID(0=全部已接入企业)")
	cmd.Flags().StringVar(&scope, "scope", "group", "同步范围: group (all/dept/contact 暂未接入)")
	return cmd
}
