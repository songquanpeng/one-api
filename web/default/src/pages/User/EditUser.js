import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Form, Card } from 'semantic-ui-react';
import { useParams, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';
import { renderQuota, renderQuotaWithPrompt } from '../../helpers/render';

const EditUser = () => {
  const { t } = useTranslation();
  const params = useParams();
  const userId = params.id;
  const [loading, setLoading] = useState(true);
  const [inputs, setInputs] = useState({
    username: '',
    display_name: '',
    password: '',
    github_id: '',
    wechat_id: '',
    email: '',
    quota: 0,
    group: 'default',
  });
  const [groupOptions, setGroupOptions] = useState([]);
  const {
    username,
    display_name,
    password,
    github_id,
    wechat_id,
    email,
    quota,
    group,
  } = inputs;
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };
  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(
        res.data.data.map((group) => ({
          key: group,
          text: group,
          value: group,
        }))
      );
    } catch (error) {
      showError(error.message);
    }
  };
  const navigate = useNavigate();
  const handleCancel = () => {
    navigate('/setting');
  };
  const loadUser = async () => {
    let res = undefined;
    if (userId) {
      res = await API.get(`/api/user/${userId}`);
    } else {
      res = await API.get(`/api/user/self`);
    }
    const { success, message, data } = res.data;
    if (success) {
      data.password = '';
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };
  useEffect(() => {
    loadUser().then();
    if (userId) {
      fetchGroups().then();
    }
  }, []);

  const submit = async () => {
    let res = undefined;
    if (userId) {
      let data = { ...inputs, id: parseInt(userId) };
      if (typeof data.quota === 'string') {
        data.quota = parseInt(data.quota);
      }
      res = await API.put(`/api/user/`, data);
    } else {
      res = await API.put(`/api/user/self`, inputs);
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('user.messages.update_success'));
    } else {
      showError(message);
    }
  };

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>{t('user.edit.title')}</Card.Header>
          <Form loading={loading} autoComplete='new-password'>
            <Form.Field>
              <Form.Input
                label={t('user.edit.username')}
                name='username'
                placeholder={t('user.edit.username_placeholder')}
                onChange={handleInputChange}
                value={username}
                autoComplete='new-password'
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                label={t('user.edit.password')}
                name='password'
                type={'password'}
                placeholder={t('user.edit.password_placeholder')}
                onChange={handleInputChange}
                value={password}
                autoComplete='new-password'
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                label={t('user.edit.display_name')}
                name='display_name'
                placeholder={t('user.edit.display_name_placeholder')}
                onChange={handleInputChange}
                value={display_name}
                autoComplete='new-password'
              />
            </Form.Field>
            {userId && (
              <>
                <Form.Field>
                  <Form.Dropdown
                    label={t('user.edit.group')}
                    placeholder={t('user.edit.group_placeholder')}
                    name='group'
                    fluid
                    search
                    selection
                    allowAdditions
                    additionLabel={t('user.edit.group_addition')}
                    onChange={handleInputChange}
                    value={inputs.group}
                    autoComplete='new-password'
                    options={groupOptions}
                  />
                </Form.Field>
                <Form.Field>
                  <Form.Input
                    label={`${t('user.edit.quota')}${renderQuotaWithPrompt(
                      quota,
                      t
                    )}`}
                    name='quota'
                    placeholder={t('user.edit.quota_placeholder')}
                    onChange={handleInputChange}
                    value={quota}
                    type={'number'}
                    autoComplete='new-password'
                  />
                </Form.Field>
              </>
            )}
            <Form.Field>
              <Form.Input
                label={t('user.edit.github_id')}
                name='github_id'
                value={github_id}
                autoComplete='new-password'
                placeholder={t('user.edit.github_id_placeholder')}
                readOnly
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                label={t('user.edit.wechat_id')}
                name='wechat_id'
                value={wechat_id}
                autoComplete='new-password'
                placeholder={t('user.edit.wechat_id_placeholder')}
                readOnly
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                label={t('user.edit.email')}
                name='email'
                value={email}
                autoComplete='new-password'
                placeholder={t('user.edit.email_placeholder')}
                readOnly
              />
            </Form.Field>
            <Button onClick={handleCancel}>
              {t('user.edit.buttons.cancel')}
            </Button>
            <Button positive onClick={submit}>
              {t('user.edit.buttons.submit')}
            </Button>
          </Form>
        </Card.Content>
      </Card>
    </div>
  );
};

export default EditUser;
