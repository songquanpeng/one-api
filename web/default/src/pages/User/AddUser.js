import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../../helpers';

const AddUser = () => {
  const { t } = useTranslation();
  const originInputs = {
    username: '',
    display_name: '',
    password: '',
  };
  const [inputs, setInputs] = useState(originInputs);
  const { username, display_name, password } = inputs;

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const submit = async () => {
    if (inputs.username === '' || inputs.password === '') return;
    const res = await API.post(`/api/user/`, inputs);
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('user.messages.create_success'));
      setInputs(originInputs);
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment>
        <Header as='h3'>{t('user.add.title')}</Header>
        <Form autoComplete='off'>
          <Form.Field>
            <Form.Input
              label={t('user.edit.username')}
              name='username'
              placeholder={t('user.edit.username_placeholder')}
              onChange={handleInputChange}
              value={username}
              autoComplete='off'
              required
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label={t('user.edit.display_name')}
              name='display_name'
              placeholder={t('user.edit.display_name_placeholder')}
              onChange={handleInputChange}
              value={display_name}
              autoComplete='off'
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
              autoComplete='off'
              required
            />
          </Form.Field>
          <Button positive type={'submit'} onClick={submit}>
            {t('user.edit.buttons.submit')}
          </Button>
        </Form>
      </Segment>
    </>
  );
};

export default AddUser;
