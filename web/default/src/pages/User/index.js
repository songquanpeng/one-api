import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card } from 'semantic-ui-react';
import UsersTable from '../../components/UsersTable';

const User = () => {
  const { t } = useTranslation();

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>{t('user.title')}</Card.Header>
          <UsersTable />
        </Card.Content>
      </Card>
    </div>
  );
};

export default User;
