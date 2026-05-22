import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Button,
  Card,
  Grid,
  Header,
  Label,
  Segment,
  Table,
} from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import {
  API,
  copy,
  removeTrailingSlash,
  showError,
  showSuccess,
  showWarning,
  timestamp2string,
} from '../../helpers';

function isCancelable(status) {
  return status === 'queued' || status === 'running';
}

function isTerminal(status) {
  return status === 'succeeded' || status === 'failed' || status === 'canceled';
}

function getTranslatedStatus(status, t) {
  if (!status) return '-';
  const translation = t(`video_task.status.${status}`);
  if (!translation || translation === `video_task.status.${status}`) {
    return status;
  }
  return translation;
}

function renderStatus(status, t) {
  const colorMap = {
    queued: 'grey',
    running: 'blue',
    succeeded: 'green',
    failed: 'red',
    canceled: 'orange',
  };

  return (
    <Label basic color={colorMap[status] || 'grey'}>
      {getTranslatedStatus(status, t)}
    </Label>
  );
}

function renderTimestamp(timestamp) {
  if (!timestamp) return '-';
  if (typeof timestamp === 'string') return timestamp;
  return timestamp2string(timestamp);
}

function formatJSON(value) {
  if (!value) return '-';
  if (typeof value === 'object') return JSON.stringify(value, null, 2);
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch (e) {
    return value;
  }
}

function getServerAddress() {
  let status = localStorage.getItem('status');
  if (status) {
    try {
      status = JSON.parse(status);
      if (status.server_address) return status.server_address;
    } catch (e) {
      return window.location.origin;
    }
  }
  return window.location.origin;
}

export function shellQuote(value) {
  return `'${String(value).replace(/'/g, "'\\''")}'`;
}

export function buildPublicQueryCurl(serverAddress, taskId) {
  const baseURL = removeTrailingSlash(serverAddress || window.location.origin);
  return `curl ${shellQuote(
    `${baseURL}/v1/videos/tasks/${taskId}`
  )} -H ${shellQuote('Authorization: Bearer <token>')}`;
}

export function shouldAcceptVideoTaskResponse(requestSeq, requestSeqRef) {
  return requestSeq === requestSeqRef.current;
}

export function invalidateVideoTaskRequest(
  requestSeqRef,
  inFlightRef,
  inFlightIdRef
) {
  requestSeqRef.current += 1;
  inFlightRef.current = false;
  inFlightIdRef.current = '';
}

export function translateTimelineStatus(status, t) {
  return getTranslatedStatus(status, t);
}

