const PaymentType = {
  epay: '易支付'
};

const CurrencyType = {
  CNY: '人民币',
  USD: '美元'
};

const PaymentConfig = {
  epay: {
    pay_domain: {
      name: '支付域名',
      description: '支付域名',
      type: 'text',
      value: ''
    },
    partner_id: {
      name: '商户号',
      description: '商户号',
      type: 'text',
      value: ''
    },
    key: {
      name: '密钥',
      description: '密钥',
      type: 'text',
      value: ''
    },
    pay_type: {
      name: '支付类型',
      description: '支付类型,如果需要跳转到易支付收银台,请选择收银台',
      type: 'select',
      value: '',
      options: [
        {
          name: '收银台',
          value: ''
        },
        {
          name: '支付宝',
          value: 'alipay'
        },
        {
          name: '微信',
          value: 'wxpay'
        },
        {
          name: 'QQ',
          value: 'qqpay'
        },
        {
          name: '京东',
          value: 'jdpay'
        },
        {
          name: '银联',
          value: 'bank'
        },
        {
          name: 'Paypal',
          value: 'paypal'
        },
        {
          name: 'USDT',
          value: 'usdt'
        }
      ]
    }
  }
};

export { PaymentConfig, PaymentType, CurrencyType };
