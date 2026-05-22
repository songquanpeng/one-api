import fs from 'fs';
import path from 'path';
import { AdminRoute } from '../../App';
import VideoTask from './index';
import VideoTaskDetail, {
  buildPublicQueryCurl,
  invalidateVideoTaskRequest,
  shouldAcceptVideoTaskResponse,
  shellQuote,
  translateTimelineStatus,
} from './VideoTaskDetail';
import {
  getStatusOptions,
  getTestModeOptions,
  getTranslatedStatus,
} from './VideoTasksTable';
import en from '../../locales/en/translation.json';
import zh from '../../locales/zh/translation.json';

const srcRoot = path.resolve(__dirname, '../..');

describe('video task admin frontend contract', () => {
  test('exports list and detail pages', () => {
    expect(VideoTask).toBeDefined();
    expect(VideoTaskDetail).toBeDefined();
    expect(AdminRoute).toBeDefined();
  });

  test('registers admin routes and header navigation', () => {
    const app = fs.readFileSync(path.join(srcRoot, 'App.js'), 'utf8');
    const header = fs.readFileSync(
      path.join(srcRoot, 'components/Header.js'),
      'utf8'
    );

    expect(app).toContain("path='/video-task'");
    expect(app).toContain("path='/video-task/:id'");
    expect(app).toContain('AdminRoute');
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

  test('builds shell-safe public query curl commands', () => {
    expect(shellQuote("https://example.com/a'b")).toBe(
      "'https://example.com/a'\\''b'"
    );
    expect(
      buildPublicQueryCurl("https://api.example.com/root'path", "task_'123")
    ).toBe(
      "curl 'https://api.example.com/root'\\''path/v1/videos/tasks/task_'\\''123' -H 'Authorization: Bearer <token>'"
    );
  });

  test('defines translated status and debug dropdown labels', () => {
    const t = (key) => {
      const keys = key.split('.');
      return keys.reduce((value, part) => value[part], en);
    };

    expect(getTranslatedStatus('queued', t)).toBe('Queued');
    expect(getTranslatedStatus('not-a-status', t)).toBe('not-a-status');
    expect(getStatusOptions(t).map((option) => option.text)).toEqual([
      'All',
      'Queued',
      'Running',
      'Succeeded',
      'Failed',
      'Canceled',
    ]);
    expect(getTestModeOptions(t).map((option) => option.text)).toEqual([
      'Success',
      'Fail',
      'Slow',
    ]);
  });

  test('invalidates stale detail requests before accepting responses', () => {
    const requestSeqRef = { current: 3 };
    const inFlightRef = { current: true };
    const inFlightIdRef = { current: 'task_old' };

    expect(shouldAcceptVideoTaskResponse(3, requestSeqRef)).toBe(true);
    invalidateVideoTaskRequest(requestSeqRef, inFlightRef, inFlightIdRef);

    expect(shouldAcceptVideoTaskResponse(3, requestSeqRef)).toBe(false);
    expect(inFlightRef.current).toBe(false);
    expect(inFlightIdRef.current).toBe('');
  });

  test('translates timeline status labels', () => {
    const t = (key) => {
      const keys = key.split('.');
      return keys.reduce((value, part) => value[part], en);
    };

    expect(translateTimelineStatus('queued', t)).toBe('Queued');
    expect(translateTimelineStatus('', t)).toBe('-');
  });
});
