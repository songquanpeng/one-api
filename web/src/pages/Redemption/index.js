import React from 'react';
import { Segment, Header } from 'semantic-ui-react';
import RedemptionsTable from '../../components/RedemptionsTable';

const Redemption = () => (
  <>
    <Segment>
      <Header as='h3'>Manage redemption code</Header>
      <RedemptionsTable/>
    </Segment>
  </>
);

export default Redemption;
