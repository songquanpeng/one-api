import React from 'react';
import { Header, Segment } from 'semantic-ui-react';
import FilesTable from '../../components/FilesTable';

const File = () => (
  <>
    <Segment>
      <Header as='h3'>管理文件</Header>
      <FilesTable />
    </Segment>
  </>
);

export default File;
