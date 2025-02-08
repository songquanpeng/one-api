import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Tab } from 'semantic-ui-react';
import SystemSetting from '../../components/SystemSetting';
import { isRoot } from '../../helpers';
import OtherSetting from '../../components/OtherSetting';
import PersonalSetting from '../../components/PersonalSetting';
import OperationSetting from '../../components/OperationSetting';

const Setting = () => {
  const { t } = useTranslation();

  let panes = [
    {
      menuItem: t('setting.tabs.personal'),
      render: () => (
        <Tab.Pane attached={false}>
          <PersonalSetting />
        </Tab.Pane>
      ),
    },
  ];

  if (isRoot()) {
    panes.push({
      menuItem: t('setting.tabs.operation'),
      render: () => (
        <Tab.Pane attached={false}>
          <OperationSetting />
        </Tab.Pane>
      ),
    });
    panes.push({
      menuItem: t('setting.tabs.system'),
      render: () => (
        <Tab.Pane attached={false}>
          <SystemSetting />
        </Tab.Pane>
      ),
    });
    panes.push({
      menuItem: t('setting.tabs.other'),
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
          <Card.Header className='header'>{t('setting.title')}</Card.Header>
          <Tab
            menu={{
              secondary: true,
              pointing: true,
              className: 'settings-tab',
            }}
            panes={panes}
          />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Setting;
