import fs from 'fs';
import path from 'path';
import VideoTask from './index';
import VideoTaskDetail from './VideoTaskDetail';
import en from '../../locales/en/translation.json';
import zh from '../../locales/zh/translation.json';

const srcRoot = path.resolve(__dirname, '../..');

describe('video task admin frontend contract', () => {
  test('exports list and detail pages', () => {
    expect(VideoTask).toBeDefined();
    expect(VideoTaskDetail).toBeDefined();
  });

  test('registers admin routes and header navigation', () => {
    const app = fs.readFileSync(path.join(srcRoot, 'App.js'), 'utf8');
    const header = fs.readFileSync(
      path.join(srcRoot, 'components/Header.js'),
      'utf8'
    );

    expect(app).toContain("path='/video-task'");
    expect(app).toContain("path='/video-task/:id'");
    expect(header).toContain("name: 'header.video_task'");
    expect(header).toContain("to: '/video-task'");
    expect(header).toContain("icon: 'film'");
    expect(header).toContain('admin: true');
  });

  test('ships required translations', () => {
    expect(en.header.video_task).toBeTruthy();
    expect(zh.header.video_task).toBeTruthy();
    expect(en.video_task.title).toBeTruthy();
    expect(zh.video_task.title).toBeTruthy();
    expect(en.video_task.messages.create_success).toBeTruthy();
    expect(zh.video_task.messages.create_success).toBeTruthy();
  });
});
