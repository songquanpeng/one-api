import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

export default function Jump() {
  const location = useLocation();
  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const jump = params.get('url');
    const allowedUrls = ['opencat://', 'ama://'];
    if (jump && allowedUrls.some((url) => jump.startsWith(url))) {
      window.location.href = jump;
    }
  }, [location]);

  return <div>正在跳转中...</div>;
}
