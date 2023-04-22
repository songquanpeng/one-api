import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Header,
  Icon,
  Pagination,
  Popup,
  Progress,
  Segment,
  Table,
} from 'semantic-ui-react';
import { API, copy, showError, showSuccess } from '../helpers';
import { useDropzone } from 'react-dropzone';
import { ITEMS_PER_PAGE } from '../constants';

const FilesTable = () => {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const { acceptedFiles, getRootProps, getInputProps } = useDropzone();
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState('0');

  const loadFiles = async (startIdx) => {
    const res = await API.get(`/api/file/?p=${startIdx}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setFiles(data);
      } else {
        let newFiles = files;
        newFiles.push(...data);
        setFiles(newFiles);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(files.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadFiles(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  useEffect(() => {
    loadFiles(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const downloadFile = (link, filename) => {
    let linkElement = document.createElement('a');
    linkElement.download = filename;
    linkElement.href = '/upload/' + link;
    linkElement.click();
  };

  const copyLink = (link) => {
    let url = window.location.origin + '/upload/' + link;
    copy(url).then();
    showSuccess('链接已复制到剪贴板');
  };

  const deleteFile = async (id, idx) => {
    const res = await API.delete(`/api/file/${id}`);
    const { success, message } = res.data;
    if (success) {
      let newFiles = [...files];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      newFiles[realIdx].deleted = true;
      // newFiles.splice(idx, 1);
      setFiles(newFiles);
      showSuccess('文件已删除！');
    } else {
      showError(message);
    }
  };

  const searchFiles = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadFiles(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/file/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setFiles(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortFile = (key) => {
    if (files.length === 0) return;
    setLoading(true);
    let sortedUsers = [...files];
    sortedUsers.sort((a, b) => {
      return ('' + a[key]).localeCompare(b[key]);
    });
    if (sortedUsers[0].id === files[0].id) {
      sortedUsers.reverse();
    }
    setFiles(sortedUsers);
    setLoading(false);
  };

  const uploadFiles = async () => {
    if (acceptedFiles.length === 0) return;
    setUploading(true);
    let formData = new FormData();
    for (let i = 0; i < acceptedFiles.length; i++) {
      formData.append('file', acceptedFiles[i]);
    }
    const res = await API.post(`/api/file`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (e) => {
        let uploadProgress = ((e.loaded / e.total) * 100).toFixed(2);
        setUploadProgress(uploadProgress);
      },
    });
    const { success, message } = res.data;
    if (success) {
      showSuccess(`${acceptedFiles.length} 个文件上传成功！`);
    } else {
      showError(message);
    }
    setUploading(false);
    setUploadProgress('0');
    setSearchKeyword('');
    loadFiles(0).then();
    setActivePage(1);
  };

  useEffect(() => {
    uploadFiles().then();
  }, [acceptedFiles]);

  return (
    <>
      <Segment
        placeholder
        {...getRootProps({ className: 'dropzone' })}
        loading={uploading || loading}
        style={{ cursor: 'pointer' }}
      >
        <Header icon>
          <Icon name='file outline' />
          拖拽上传或点击上传
          <input {...getInputProps()} />
        </Header>
      </Segment>
      {uploading ? (
        <Progress
          percent={uploadProgress}
          success
          progress='percent'
        ></Progress>
      ) : (
        <></>
      )}
      <Form onSubmit={searchFiles}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder='搜索文件的名称，上传者以及描述信息 ...'
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>

      <Table basic>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortFile('filename');
              }}
            >
              文件名
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortFile('uploader_id');
              }}
            >
              上传者
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortFile('email');
              }}
            >
              上传时间
            </Table.HeaderCell>
            <Table.HeaderCell>操作</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {files
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((file, idx) => {
              if (file.deleted) return <></>;
              return (
                <Table.Row key={file.id}>
                  <Table.Cell>
                    <a href={'/upload/' + file.link} target='_blank'>
                      {file.filename}
                    </a>
                  </Table.Cell>
                  <Popup
                    content={'上传者 ID：' + file.uploader_id}
                    trigger={<Table.Cell>{file.uploader}</Table.Cell>}
                  />
                  <Table.Cell>{file.upload_time}</Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button
                        size={'small'}
                        positive
                        onClick={() => {
                          downloadFile(file.link, file.filename);
                        }}
                      >
                        下载
                      </Button>
                      <Button
                        size={'small'}
                        negative
                        onClick={() => {
                          deleteFile(file.id, idx).then();
                        }}
                      >
                        删除
                      </Button>
                      <Button
                        size={'small'}
                        onClick={() => {
                          copyLink(file.link);
                        }}
                      >
                        复制链接
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='6'>
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='small'
                siblingRange={1}
                totalPages={
                  Math.ceil(files.length / ITEMS_PER_PAGE) +
                  (files.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default FilesTable;
