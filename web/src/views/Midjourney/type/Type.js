export const ACTION_TYPE = {
  IMAGINE: { value: 'IMAGINE', text: '绘图', color: 'primary' },
  UPSCALE: { value: 'UPSCALE', text: '放大', color: 'orange' },
  VARIATION: { value: 'VARIATION', text: '变换', color: 'default' },
  HIGH_VARIATION: { value: 'HIGH_VARIATION', text: '强变换', color: 'default' },
  LOW_VARIATION: { value: 'LOW_VARIATION', text: '弱变换', color: 'default' },
  PAN: { value: 'PAN', text: '平移', color: 'secondary' },
  DESCRIBE: { value: 'DESCRIBE', text: '图生文', color: 'secondary' },
  BLEND: { value: 'BLEND', text: '图混合', color: 'secondary' },
  SHORTEN: { value: 'SHORTEN', text: '缩词', color: 'secondary' },
  REROLL: { value: 'REROLL', text: '重绘', color: 'secondary' },
  INPAINT: { value: 'INPAINT', text: '局部重绘-提交', color: 'secondary' },
  ZOOM: { value: 'ZOOM', text: '变焦', color: 'secondary' },
  CUSTOM_ZOOM: { value: 'CUSTOM_ZOOM', text: '自定义变焦-提交', color: 'secondary' },
  MODAL: { value: 'MODAL', text: '窗口处理', color: 'secondary' },
  SWAP_FACE: { value: 'SWAP_FACE', text: '换脸', color: 'secondary' }
};

export const CODE_TYPE = {
  1: { value: 1, text: '已提交', color: 'primary' },
  21: { value: 21, text: '等待中', color: 'orange' },
  22: { value: 22, text: '重复提交', color: 'default' },
  0: { value: 0, text: '未提交', color: 'default' }
};

export const STATUS_TYPE = {
  SUCCESS: { value: 'SUCCESS', text: '成功', color: 'success' },
  NOT_START: { value: 'NOT_START', text: '未启动', color: 'default' },
  SUBMITTED: { value: 'SUBMITTED', text: '队列中', color: 'secondary' },
  IN_PROGRESS: { value: 'IN_PROGRESS', text: '执行中', color: 'primary' },
  FAILURE: { value: 'FAILURE', text: '失败', color: 'orange' },
  MODAL: { value: 'MODAL', text: '窗口等待', color: 'default' }
};
