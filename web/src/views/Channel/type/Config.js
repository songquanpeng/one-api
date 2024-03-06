const defaultConfig = {
  input: {
    name: '',
    type: 1,
    key: '',
    base_url: '',
    other: '',
    proxy: '',
    test_model: '',
    model_mapping: '',
    models: [],
    groups: ['default']
  },
  inputLabel: {
    name: '渠道名称',
    type: '渠道类型',
    base_url: '渠道API地址',
    key: '密钥',
    other: '其他参数',
    proxy: '代理地址',
    test_model: '测速模型',
    models: '模型',
    model_mapping: '模型映射关系',
    groups: '用户组'
  },
  prompt: {
    type: '请选择渠道类型',
    name: '请为渠道命名',
    base_url: '可空，请输入中转API地址，例如通过cloudflare中转',
    key: '请输入渠道对应的鉴权密钥',
    other: '',
    proxy: '单独设置代理地址，支持http和socks5，例如：http://127.0.0.1:1080',
    test_model: '用于测试使用的模型，为空时无法测速,如：gpt-3.5-turbo',
    models: '请选择该渠道所支持的模型',
    model_mapping:
      '请输入要修改的模型映射关系，格式为：api请求模型ID:实际转发给渠道的模型ID，使用JSON数组表示，例如：{"gpt-3.5": "gpt-35"}',
    groups: '请选择该渠道所支持的用户组'
  },
  modelGroup: 'OpenAI'
};

const typeConfig = {
  3: {
    inputLabel: {
      base_url: 'AZURE_OPENAI_ENDPOINT',
      other: '默认 API 版本'
    },
    prompt: {
      base_url: '请填写AZURE_OPENAI_ENDPOINT',
      other: '请输入默认API版本，例如：2023-06-01-preview'
    }
  },
  11: {
    input: {
      models: ['PaLM-2'],
      test_model: 'PaLM-2'
    },
    modelGroup: 'Google PaLM'
  },
  14: {
    input: {
      models: ['claude-instant-1.2', 'claude-2.0', 'claude-2.1','claude-3-opus-20240229','claude-3-sonnet-20240229'],
      test_model: 'claude-3-sonnet-20240229'
    },
    modelGroup: 'Anthropic'
  },
  15: {
    input: {
      models: ['ERNIE-Bot', 'ERNIE-Bot-turbo', 'ERNIE-Bot-4', 'Embedding-V1'],
      test_model: 'ERNIE-Bot'
    },
    prompt: {
      key: '按照如下格式输入：APIKey|SecretKey'
    },
    modelGroup: 'Baidu'
  },
  16: {
    input: {
      models: ['chatglm_turbo', 'chatglm_pro', 'chatglm_std', 'chatglm_lite'],
      test_model: 'chatglm_lite'
    },
    modelGroup: 'Zhipu'
  },
  17: {
    inputLabel: {
      other: '插件参数'
    },
    input: {
      models: [
        'qwen-turbo',
        'qwen-plus',
        'qwen-max',
        'qwen-max-longcontext',
        'text-embedding-v1',
        'qwen-turbo-internet',
        'qwen-plus-internet',
        'qwen-max-internet',
        'qwen-max-longcontext-internet'
      ],
      test_model: 'qwen-turbo'
    },
    prompt: {
      other: '请输入插件参数，即 X-DashScope-Plugin 请求头的取值'
    },
    modelGroup: 'Ali'
  },
  18: {
    inputLabel: {
      other: '版本号'
    },
    input: {
      models: ['SparkDesk', 'SparkDesk-v1.1', 'SparkDesk-v2.1', 'SparkDesk-v3.1', 'SparkDesk-v3.5']
    },
    prompt: {
      key: '按照如下格式输入：APPID|APISecret|APIKey',
      other: '请输入版本号，例如：v3.1'
    },
    modelGroup: 'Xunfei'
  },
  19: {
    input: {
      models: ['360GPT_S2_V9', 'embedding-bert-512-v1', 'embedding_s1_v1', 'semantic_similarity_s1_v1'],
      test_model: '360GPT_S2_V9'
    },
    modelGroup: '360'
  },
  22: {
    prompt: {
      key: '按照如下格式输入：APIKey-AppId，例如：fastgpt-0sp2gtvfdgyi4k30jwlgwf1i-64f335d84283f05518e9e041'
    }
  },
  23: {
    input: {
      models: ['hunyuan'],
      test_model: 'hunyuan'
    },
    prompt: {
      key: '按照如下格式输入：AppId|SecretId|SecretKey'
    },
    modelGroup: 'Tencent'
  },
  25: {
    inputLabel: {
      other: '版本号'
    },
    input: {
      models: ['gemini-pro', 'gemini-pro-vision'],
      test_model: 'gemini-pro'
    },
    prompt: {
      other: '请输入版本号，例如：v1'
    },
    modelGroup: 'Google Gemini'
  },
  26: {
    input: {
      models: ['Baichuan2-Turbo', 'Baichuan2-Turbo-192k', 'Baichuan2-53B', 'Baichuan-Text-Embedding'],
      test_model: 'Baichuan2-Turbo'
    },
    modelGroup: 'Baichuan'
  },
  24: {
    input: {
      models: ['tts-1', 'tts-1-hd']
    },
    prompt: {
      test_model: ''
    }
  },
  27: {
    input: {
      models: ['abab5.5-chat', 'abab5.5s-chat', 'abab6-chat', 'embo-01'],
      test_model: 'abab5.5-chat'
    },
    prompt: {
      key: '按照如下格式输入：APISecret|groupID'
    }
  },
  28: {
    input: {
      models: ['deepseek-coder', 'deepseek-chat'],
      test_model: 'deepseek-chat'
    }
  },
  29: {
    input: {
      models: ['moonshot-v1-8k', 'moonshot-v1-32k', 'moonshot-v1-128k'],
      test_model: 'moonshot-v1-8k'
    }
  }
};

export { defaultConfig, typeConfig };
