import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Card,
  Checkbox,
  Dropdown,
  Form,
  Header,
  Icon,
  Image,
  Label,
  Loader,
  Message,
  Segment,
} from 'semantic-ui-react';
import { API, showError, showNotice, showSuccess } from '../../helpers';

const SESSION_STORAGE_KEY = 'playground_session_id';

// 把界面消息结构转换为 API 需要的 content 片段数组
const toApiMessage = (msg) => ({
  role: msg.role,
  content: msg.parts.map((part) =>
    part.type === 'text'
      ? { type: 'text', text: part.text }
      : { type: 'image_url', image_url: { url: part.url } }
  ),
});

const Playground = () => {
  const { t } = useTranslation();
  const [config, setConfig] = useState(null);
  const [configLoading, setConfigLoading] = useState(true);
  const [model, setModel] = useState('');
  const [group, setGroup] = useState('');
  const [searchEnabled, setSearchEnabled] = useState(false);
  const [sessionId, setSessionId] = useState('');
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [images, setImages] = useState([]); // {name, url}
  const [sending, setSending] = useState(false);
  const [uploading, setUploading] = useState(false);
  const messagesEndRef = useRef(null);
  const fileInputRef = useRef(null);

  useEffect(() => {
    loadConfig().then();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  };

  const loadConfig = async () => {
    try {
      const res = await API.get('/api/playground/config');
      const { success, data } = res.data;
      if (success && data && data.enabled) {
        setConfig(data);
        setModel(data.models[0] || '');
        setGroup(data.groups[0] || 'default');
        setSearchEnabled(!!data.search_provider);
        // 尝试恢复上次的临时会话
        const storedId = localStorage.getItem(SESSION_STORAGE_KEY);
        if (storedId) {
          await restoreSession(storedId, data);
        }
      } else {
        setConfig(data || { enabled: false });
      }
    } catch (error) {
      showError(error.message || '无法连接服务器');
      setConfig({ enabled: false });
    }
    setConfigLoading(false);
  };

  const restoreSession = async (id, cfg) => {
    try {
      const res = await API.get(`/api/playground/session/${id}`);
      const { success, data } = res.data;
      if (success && data) {
        setSessionId(data.id);
        setModel(data.model || (cfg && cfg.models[0]) || '');
        setGroup(data.group || (cfg && cfg.groups[0]) || 'default');
        setMessages(
          (data.messages || []).map((m) => ({
            role: m.role,
            parts: (m.content || []).map((part) =>
              part.type === 'image_url'
                ? { type: 'image_url', url: part.image_url?.url || '' }
                : { type: 'text', text: part.text || '' }
            ),
          }))
        );
        showSuccess(t('playground.session_restored'));
      } else {
        localStorage.removeItem(SESSION_STORAGE_KEY);
      }
    } catch (error) {
      localStorage.removeItem(SESSION_STORAGE_KEY);
    }
  };

  const ensureSession = async () => {
    if (sessionId) return sessionId;
    const res = await API.post('/api/playground/session', { model, group });
    const { success, message, data } = res.data;
    if (!success) {
      throw new Error(message || '创建会话失败');
    }
    setSessionId(data.id);
    localStorage.setItem(SESSION_STORAGE_KEY, data.id);
    return data.id;
  };

  const newSession = async () => {
    if (sessionId) {
      try {
        await API.delete(`/api/playground/session/${sessionId}`);
      } catch (error) {
        // 忽略销毁失败（可能已过期）
      }
    }
    setSessionId('');
    setMessages([]);
    localStorage.removeItem(SESSION_STORAGE_KEY);
  };

  const uploadImages = async (e) => {
    const files = Array.from(e.target.files || []);
    e.target.value = '';
    if (files.length === 0) return;
    const maxSize = config?.limits?.max_image_size_mb || 5;
    const accepted = [];
    for (const file of files) {
      if (file.size > maxSize * 1024 * 1024) {
        showError(t('playground.image_too_large', { n: maxSize }));
        continue;
      }
      accepted.push(file);
    }
    if (accepted.length === 0) return;
    setUploading(true);
    for (const file of accepted) {
      try {
        const formData = new FormData();
        formData.append('file', file);
        const res = await API.post('/api/playground/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
        const { success, message, data } = res.data;
        if (success) {
          setImages((prev) => [...prev, { name: data.name, url: data.url }]);
        } else {
          showError(message || t('playground.upload_failed'));
        }
      } catch (error) {
        showError(t('playground.upload_failed'));
      }
    }
    setUploading(false);
  };

  const send = async () => {
    const text = input.trim();
    if ((!text && images.length === 0) || sending) return;

    const userParts = [];
    if (text) userParts.push({ type: 'text', text });
    images.forEach((img) => userParts.push({ type: 'image_url', url: img.url }));

    const history = [...messages, { role: 'user', parts: userParts }];
    setMessages([...history, { role: 'assistant', parts: [], streaming: true }]);
    setInput('');
    setImages([]);
    setSending(true);

    let sid = sessionId;
    try {
      sid = await ensureSession();
    } catch (error) {
      setMessages(history);
      setSending(false);
      showError(error.message);
      return;
    }

    const updateAssistant = (patch) => {
      setMessages((prev) => {
        const next = [...prev];
        const last = next[next.length - 1];
        if (!last || last.role !== 'assistant') return prev;
        next[next.length - 1] = { ...last, ...patch };
        return next;
      });
    };

    try {
      const resp = await fetch('/api/playground/chat', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: sid,
          model,
          group,
          messages: history.map(toApiMessage),
          search: { enabled: searchEnabled },
          stream: true,
        }),
      });
      if (!resp.ok) {
        throw new Error(`HTTP ${resp.status}`);
      }

      const reader = resp.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let assistantText = '';
      let searchResults = null;
      let searchError = '';

      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        let idx;
        while ((idx = buffer.indexOf('\n\n')) >= 0) {
          const rawEvent = buffer.slice(0, idx);
          buffer = buffer.slice(idx + 2);
          for (const line of rawEvent.split('\n')) {
            if (!line.startsWith('data:')) continue;
            const payload = line.slice(5).trim();
            if (!payload || payload === '[DONE]') continue;
            let chunk;
            try {
              chunk = JSON.parse(payload);
            } catch (e) {
              continue;
            }
            if (chunk.type === 'meta') {
              if (chunk.content && chunk.content !== sid) {
                sid = chunk.content;
                setSessionId(sid);
                localStorage.setItem(SESSION_STORAGE_KEY, sid);
              }
            } else if (chunk.type === 'search_results') {
              searchResults = chunk.results || [];
              updateAssistant({ searchResults });
            } else if (chunk.type === 'search_error') {
              searchError = chunk.error || '';
              updateAssistant({ searchError });
            } else if (chunk.type === 'error') {
              updateAssistant({
                streaming: false,
                error: chunk.error || t('playground.stream_interrupted'),
              });
            } else if (chunk.type === 'done') {
              // 流结束
            } else {
              const delta = chunk.choices?.[0]?.delta?.content;
              if (typeof delta === 'string') {
                assistantText += delta;
                updateAssistant({
                  parts: [{ type: 'text', text: assistantText }],
                });
              }
            }
          }
        }
      }

      if (searchError) {
        showNotice(
          t('playground.search_failed', { msg: searchError })
        );
      }
      updateAssistant({
        streaming: false,
        parts: assistantText
          ? [{ type: 'text', text: assistantText }]
          : [{ type: 'text', text: t('playground.stream_interrupted') }],
        searchResults: searchResults || undefined,
      });
    } catch (error) {
      updateAssistant({
        streaming: false,
        error: error.message || t('playground.stream_interrupted'),
      });
      showError(error.message || t('playground.stream_interrupted'));
    }
    setSending(false);
  };

  const renderParts = (parts) => {
    const texts = parts
      .filter((p) => p.type === 'text')
      .map((p) => p.text)
      .join('\n');
    const imgs = parts.filter((p) => p.type === 'image_url');
    return (
      <>
        {texts && (
          <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
            {texts}
          </div>
        )}
        {imgs.length > 0 && (
          <Image.Group size='small' style={{ marginTop: texts ? 8 : 0 }}>
            {imgs.map((img, i) => (
              <Image key={i} src={img.url} style={{ maxWidth: 220 }} />
            ))}
          </Image.Group>
        )}
      </>
    );
  };

  const renderMessage = (msg, index) => {
    const isUser = msg.role === 'user';
    return (
      <div key={index} className={`playground-msg ${isUser ? 'user' : 'assistant'}`}>
        <Label basic size='tiny' className='playground-msg-role'>
          {isUser ? t('playground.you') : t('playground.assistant')}
        </Label>
        <div className='playground-msg-body'>
          {renderParts(msg.parts)}
          {msg.streaming && (
            <Loader active inline size='tiny' style={{ marginLeft: 6 }} />
          )}
          {msg.error && (
            <Message error size='mini' style={{ marginTop: 6 }}>
              {t('playground.error_prefix')}
              {msg.error}
            </Message>
          )}
          {msg.searchError && (
            <Message warning size='mini' style={{ marginTop: 6 }}>
              {t('playground.search_failed', { msg: msg.searchError })}
            </Message>
          )}
          {msg.searchResults && msg.searchResults.length > 0 && (
            <Segment secondary className='playground-search-results'>
              <Header as='h5' style={{ marginBottom: 4 }}>
                <Icon name='search' />
                {t('playground.search_results')}
              </Header>
              {msg.searchResults.map((r, i) => (
                <div key={i} className='playground-search-item'>
                  <a href={r.url} target='_blank' rel='noreferrer'>
                    [{i + 1}] {r.title || r.url}
                  </a>
                  {r.snippet && (
                    <div className='playground-search-snippet'>{r.snippet}</div>
                  )}
                </div>
              ))}
            </Segment>
          )}
        </div>
      </div>
    );
  };

  if (configLoading) {
    return (
      <Segment basic textAlign='center' style={{ paddingTop: 60 }}>
        <Loader active inline='centered' size='large' />
      </Segment>
    );
  }

  if (!config || !config.enabled) {
    return (
      <div className='dashboard-container'>
        <Message
          warning
          size='large'
          header={t('playground.disabled')}
          content={t('playground.disabled_hint')}
        />
      </div>
    );
  }

  const modelOptions = (config.models || []).map((m) => ({
    key: m,
    text: m,
    value: m,
  }));
  const groupOptions = (config.groups || []).map((g) => ({
    key: g,
    text: g,
    value: g,
  }));
  const providerConfigured = !!config.search_provider;
  const searchDisabled = !providerConfigured;

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>
            <Icon name='gamepad' />
            {t('playground.title')}
            <span className='playground-free-hint'>
              {t('playground.hint_free')}
            </span>
          </Card.Header>
          <Form>
            <Form.Group widths='equal'>
              <Form.Field
                control={Dropdown}
                selection
                search
                fluid
                label={t('playground.model')}
                options={modelOptions}
                value={model}
                onChange={(_, { value }) => setModel(value)}
                placeholder={config.models?.length ? '' : t('playground.no_models')}
                noResultsMessage={t('playground.no_models')}
              />
              <Form.Field
                control={Dropdown}
                selection
                fluid
                label={t('playground.group')}
                options={groupOptions}
                value={group}
                onChange={(_, { value }) => setGroup(value)}
              />
              <Form.Field>
                <label>{t('playground.search')}</label>
                <Checkbox
                  toggle
                  checked={searchEnabled}
                  disabled={searchDisabled}
                  onChange={(e, { checked }) => setSearchEnabled(checked)}
                  label={
                    searchDisabled ? t('playground.no_provider') : t('playground.search_configured')
                  }
                />
              </Form.Field>
              <Form.Field>
                <label>&nbsp;</label>
                <Button
                  fluid
                  onClick={newSession}
                  loading={sending}
                  disabled={messages.length === 0}
                >
                  <Icon name='plus' />
                  {t('playground.new_session')}
                </Button>
              </Form.Field>
            </Form.Group>
          </Form>

          <div className='playground-messages'>
            {messages.length === 0 && (
              <Segment placeholder basic>
                <Header icon>
                  <Icon name='gamepad' />
                  {t('playground.title')}
                  <Header.Subheader style={{ marginTop: 8 }}>
                    {t('playground.input_placeholder')}
                    <br />
                    {t('playground.session_info', {
                      n: config.limits?.session_ttl_minutes || 120,
                    })}
                  </Header.Subheader>
                </Header>
              </Segment>
            )}
            {messages.map(renderMessage)}
            <div ref={messagesEndRef} />
          </div>

          {images.length > 0 && (
            <Image.Group size='small' className='playground-pending-images'>
              {images.map((img, i) => (
                <div key={i} className='playground-pending-image'>
                  <Image src={img.url} avatar style={{ width: 64, height: 64, objectFit: 'cover' }} />
                  <Button
                    circular
                    basic
                    size='mini'
                    icon='close'
                    title={t('playground.remove_image')}
                    onClick={() =>
                      setImages((prev) => prev.filter((_, j) => j !== i))
                    }
                  />
                </div>
              ))}
            </Image.Group>
          )}

          <Form>
            <Form.TextArea
              rows={3}
              value={input}
              onChange={(e, { value }) => setInput(value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  send().then();
                }
              }}
              placeholder={t('playground.input_placeholder')}
            />
            <Form.Group style={{ marginTop: 8 }}>
              <Button
                primary
                onClick={send}
                loading={sending}
                disabled={sending || (!input.trim() && images.length === 0)}
              >
                <Icon name='send' />
                {sending ? t('playground.sending') : t('playground.send')}
              </Button>
              <Button
                onClick={() => fileInputRef.current?.click()}
                loading={uploading}
                disabled={uploading}
              >
                <Icon name='image' />
                {uploading ? t('playground.uploading') : t('playground.attach')}
              </Button>
            </Form.Group>
            <input
              ref={fileInputRef}
              type='file'
              accept='image/png,image/jpeg,image/gif,image/webp'
              multiple
              style={{ display: 'none' }}
              onChange={uploadImages}
            />
          </Form>
        </Card.Content>
      </Card>
    </div>
  );
};

export default Playground;
