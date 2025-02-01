import React from 'react';
import { Card } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import LogsTable from '../../components/LogsTable';

const Log = () => {
  const { t } = useTranslation();
  
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>{t('log.title')}</Card.Header>
          <LogsTable />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Log;
