import { MenuItem } from '@/types/menu';

export const initialMenu: MenuItem[] = [
  {
    key: '1',
    title: '首页',
    icon: 'fa-home',
    order: 1,
    children: [
      {
        key: '1-1',
        title: '仪表盘',
        icon: 'fa-tachometer-alt',
        order: 1,
      },
      {
        key: '1-2',
        title: '项目管理',
        icon: 'fa-project-diagram',
        order: 2,
      },
      {
        key: '1-3',
        title: '设置',
        icon: 'fa-cog',
        order: 3,
        children: [
          {
            key: '1-3-1',
            title: '用户设置',
            icon: 'fa-user',
            eventType: 'url',
            url: 'http://www.baidu.com',
            order: 1,
          },
          {
            key: '1-3-2',
            title: '权限设置',
            icon: 'fa-shield-alt',
            order: 2,
          },
        ],
      },
    ],
  },
];