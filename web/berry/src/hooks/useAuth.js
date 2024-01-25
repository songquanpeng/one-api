import { isAdmin } from 'utils/common';
import { useNavigate } from 'react-router-dom';
const navigate = useNavigate();

const useAuth = () => {
  const userIsAdmin = isAdmin();

  if (!userIsAdmin) {
    navigate('/panel/404');
  }
};

export default useAuth;
