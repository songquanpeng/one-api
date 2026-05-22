import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  Button,
  Dropdown,
  Form,
  Label,
  Modal,
  Pagination,
  Table,
} from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import {
  API,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';

const TEST_MODE_OPTIONS = [
  { key: 'success', text: 'success', value: 'success' },
  { key: 'fail', text: 'fail', value: 'fail' },
  { key: 'slow', text: 'slow', value: 'slow' },
];

const DEFAULT_DEBUG_FORM = {
  prompt: '',
  input_image_url: '',
  model: 'fake-seedance-lite',
  duration: 5,
  aspect_ratio: '16:9',
  resolution: '720p',
  test_mode: 'success',
};

function isCancelable(status) {
  return status === 'queued' || status === 'running';
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

function datetimeToTimestamp(value) {
  if (!value) return '';
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) return '';
  return String(Math.floor(timestamp / 1000));
}

const VideoTasksTable = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [filters, setFilters] = useState({
    status: '',
    model: '',
    user_id: '',
    task_id: '',
    start_timestamp: '',
    end_timestamp: '',
  });
  const [debugForm, setDebugForm] = useState(DEFAULT_DEBUG_FORM);
  const statusOptions = [
    { key: 'all', text: t('video_task.status.all'), value: '' },
    { key: 'queued', text: 'queued', value: 'queued' },
    { key: 'running', text: 'running', value: 'running' },
    { key: 'succeeded', text: 'succeeded', value: 'succeeded' },
    { key: 'failed', text: 'failed', value: 'failed' },
    { key: 'canceled', text: 'canceled', value: 'canceled' },
  ];

  const updateFilter = (name, value) => {
    setFilters((filters) => ({ ...filters, [name]: value }));
  };

  const updateDebugForm = (name, value) => {
    setDebugForm((form) => ({ ...form, [name]: value }));
  };

  const buildQuery = (page) => {
    const params = new URLSearchParams();
    params.set('p', String(page));
    params.set('status', filters.status);
    params.set('model', filters.model.trim());
    params.set('user_id', filters.user_id.trim());
    params.set('task_id', filters.task_id.trim());
    params.set('start_timestamp', datetimeToTimestamp(filters.start_timestamp));
    params.set('end_timestamp', datetimeToTimestamp(filters.end_timestamp));
    return params.toString();
  };

  const loadTasks = async (page) => {
    setLoading(true);
    const res = await API.get(`/api/video_task/?${buildQuery(page)}`);
    const { success, message, data } = res.data;
    if (success) {
      if (page === 0) {
        setTasks(data || []);
      } else {
        let newTasks = [...tasks];
        newTasks.splice(page * ITEMS_PER_PAGE, data.length, ...(data || []));
        setTasks(newTasks);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const refresh = async () => {
    setActivePage(1);
    await loadTasks(0);
  };

  useEffect(() => {
    refresh().then();
    // Follow existing admin tables: filters are applied only by the query button.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(tasks.length / ITEMS_PER_PAGE) + 1) {
        await loadTasks(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  const cancelTask = async (taskId) => {
    const res = await API.post(`/api/video_task/${taskId}/cancel`);
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('video_task.messages.cancel_success'));
      await refresh();
    } else {
      showError(message);
    }
  };

  const createDebugTask = async () => {
    if (!debugForm.prompt.trim()) {
      showError(t('video_task.messages.prompt_required'));
      return;
    }
    setCreating(true);
    const res = await API.post('/api/video_task/debug', {
      ...debugForm,
      prompt: debugForm.prompt.trim(),
      input_image_url: debugForm.input_image_url.trim(),
      duration: Number(debugForm.duration),
    });
    const { success, message, data } = res.data;
    setCreating(false);
    if (success) {
      showSuccess(t('video_task.messages.create_success'));
      setCreateOpen(false);
      setDebugForm(DEFAULT_DEBUG_FORM);
      navigate(`/video-task/${data.id}`);
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Form>
        <Form.Group>
          <Form.Field width={3}>
            <label>{t('video_task.filters.status')}</label>
            <Dropdown
              selection
              options={statusOptions}
              value={filters.status}
              onChange={(_, { value }) => updateFilter('status', value)}
            />
          </Form.Field>
          <Form.Input
            fluid
            label={t('video_task.filters.model')}
            size='small'
            width={3}
            value={filters.model}
            name='model'
            placeholder={t('video_task.filters.optional')}
            onChange={(_, { value }) => updateFilter('model', value)}
          />
          <Form.Input
            fluid
            label={t('video_task.filters.user_id')}
            size='small'
            width={2}
            value={filters.user_id}
            name='user_id'
            placeholder={t('video_task.filters.optional')}
            onChange={(_, { value }) => updateFilter('user_id', value)}
          />
          <Form.Input
            fluid
            label={t('video_task.filters.task_id')}
            size='small'
            width={4}
            value={filters.task_id}
            name='task_id'
            placeholder={t('video_task.filters.optional')}
            onChange={(_, { value }) => updateFilter('task_id', value)}
          />
          <Form.Button
            fluid
            label={t('video_task.buttons.query')}
            size='small'
            width={2}
            onClick={refresh}
            loading={loading}
          >
            {t('video_task.buttons.submit')}
          </Form.Button>
        </Form.Group>
        <Form.Group>
          <Form.Input
            fluid
            label={t('video_task.filters.start_time')}
            size='small'
            width={4}
            value={filters.start_timestamp}
            type='datetime-local'
            name='start_timestamp'
            onChange={(_, { value }) => updateFilter('start_timestamp', value)}
          />
          <Form.Input
            fluid
            label={t('video_task.filters.end_time')}
            size='small'
            width={4}
            value={filters.end_timestamp}
            type='datetime-local'
            name='end_timestamp'
            onChange={(_, { value }) => updateFilter('end_timestamp', value)}
          />
        </Form.Group>
      </Form>

      <Button size='small' primary onClick={() => setCreateOpen(true)}>
        {t('video_task.buttons.create_debug')}
      </Button>
      <Button size='small' onClick={refresh} loading={loading}>
        {t('video_task.buttons.refresh')}
      </Button>

      <Table basic='very' compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>{t('video_task.table.task_id')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.status')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.model')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.user')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.token')}</Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.provider')}</Table.HeaderCell>
            <Table.HeaderCell>
              {t('video_task.table.created_at')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('video_task.table.actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {tasks
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((task) => (
              <Table.Row key={task.id}>
                <Table.Cell>
                  <code>{task.id}</code>
                </Table.Cell>
                <Table.Cell>{renderStatus(task.status)}</Table.Cell>
                <Table.Cell>{task.model || '-'}</Table.Cell>
                <Table.Cell>{task.user_id || '-'}</Table.Cell>
                <Table.Cell>{task.token_id || '-'}</Table.Cell>
                <Table.Cell>{task.provider || '-'}</Table.Cell>
                <Table.Cell>{renderTimestamp(task.created_at)}</Table.Cell>
                <Table.Cell>
                  <Button size='mini' as={Link} to={`/video-task/${task.id}`}>
                    {t('video_task.buttons.view')}
                  </Button>
                  {isCancelable(task.status) && (
                    <Button
                      size='mini'
                      color='orange'
                      onClick={() => cancelTask(task.id)}
                    >
                      {t('video_task.buttons.cancel')}
                    </Button>
                  )}
                </Table.Cell>
              </Table.Row>
            ))}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='8'>
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='small'
                siblingRange={1}
                totalPages={
                  Math.ceil(tasks.length / ITEMS_PER_PAGE) +
                  (tasks.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} size='small'>
        <Modal.Header>{t('video_task.debug.title')}</Modal.Header>
        <Modal.Content>
          <Form>
            <Form.TextArea
              label={t('video_task.debug.prompt')}
              value={debugForm.prompt}
              name='prompt'
              onChange={(_, { value }) => updateDebugForm('prompt', value)}
            />
            <Form.Input
              label={t('video_task.debug.input_image_url')}
              value={debugForm.input_image_url}
              name='input_image_url'
              onChange={(_, { value }) =>
                updateDebugForm('input_image_url', value)
              }
            />
            <Form.Input
              label={t('video_task.debug.model')}
              value={debugForm.model}
              name='model'
              onChange={(_, { value }) => updateDebugForm('model', value)}
            />
            <Form.Group widths='equal'>
              <Form.Input
                label={t('video_task.debug.duration')}
                type='number'
                min='1'
                value={debugForm.duration}
                name='duration'
                onChange={(_, { value }) => updateDebugForm('duration', value)}
              />
              <Form.Input
                label={t('video_task.debug.aspect_ratio')}
                value={debugForm.aspect_ratio}
                name='aspect_ratio'
                onChange={(_, { value }) =>
                  updateDebugForm('aspect_ratio', value)
                }
              />
              <Form.Input
                label={t('video_task.debug.resolution')}
                value={debugForm.resolution}
                name='resolution'
                onChange={(_, { value }) => updateDebugForm('resolution', value)}
              />
            </Form.Group>
            <Form.Field>
              <label>{t('video_task.debug.test_mode')}</label>
              <Dropdown
                selection
                options={TEST_MODE_OPTIONS}
                value={debugForm.test_mode}
                onChange={(_, { value }) => updateDebugForm('test_mode', value)}
              />
            </Form.Field>
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setCreateOpen(false)}>
            {t('video_task.buttons.close')}
          </Button>
          <Button primary onClick={createDebugTask} loading={creating}>
            {t('video_task.buttons.create')}
          </Button>
        </Modal.Actions>
      </Modal>
    </>
  );
};

export default VideoTasksTable;
