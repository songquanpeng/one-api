// 从localStorage中获取 headerScript
// 用useEffect将headerScript插入到页面中

import { useEffect, useRef } from 'react';
import { getHeaderScript } from '../helpers';

const useHeaderScript = () => {
  const oldScript = useRef('');
  const headerScript = getHeaderScript();

  useEffect(() => {
    if (!oldScript.current || oldScript.current !== headerScript) {
      if (headerScript) {
        // Insert the code into the header
        document.head.innerHTML += headerScript;
        oldScript.current = headerScript;
      }
    }
  }, [headerScript]);

  return headerScript;
};

export default useHeaderScript;
