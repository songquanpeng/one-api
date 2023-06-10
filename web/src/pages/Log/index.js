import React from 'react';
import { Header, Segment } from 'semantic-ui-react';
import LogsTable from '../../components/LogsTable';

const Token = () => (
  <>
    <Segment>
      <Header as='h3'>额度明细</Header>
      <LogsTable />
    </Segment>
  </>
);

export default Token;
