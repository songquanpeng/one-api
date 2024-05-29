const config = {
  // basename: only at build time to set, and Don't add '/' at end off BASENAME for breadcrumbs, also Don't put only '/' use blank('') instead,
  // like '/berry-material-react/react/default'
  basename: '/',
  defaultPath: '/panel/dashboard',
  fontFamily: `'Roboto', Helvetica, sans-serif`,
  borderRadius: 12,
  siteInfo: {
    chat_link: '',
    display_in_currency: true,
    email_verification: false,
    footer_html: '',
    github_client_id: '',
    github_oauth: false,
    logo: '',
    quota_per_unit: 500000,
    server_address: '',
    start_time: 0,
    system_name: 'One API',
    top_up_link: '',
    turnstile_check: false,
    turnstile_site_key: '',
    version: '',
    wechat_login: false,
    wechat_qrcode: '',
    lark_login: false,
    lark_client_id: '',
    telegram_bot: '',
    isLoading: true,  // 添加加载状态
  }
};

export default config;
