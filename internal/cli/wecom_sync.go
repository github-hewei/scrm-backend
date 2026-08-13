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
		Long: `企业微信数据同步命令，从企微拉取通讯录/客户/客户群数据写入本地。

示例:
  # 同步指定企业全部数据
  cli wecom-sync --store-id=5

  # 只同步某企业的客户
  cli wecom-sync --store-id=5 --scope=contact

  # 只同步通讯录（部门+成员）
  cli wecom-sync --store-id=5 --scope=dept`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.NewWecomSyncRunner(log, svc).Run(cmd.Context(), storeId, scope)
		},
	}
	cmd.Flags().Uint32Var(&storeId, "store-id", 0, "企业ID(0=全部已接入企业)")
	cmd.Flags().StringVar(&scope, "scope", "all", "同步范围: all|dept|contact|group")
	return cmd
}
