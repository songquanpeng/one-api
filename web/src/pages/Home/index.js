import React, {useContext, useEffect, useState} from 'react';
import {Card, Grid, Header, Segment} from 'semantic-ui-react';
import {API, showError, showNotice, timestamp2string} from '../../helpers';
import {StatusContext} from '../../context/Status';
import {marked} from 'marked';

const Home = () => {
    const [statusState, statusDispatch] = useContext(StatusContext);
    const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
    const [homePageContent, setHomePageContent] = useState('');

    const displayNotice = async () => {
        const res = await API.get('/api/notice');
        const {success, message, data} = res.data;
        if (success) {
            let oldNotice = localStorage.getItem('notice');
            if (data !== oldNotice && data !== '') {
                showNotice(data);
                localStorage.setItem('notice', data);
            }
        } else {
            showError(message);
        }
    };

    const displayHomePageContent = async () => {
        setHomePageContent(localStorage.getItem('home_page_content') || '');
        const res = await API.get('/api/home_page_content');
        const {success, message, data} = res.data;
        if (success) {
            let content = data;
            if (!data.startsWith('https://')) {
                content = marked.parse(data);
            }
            setHomePageContent(content);
            localStorage.setItem('home_page_content', content);
        } else {
            showError(message);
            setHomePageContent('加载首页内容失败...');
        }
        setHomePageContentLoaded(true);
    };

    const getStartTimeString = () => {
        const timestamp = statusState?.status?.start_time;
        return timestamp2string(timestamp);
    };

    useEffect(() => {
        displayNotice().then();
        displayHomePageContent().then();
    }, []);
    return (
        <>
            {
                // homePageContentLoaded && homePageContent === '' ?
                <>
                    <Segment>
                        <Header as='h3'>当前状态</Header>
                        <Grid columns={2} stackable>
                            <Grid.Column>
                                <Card fluid>
                                    <Card.Content>
                                        <Card.Header>GPT-3.5</Card.Header>
                                        <Card.Meta>信息总览</Card.Meta>
                                        <Card.Description>
                                            <p>通道：官方通道</p>
                                            <p>状态：存活</p>
                                            <p>价格：{statusState?.status?.base_price}R&nbsp;/&nbsp;刀</p>
                                        </Card.Description>
                                    </Card.Content>
                                </Card>
                            </Grid.Column>
                            <Grid.Column>
                                <Card fluid>
                                    <Card.Content>
                                        <Card.Header>GPT-4</Card.Header>
                                        <Card.Meta>信息总览</Card.Meta>
                                        <Card.Description>
                                          <p>通道：官方通道｜低价通道</p>
                                            <p>
                                                状态：
                                                {statusState?.status?.stable_price===-1?
                                                    <span style={{color:'red'}}>不&nbsp;&nbsp;&nbsp;可&nbsp;&nbsp;&nbsp;用</span>
                                                    :
                                                    <span style={{color:'green'}}>可&emsp;&emsp;用</span>
                                                }
                                                ｜
                                                {statusState?.status?.normal_price===-1?
                                                    <span style={{color:'red'}}>不&nbsp;&nbsp;&nbsp;可&nbsp;&nbsp;&nbsp;用</span>
                                                    :
                                                    <span style={{color:'green'}}>可&emsp;&emsp;用</span>
                                                }
                                            </p>
                                            <p>
                                              价格：{statusState?.status?.stable_price}R&nbsp;/&nbsp;刀｜{statusState?.status?.normal_price}R&nbsp;/&nbsp;刀
                                            </p>
                                        </Card.Description>
                                    </Card.Content>
                                </Card>
                            </Grid.Column>
                        </Grid>
                    </Segment>
                  {
                    homePageContent.startsWith('https://') ? <iframe
                        src={homePageContent}
                        style={{ width: '100%', height: '100vh', border: 'none' }}
                    /> : <div style={{ fontSize: 'larger' }} dangerouslySetInnerHTML={{ __html: homePageContent }}></div>
                  }
                </>
                //     :
              //     <>

                // </>
            }

        </>
    );
};

export default Home;