const VideoTaskDetail = () => {
  const { t } = useTranslation();
  const { id } = useParams();
  const navigate = useNavigate();
  const [task, setTask] = useState(null);
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(false);
  const requestSeqRef = useRef(0);
  const inFlightRef = useRef(false);
  const inFlightIdRef = useRef('');
  const latestTaskRef = useRef(null);
  const taskStatus = task?.status;

  const loadTask = useCallback(async (options = {}) => {
    if (!options.force && inFlightRef.current && inFlightIdRef.current === id) {
      return latestTaskRef.current;
    }
    const requestSeq = requestSeqRef.current + 1;
    requestSeqRef.current = requestSeq;
    inFlightRef.current = true;
    inFlightIdRef.current = id;
    setLoading(true);
    try {
      const res = await API.get(`/api/video_task/${id}`);
      if (!shouldAcceptVideoTaskResponse(requestSeq, requestSeqRef)) {
        return latestTaskRef.current;
      }
      if (!res || !res.data) {
        showError(t('video_task.messages.request_failed'));
        return latestTaskRef.current;
      }
      const { success, message, data } = res.data;
      if (success) {
        const nextTask = data.task;
        latestTaskRef.current = nextTask;
        setTask(nextTask);
        setEvents(data.events || []);
        return nextTask;
      }
      showError(message || t('video_task.messages.request_failed'));
      return latestTaskRef.current;
    } catch (error) {
      if (shouldAcceptVideoTaskResponse(requestSeq, requestSeqRef)) {
        showError(error?.message || t('video_task.messages.request_failed'));
      }
      return latestTaskRef.current;
    } finally {
      if (shouldAcceptVideoTaskResponse(requestSeq, requestSeqRef)) {
        setLoading(false);
        inFlightRef.current = false;
        inFlightIdRef.current = '';
      }
    }
  }, [id, t]);

  useEffect(() => {
    invalidateVideoTaskRequest(requestSeqRef, inFlightRef, inFlightIdRef);
    latestTaskRef.current = null;
    setTask(null);
    setEvents([]);
    loadTask({ force: true }).then();

    return () => {
      invalidateVideoTaskRequest(requestSeqRef, inFlightRef, inFlightIdRef);
    };
  }, [id, loadTask]);

  useEffect(() => {
    let stopped = false;
    let timer;

    const poll = async () => {
      if (stopped) return;
      const currentTask = latestTaskRef.current;
      if (!currentTask || isTerminal(currentTask.status)) return;
      const nextTask = await loadTask();
      if (stopped || !nextTask || isTerminal(nextTask.status)) return;
      timer = setTimeout(poll, 2000);
    };

    if (taskStatus && !isTerminal(taskStatus)) {
      timer = setTimeout(poll, 2000);
    }

    return () => {
      stopped = true;
      if (timer) clearTimeout(timer);
    };
  }, [loadTask, taskStatus]);

  const cancelTask = async () => {
    invalidateVideoTaskRequest(requestSeqRef, inFlightRef, inFlightIdRef);
    setLoading(true);
    try {
      const res = await API.post(`/api/video_task/${id}/cancel`);
      if (!res || !res.data) {
        showError(t('video_task.messages.request_failed'));
        return;
      }
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(t('video_task.messages.cancel_success'));
        if (data) {
          latestTaskRef.current = data;
          setTask(data);
        }
        await loadTask({ force: true });
      } else {
        showError(message || t('video_task.messages.request_failed'));
      }
    } catch (error) {
      showError(error?.message || t('video_task.messages.request_failed'));
    } finally {
      setLoading(false);
    }
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('video_task.messages.copy_success'));
    } else {
      showWarning(t('video_task.messages.copy_failed'));
    }
  };

  const copyQueryCommand = async () => {
    await copyText(buildPublicQueryCurl(getServerAddress(), id));
  };

  if (!task) {
    return (
      <div className='dashboard-container'>
        <Segment loading={loading} />
      </div>
    );
  }

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>
            {t('video_task.detail.title')}
          </Card.Header>
          <Button size='small' onClick={() => copyText(id)}>
            {t('video_task.buttons.copy_task_id')}
          </Button>
          <Button size='small' onClick={copyQueryCommand}>
            {t('video_task.buttons.copy_query')}
          </Button>
          {isCancelable(task.status) && (
            <Button size='small' color='orange' onClick={cancelTask}>
              {t('video_task.buttons.cancel')}
            </Button>
          )}
          <Button size='small' onClick={loadTask} loading={loading}>
            {t('video_task.buttons.refresh')}
          </Button>
          <Button size='small' onClick={() => navigate('/video-task')}>
            {t('video_task.buttons.back')}
          </Button>

          <Grid columns={2} stackable>
            <Grid.Column>
              <Header as='h4'>{t('video_task.detail.basic')}</Header>
              <Table compact basic='very'>
                <Table.Body>
                  <Table.Row>
                    <Table.Cell>{t('video_task.fields.task_id')}</Table.Cell>
                    <Table.Cell>
                      <code>{task.id}</code>
                    </Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>{t('video_task.fields.status')}</Table.Cell>
                    <Table.Cell>{renderStatus(task.status, t)}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>{t('video_task.fields.model')}</Table.Cell>
                    <Table.Cell>{task.model || '-'}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>{t('video_task.fields.user_id')}</Table.Cell>
                    <Table.Cell>{task.user_id || '-'}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>{t('video_task.fields.token_id')}</Table.Cell>
                    <Table.Cell>{task.token_id || '-'}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>{t('video_task.fields.prompt')}</Table.Cell>
                    <Table.Cell style={{ whiteSpace: 'pre-wrap' }}>
                      {task.prompt || '-'}
                    </Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>
                      {t('video_task.fields.input_image_url')}
                    </Table.Cell>
                    <Table.Cell>{task.input_image_url || '-'}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>{t('video_task.fields.provider')}</Table.Cell>
                    <Table.Cell>{task.provider || '-'}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>
                      {t('video_task.fields.provider_job_id')}
                    </Table.Cell>
                    <Table.Cell>{task.provider_job_id || '-'}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell>
                      {t('video_task.fields.billing_status')}
                    </Table.Cell>
                    <Table.Cell>{task.billing_status || '-'}</Table.Cell>
                  </Table.Row>
                </Table.Body>
              </Table>
            </Grid.Column>
            <Grid.Column>
              <Header as='h4'>{t('video_task.detail.timestamps')}</Header>
              <Table compact basic='very'>
                <Table.Body>
                  {[
                    'created_at',
                    'updated_at',
                    'started_at',
                    'completed_at',
                    'canceled_at',
                  ].map((field) => (
                    <Table.Row key={field}>
                      <Table.Cell>{t(`video_task.fields.${field}`)}</Table.Cell>
                      <Table.Cell>{renderTimestamp(task[field])}</Table.Cell>
                    </Table.Row>
                  ))}
                </Table.Body>
              </Table>
            </Grid.Column>
          </Grid>

          {task.error_code && (
            <Segment>
              <Header as='h4'>{t('video_task.detail.error')}</Header>
              <Label basic color='red'>
                {task.error_code}
              </Label>
              <div style={{ marginTop: '8px', whiteSpace: 'pre-wrap' }}>
                {task.error_message || '-'}
              </div>
            </Segment>
          )}

          <Header as='h4'>{t('video_task.detail.request_json')}</Header>
          <pre style={{ overflowX: 'auto' }}>{formatJSON(task.params_json)}</pre>

          <Header as='h4'>{t('video_task.detail.result_json')}</Header>
          <pre style={{ overflowX: 'auto' }}>
            {formatJSON(task.result_json)}
          </pre>

          <Header as='h4'>{t('video_task.detail.provider_response_json')}</Header>
          <pre style={{ overflowX: 'auto' }}>
            {formatJSON(task.provider_response_json)}
          </pre>

          <Header as='h4'>{t('video_task.detail.events')}</Header>
          <Table compact basic='very' size='small'>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>
                  {t('video_task.events.created_at')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('video_task.events.event_type')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('video_task.events.from_status')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('video_task.events.to_status')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('video_task.events.message')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('video_task.events.metadata_json')}
                </Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {events.map((event) => (
                <Table.Row key={event.id}>
                  <Table.Cell>{renderTimestamp(event.created_at)}</Table.Cell>
                  <Table.Cell>{event.event_type || '-'}</Table.Cell>
                  <Table.Cell>
                    {translateTimelineStatus(event.from_status, t)}
                  </Table.Cell>
                  <Table.Cell>
                    {translateTimelineStatus(event.to_status, t)}
                  </Table.Cell>
                  <Table.Cell>{event.message || '-'}</Table.Cell>
                  <Table.Cell>
                    <pre style={{ overflowX: 'auto' }}>
                      {formatJSON(event.metadata_json)}
                    </pre>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        </Card.Content>
      </Card>
    </div>
  );
};

export default VideoTaskDetail;
