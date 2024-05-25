export const CHAT_LINKS = [
  {
    name: 'ChatGPT Next',
    url: 'https://app.nextchat.dev/#/?settings={"key":"{key}","url":"{server}"}',
    show: true,
    sort: 1
  },
  {
    name: 'chatgpt-web-midjourney-proxy',
    url: 'https://vercel.ddaiai.com/#/?settings={"key":"{key}","url":"{server}"}',
    show: true,
    sort: 2
  },
  {
    name: 'AMA 问天',
    url: 'ama://set-api-key?server={server}&key={key}',
    show: false,
    sort: 3
  },
  {
    name: 'OpenCat',
    url: 'opencat://team/join?domain={server}&token={key}',
    show: false,
    sort: 4
  }
];
