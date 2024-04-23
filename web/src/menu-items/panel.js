// assets
import {
  IconDashboard,
  IconSitemap,
  IconArticle,
  IconCoin,
  IconSettingsCog,
  IconKey,
  IconCreditCard,
  IconUser,
  IconUserScan,
  IconChartHistogram,
  IconBrandTelegram,
  IconReceipt2,
  IconBrush,
  IconBrandGithubCopilot
} from '@tabler/icons-react';

// constant
const icons = {
  IconDashboard,
  IconSitemap,
  IconArticle,
  IconCoin,
  IconSettingsCog,
  IconKey,
  IconCreditCard,
  IconUser,
  IconUserScan,
  IconChartHistogram,
  IconBrandTelegram,
  IconReceipt2,
  IconBrush,
  IconBrandGithubCopilot
};

// ==============================|| DASHBOARD MENU ITEMS ||============================== //

const panel = {
  id: 'panel',
  type: 'group',
  children: [
    {
      id: 'dashboard',
      title: '仪表盘',
      type: 'item',
      url: '/panel/dashboard',
      icon: icons.IconDashboard,
      breadcrumbs: false,
      isAdmin: false
    },
    {
      id: 'analytics',
      title: '分析',
      type: 'item',
      url: '/panel/analytics',
      icon: icons.IconChartHistogram,
      breadcrumbs: false,
      isAdmin: true
    },
    {
      id: 'channel',
      title: '渠道',
      type: 'item',
      url: '/panel/channel',
      icon: icons.IconSitemap,
      breadcrumbs: false,
      isAdmin: true
    },
    {
      id: 'token',
      title: '令牌',
      type: 'item',
      url: '/panel/token',
      icon: icons.IconKey,
      breadcrumbs: false
    },
    {
      id: 'log',
      title: '日志',
      type: 'item',
      url: '/panel/log',
      icon: icons.IconArticle,
      breadcrumbs: false
    },
    {
      id: 'redemption',
      title: '兑换',
      type: 'item',
      url: '/panel/redemption',
      icon: icons.IconCoin,
      breadcrumbs: false,
      isAdmin: true
    },
    {
      id: 'topup',
      title: '充值',
      type: 'item',
      url: '/panel/topup',
      icon: icons.IconCreditCard,
      breadcrumbs: false
    },
    {
      id: 'midjourney',
      title: 'Midjourney',
      type: 'item',
      url: '/panel/midjourney',
      icon: icons.IconBrush,
      breadcrumbs: false
    },
    {
      id: 'user',
      title: '用户',
      type: 'item',
      url: '/panel/user',
      icon: icons.IconUser,
      breadcrumbs: false,
      isAdmin: true
    },
    {
      id: 'profile',
      title: '个人设置',
      type: 'item',
      url: '/panel/profile',
      icon: icons.IconUserScan,
      breadcrumbs: false,
      isAdmin: true
    },
    {
      id: 'pricing',
      title: '模型价格',
      type: 'item',
      url: '/panel/pricing',
      icon: icons.IconReceipt2,
      breadcrumbs: false,
      isAdmin: true
    },
    {
      id: 'model_price',
      title: '可用模型',
      type: 'item',
      url: '/panel/model_price',
      icon: icons.IconBrandGithubCopilot,
      breadcrumbs: false,
      isAdmin: true
    },
    {
      id: 'setting',
      title: '设置',
      type: 'item',
      url: '/panel/setting',
      icon: icons.IconSettingsCog,
      breadcrumbs: false,
      isAdmin: true
    },
    {
      id: 'telegram',
      title: 'Telegram Bot',
      type: 'item',
      url: '/panel/telegram',
      icon: icons.IconBrandTelegram,
      breadcrumbs: false,
      isAdmin: true
    }
  ]
};

export default panel;
