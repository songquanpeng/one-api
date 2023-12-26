import { isAdmin } from 'utils/common';
import { useNavigate } from 'react-router-dom';

const useAuth = () => {
  const userIsAdmin = isAdmin();
  const navigate = useNavigate();

  if (!userIsAdmin) {
    navigate('/panel/404');
  }
};

export default useAuth;
