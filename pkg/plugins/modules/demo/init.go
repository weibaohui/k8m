package demo

import "k8s.io/klog/v2"

func init() {
	if err := InitDB(); err != nil {
		klog.V(6).Infof("初始化Demo插件数据表失败: %v", err)
	} else {
		klog.V(6).Infof("初始化Demo插件数据表成功")
	}
}

