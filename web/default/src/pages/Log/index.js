import React from 'react';
import { Card } from 'semantic-ui-react';
import LogsTable from '../../components/LogsTable';

const Log = () => (
  <div className='dashboard-container'>
    <Card fluid className='chart-card'>
      <Card.Content>
        {/*<Card.Header className='header'>操作日志</Card.Header>*/}
        <LogsTable />
      </Card.Content>
    </Card>
  </div>
);

export default Log;
