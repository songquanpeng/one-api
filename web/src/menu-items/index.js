import panel from './panel';

// ==============================|| MENU ITEMS ||============================== //

const menuItems = {
  items: [panel],
  urlMap: {}
};

// Initialize urlMap
menuItems.urlMap = menuItems.items.reduce((map, item) => {
  item.children.forEach((child) => {
    map[child.url] = child;
  });
  return map;
}, {});

export default menuItems;
