import React, { useEffect, useState } from 'react';
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

function renderStatus(status) {
  const colorMap = {
    queued: 'grey',
    running: 'blue',
    succeeded: 'green',
    failed: 'red',
    canceled: 'orange',
  };

  return (
    <Label basic color={colorMap[status] || 'grey'}>
      {status || '-'}
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
    status = JSON.parse(status);
    if (status.server_address) return status.server_address;
  }
  return window.location.origin;
}

const VideoTaskDetail = () => {
  const { t } = useTranslation();
  const { id } = useParams();
  const navigate = useNavigate();
  const [task, setTask] = useState(null);
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(false);

  const loadTask = async () => {
    setLoading(true);
    const res = await API.get(`/api/video_task/${id}`);
    const { success, message, data } = res.data;
    if (success) {
      setTask(data.task);
      setEvents(data.events || []);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadTask().then();
    // Reload when the route id changes; polling below owns subsequent refreshes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  useEffect(() => {
    if (!task || isTerminal(task.status)) return undefined;
    const timer = setInterval(() => {
      loadTask().then();
    }, 2000);
    return () => clearInterval(timer);
    // The interval should reset only when the observed task status changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, task?.status]);

  const cancelTask = async () => {
    const res = await API.post(`/api/video_task/${id}/cancel`);
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(t('video_task.messages.cancel_success'));
      if (data) setTask(data);
      await loadTask();
    } else {
      showError(message);
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
    const command = `curl '${getServerAddress()}/v1/videos/tasks/${id}' -H 'Authorization: Bearer <token>'`;
    await copyText(command);
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
                    <Table.Cell>{renderStatus(task.status)}</Table.Cell>
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
                  <Table.Cell>{event.from_status || '-'}</Table.Cell>
                  <Table.Cell>{event.to_status || '-'}</Table.Cell>
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
