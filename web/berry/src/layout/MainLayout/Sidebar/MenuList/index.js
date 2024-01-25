// material-ui
import { Typography } from '@mui/material';

// project imports
import NavGroup from './NavGroup';
import menuItem from 'menu-items';
import { isAdmin } from 'utils/common';

// ==============================|| SIDEBAR MENU LIST ||============================== //
const MenuList = () => {
  const userIsAdmin = isAdmin();

  return (
    <>
      {menuItem.items.map((item) => {
        if (item.type !== 'group') {
          return (
            <Typography key={item.id} variant="h6" color="error" align="center">
              Menu Items Error
            </Typography>
          );
        }

        const filteredChildren = item.children.filter((child) => !child.isAdmin || userIsAdmin);

        if (filteredChildren.length === 0) {
          return null;
        }

        return <NavGroup key={item.id} item={{ ...item, children: filteredChildren }} />;
      })}
    </>
  );
};

export default MenuList;
