import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

export default function Jump() {
  const location = useLocation();

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const jump = params.get('url');
    if (jump) {
      window.location.href = jump;
    }
  }, [location]);

  return <div>正在跳转中...</div>;
}
