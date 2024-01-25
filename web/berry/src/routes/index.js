import { useRoutes } from 'react-router-dom';

// routes
import MainRoutes from './MainRoutes';
import OtherRoutes from './OtherRoutes';

// ==============================|| ROUTING RENDER ||============================== //

export default function ThemeRoutes() {
  return useRoutes([MainRoutes, OtherRoutes]);
}
