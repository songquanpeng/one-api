import React, { useEffect, useState } from "react";
import { API } from 'utils/api';
import { showError, showSuccess } from 'utils/common';

const Chat = () => {
  const [value, setValue] = useState([]);
  const [isLoading, setIsLoading] = useState(true); // 加载状态

  const loadTokens = async () => {
    setIsLoading(true); // 开始加载
    const res = await API.get(`/api/token/?p=0`);
    const { success, message, data } = res.data;
    setValue(data);
    setIsLoading(false); // 加载完成
  };

  useEffect(() => {
    loadTokens();
  }, []);

  if (isLoading) {
    return <div>Loading...</div>;
  } else if (value.length) {
    const siteInfo = JSON.parse(localStorage.getItem("siteInfo"));
    const chatLink = siteInfo.chat_link + `#/?settings={"key":"sk-${value[0]?.key}","url":"${siteInfo.server_address}"}`;
    return (
      <iframe
        src={chatLink}
        style={{ width: "100%", height: "85vh", border: "none" }}
      />
    );
  } else {
    showError("未找到可用令牌，请先创建令牌");
  }
};

export default Chat;