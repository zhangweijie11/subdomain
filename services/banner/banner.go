package banner

import (
	"fmt"
	osutil "github.com/projectdiscovery/utils/os"
	"gitlab.example.com/zhangweijie/subdomain/services/privileges"
	"gitlab.example.com/zhangweijie/tool-sdk/middleware/logger"
)

// ShowNetworkCapabilities 显示网络功能扫描类型可能与正在运行的用户
func ShowNetworkCapabilities() {
	var accessLevel string

	switch {
	case privileges.IsOSSupported && privileges.IsPrivileged:
		accessLevel = "root"
		if osutil.IsLinux() {
			accessLevel = "CAP_NET_RAW"
		}
	default:
		accessLevel = "non root"
	}
	logger.Info(fmt.Sprintf("Running scan with %s privileges", accessLevel))
}
