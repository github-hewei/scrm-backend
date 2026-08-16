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

当前支持通讯录(dept)/客户(contact)/客户群(group)/全部(all)同步。

示例:
  # 同步指定企业全部数据（通讯录+客户+客户群）
  cli wecom-sync --store-id=5 --scope=all

  # 只同步某企业的通讯录（部门+成员）
  cli wecom-sync --store-id=5 --scope=dept

  # 只同步某企业的客户
  cli wecom-sync --store-id=5 --scope=contact

  # 只同步某企业的客户群
  cli wecom-sync --store-id=5 --scope=group`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.NewWecomSyncRunner(log, svc).Run(cmd.Context(), storeId, scope)
		},
	}
	cmd.Flags().Uint32Var(&storeId, "store-id", 0, "企业ID(0=全部已接入企业)")
	cmd.Flags().StringVar(&scope, "scope", "group", "同步范围: all|dept|contact|group")
	return cmd
}
