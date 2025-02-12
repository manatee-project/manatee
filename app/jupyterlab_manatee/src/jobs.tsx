/**
 * Copyright 2024 TikTok Pte. Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { PanelWithToolbar, ReactWidget} from '@jupyterlab/ui-components';
import { ITranslator, nullTranslator } from '@jupyterlab/translation';
import * as React from 'react';
import { useState, useEffect } from 'react';
import { ServerConnection } from '@jupyterlab/services';
import { showDialog, Dialog } from '@jupyterlab/apputils';
import { ConfigProvider, Table, TableColumnProps, Tag, Button } from '@arco-design/web-react';
import enUS from '@arco-design/web-react/es/locale/en-US';
import "@arco-design/web-react/dist/css/arco.css";

const statusMap: Map<number, {color: string, text: string}> = new Map([
    [1, {color: 'green', text: 'Image Building'}],
    [2, {color: 'red', text: 'Image Building Failed'}],
    [3, {color: 'gray', text: 'VM Waiting'}],
    [4, {color: 'green', text: 'VM Running'}],
    [5, {color: '#86909c', text: 'VM Finished'}],
    [6, {color: 'red', text: 'VM Killed'}],
    [7, {color: 'red', text: 'VM Failed'}],
    [8, {color: 'gray', text: 'VM Other'}],
    [9, {color: 'red', text: 'VM Launch Failed'}]
]);

const columns: TableColumnProps[] = [
    {
        title: 'ID',
        dataIndex: 'id',
        width: 2
    },
    {
        title: 'Jupyter File',
        dataIndex: 'jupyter_file_name',
        ellipsis: true,
        width:  5
    },
    {
        title: 'Status',
        dataIndex: 'job_status',
        ellipsis: true,
        width:  5,
        render: (col, item: any, index) => {
            let status = statusMap.get(col)
            let color = 'gray'
            let text = 'unknown'
            if (status !== undefined) {
                color = status.color
                text = status.text
            }
            return (
                <Tag color={color}>{text}</Tag>
            )
        }
    },
    {
        title: 'Created At',
        dataIndex: 'created_at',
        ellipsis: true,
        width: 5
    },
    {
        title: 'Updated At',
        dataIndex: 'updated_at',
        ellipsis: true,
        width: 5
    },
    {
        title: 'Ouput',
        ellipsis: true,
        width: 5,
        render: (col, item: any, index) => {
            if (item.job_status == 5) {
                return (
                    <Button type='primary' onClick={async (e: Event) => {
                        const id = item.id
                        const request = JSON.stringify({id: id});
                        const settings = ServerConnection.makeSettings();
                        const result = await showDialog({
                            title: "Download the output of the Job?",
                            body: 'Job id: ' + id,
                            buttons: [Dialog.okButton(), Dialog.cancelButton()]
                        });
                        if (result.button.accept) {
                            ServerConnection.makeRequest(settings.baseUrl + "manatee/output", {
                                body: request, method: "POST"
                            }, settings).then(response => {
                                if (response.status !== 200) {
                                    showDialog({
                                        title: "Download Failed",
                                        buttons: [Dialog.okButton(), Dialog.cancelButton()]
                                    });
                                    console.error(response)
                                    return;
                                }
                                response.body?.getReader().read().then(({done, value}) => {
                                    if (done) {
                                        console.error("stream is closed");
                                        return;
                                    }
                                    let decoder = new TextDecoder('utf-8');
                                    let result = JSON.parse(decoder.decode(value))
                                    if (result.code == 0) {
                                        showDialog({
                                            title: "Download Successful",
                                            body: 'Filename: ' + result.filename,
                                            buttons: [Dialog.okButton(), Dialog.cancelButton()]
                                        });
                                    } else {
                                        showDialog({
                                            title: "Download Failed",
                                            body: 'Error: ' + result.msg,
                                            buttons: [Dialog.okButton(), Dialog.cancelButton()]
                                        });
                                    }
                                })
                            });
                        }
                       
                    }}>Download</Button>
                )
            } else {
                return (<div></div>)
            }
        }
    },
    {
        title: 'Attestation',
        ellipsis: true,
        width: 5,
        render: (col, item: any, index) => {
            if (item.job_status == 5) {
                return (
                    <Button type='primary' onClick={async (e: Event) => {
                        const id = item.id
                        const settings = ServerConnection.makeSettings();
                        const result = await showDialog({
                            title: "Get Attestation Report of the Job?",
                            body: 'Job id: ' + id,
                            buttons: [Dialog.okButton(), Dialog.cancelButton()]
                        });
                        let requestUrlWithParams = settings.baseUrl + "manatee/attestation" + "?id=" + id;
                        if (result.button.accept) {
                            ServerConnection.makeRequest(requestUrlWithParams, {
                                method: "GET"
                            }, settings).then(response => {
                                if (response.status !== 200) {
                                    showDialog({
                                        title: "Get Attestation Report Failed",
                                        buttons: [Dialog.okButton(), Dialog.cancelButton()]
                                    });
                                    console.error(response)
                                    return;
                                }
                                response.body?.getReader().read().then(({done, value}) => {
                                    if (done) {
                                        console.error("stream is closed");
                                        return;
                                    }
                                    let decoder = new TextDecoder('utf-8');
                                    let result = JSON.parse(decoder.decode(value))
                                    if (result.code == 0) {
                                        showDialog({
                                            title: "Get Attestation Report Successful",
                                            body: 'OIDC Token: ' + result.token,
                                            buttons: [Dialog.okButton(), Dialog.cancelButton()]
                                        });
                                    } else {
                                        showDialog({
                                            title: "Get Attestation Report Failed",
                                            body: 'Error: ' + result.msg,
                                            buttons: [Dialog.okButton(), Dialog.cancelButton()]
                                        });
                                    }
                                })
                            });
                        }
                       
                    }}>Access Report</Button>
                )
            } else {
                return (<div></div>)
            }
        }
    }
];

const JobTableComponent = (): JSX.Element => {
    const [data, setData] = useState([]);
    const [loading, setLoading] = useState(false);
    const [pagination, setPagination] = useState({
        sizeCanChange: true,
        showTotal: true,
        total: 0,
        pageSize: 10,
        current: 1,
        pageSizeChangeResetCurrent: true,
    });

    function onChangeTable(pagination: any) {
        const { current, pageSize } = pagination;
        setLoading(true);
        setTimeout(() => {
          fetchJobs(current, pageSize)
          setPagination((pagination) => ({ ...pagination, current, pageSize }));
          setLoading(false);
        }, 1000);
    }

    useEffect(() => {
        setLoading(true);
        fetchJobs(1, 10);
        setLoading(false);
    }, []);
    const fetchJobs = (current: number, pageSize: number = 10) => {
        const settings = ServerConnection.makeSettings();

        let requestUrlWithParams = settings.baseUrl + "manatee/jobs" + "?page=" + current + "&page_size=" + pageSize;
        ServerConnection.makeRequest(requestUrlWithParams, {
            method: "GET"
        }, settings).then(response => {
            if (response.status !== 200) {
                console.error(response)
                return;
            }
            response.body?.getReader().read().then(({done, value}) => {
                if (done) {
                    console.error("stream is closed");
                    return;
                }
                let decoder = new TextDecoder('utf-8');
                let result = JSON.parse(decoder.decode(value))
                if (result.code == 0) {
                    setData(result.jobs.map( (job: any) => {
                        const update_date = new Date(job.updated_at);
                        const create_date = new Date(job.created_at);
                        job.updated_at = update_date.toLocaleString();
                        job.created_at = create_date.toLocaleString();
                        return job
                    }))
                    setPagination((pagination) => ({...pagination, total: result.total}))
                } else {
                    console.error('error:', result.msg)
                }
                
            });
        });
    }

    return (
        <ConfigProvider locale={enUS}>
            <Table columns={columns} scroll={{x:1200, y:500}} data={data} onChange={onChangeTable} loading={loading} pagination={pagination} pagePosition={'bl'}></Table>
        </ConfigProvider>
    );
  };

class JobTableWidget extends ReactWidget {
    constructor() {
        super()
        this.addClass('jp-react-widget');
    }
    
    protected render(): JSX.Element {
        return (
        <div>
            <JobTableComponent />
        </div>)
    }
}

export class DataCleanRoomJobs extends PanelWithToolbar {
    constructor(options: DataCleanRoomJobs.IOptions) {
        super();
        const trans = (options.translator ?? nullTranslator).load('jupyterlab');
        this.title.label = trans.__('Jobs');
        let body = new JobTableWidget();
        this.addWidget(body);
    }; 
}

export namespace DataCleanRoomJobs {
    export interface IOptions {
        translator?: ITranslator;
    }
}