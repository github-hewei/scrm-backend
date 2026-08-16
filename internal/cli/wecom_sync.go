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
		Long: `企业微信数据同步命令，从企微拉取客户/客户群数据写入本地。

当前支持客户群(group)与外部联系人/客户(contact)同步，通讯录同步接入中。

示例:
  # 同步指定企业的客户群
  cli wecom-sync --store-id=5 --scope=group

  # 同步指定企业的客户
  cli wecom-sync --store-id=5 --scope=contact

  # 同步全部已接入企业的客户群与客户
  cli wecom-sync`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.NewWecomSyncRunner(log, svc).Run(cmd.Context(), storeId, scope)
		},
	}
	cmd.Flags().Uint32Var(&storeId, "store-id", 0, "企业ID(0=全部已接入企业)")
	cmd.Flags().StringVar(&scope, "scope", "group", "同步范围: group|contact (dept 暂未接入)")
	return cmd
}
