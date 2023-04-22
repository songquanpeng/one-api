import React from 'react';
import { Segment, Header } from 'semantic-ui-react';

const NotFound = () => (
  <>
    <Header
      block
      as="h4"
      content="404"
      attached="top"
      icon="info"
      className="small-icon"
    />
    <Segment attached="bottom">
      未找到所请求的页面
    </Segment>
  </>
);

export default NotFound;
