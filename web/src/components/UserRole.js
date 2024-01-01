import React from 'react';
import { Label } from 'semantic-ui-react';

const UserRole = ({ role }) => {
  switch (role) {
    case 1:
      return <Label>普通用户</Label>;
    case 10:
      return <Label color='yellow'>管理员</Label>;
    case 100:
      return <Label color='orange'>超级管理员</Label>;
    default:
      return <Label color='red'>未知身份</Label>;
  }
};

export default UserRole;
