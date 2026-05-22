import React from 'react';
import { Card } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import VideoTasksTable from './VideoTasksTable';

const VideoTask = () => {
  const { t } = useTranslation();

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>{t('video_task.title')}</Card.Header>
          <VideoTasksTable />
        </Card.Content>
      </Card>
    </div>
  );
};

export default VideoTask;
