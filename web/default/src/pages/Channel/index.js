import React from 'react';
import { Card } from 'semantic-ui-react';
import ChannelsTable from '../../components/ChannelsTable';

const Channel = () => (
  <div className='dashboard-container'>
    <Card fluid className='chart-card'>
      <Card.Content>
        <Card.Header className='header'>管理渠道</Card.Header>
        <ChannelsTable />
      </Card.Content>
    </Card>
  </div>
);

export default Channel;
