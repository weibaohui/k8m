import * as React from 'react';
import { message } from 'antd';

interface KubeconfigDownloadButtonProps {
  data: {
    id: number;
    cluster: string;
    display_name?: string;
  };
}

const KubeconfigDownloadButton: React.FC<KubeconfigDownloadButtonProps> = ({ data }) => {
  const handleDownload = () => {
    try {
      const token = localStorage.getItem('token') || '';
      const params = new URLSearchParams({
        token: token
      }).toString();

      let url = `/mgm/plugins/kubeconfig_export/kubeconfig/${data.id}/export`;
      if (params) {
        url += `?${params}`;
      }

      const a = document.createElement('a');
      a.href = url;
      a.download = `${data.cluster}-kubeconfig.yaml`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);

      message.success('Kubeconfig 文件正在下载...');
    } catch (e) {
      message.error('下载失败，请重试');
    }
  };

  return (
    <button
      className="cxd-Button cxd-Button--primary cxd-Button--size-default"
      onClick={handleDownload}
    >
      <i className="fa-solid fa-download" style={{ marginRight: 4 }}></i>
      导出 Kubeconfig
    </button>
  );
};

export default KubeconfigDownloadButton;