import React from 'react';
import { Header, Segment } from 'semantic-ui-react';
import ChannelsTable from '../../components/ChannelsTable';

const File = () => (
  <>
    <Segment>
      <Header as='h3'>Channel Management</Header>
      <ChannelsTable />
    </Segment>
  </>
);

export default File;
