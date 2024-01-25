// material-ui
import logo from 'assets/images/logo.svg';
import { useSelector } from 'react-redux';

/**
 * if you want to use image instead of <svg> uncomment following.
 *
 * import logoDark from 'assets/images/logo-dark.svg';
 * import logo from 'assets/images/logo.svg';
 *
 */

// ==============================|| LOGO SVG ||============================== //

const Logo = () => {
  const siteInfo = useSelector((state) => state.siteInfo);

  return <img src={siteInfo.logo || logo} alt={siteInfo.system_name} height="50" />;
};

export default Logo;
