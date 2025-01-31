import React from 'react';
import { Card, Tab } from 'semantic-ui-react';
import SystemSetting from '../../components/SystemSetting';
import { isRoot } from '../../helpers';
import OtherSetting from '../../components/OtherSetting';
import PersonalSetting from '../../components/PersonalSetting';
import OperationSetting from '../../components/OperationSetting';

const Setting = () => {
  let panes = [
    {
      menuItem: '个人设置',
      render: () => (
        <Tab.Pane attached={false}>
          <PersonalSetting />
        </Tab.Pane>
      ),
    },
  ];

  if (isRoot()) {
    panes.push({
      menuItem: '运营设置',
      render: () => (
        <Tab.Pane attached={false}>
          <OperationSetting />
        </Tab.Pane>
      ),
    });
    panes.push({
      menuItem: '系统设置',
      render: () => (
        <Tab.Pane attached={false}>
          <SystemSetting />
        </Tab.Pane>
      ),
    });
    panes.push({
      menuItem: '其他设置',
      render: () => (
        <Tab.Pane attached={false}>
          <OtherSetting />
        </Tab.Pane>
      ),
    });
  }

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>系统设置</Card.Header>
          <Tab
            menu={{
              secondary: true,
              pointing: true,
              className: 'settings-tab', // 添加自定义类名以便样式化
            }}
            panes={panes}
          />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Setting;
