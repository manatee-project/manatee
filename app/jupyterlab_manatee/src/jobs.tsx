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
import { ConfigProvider, Table, TableColumnProps, Descriptions, Button, Tag} from '@arco-design/web-react';
import enUS from '@arco-design/web-react/es/locale/en-US';
import "@arco-design/web-react/dist/css/arco.css";
import {IconCheckCircle, IconSync, IconExclamationCircle, IconCloseCircle, IconCheckSquare, IconDownload, IconRefresh, IconLoading} from '@arco-design/web-react/icon';

const statusMap: Map<number, {icon: React.JSX.Element, color: string, text: string}> = new Map([
    [1, {icon: <IconSync spin style={{ fontSize: 20 }}/>,                          color: 'green', text: 'Image Building'}],
    [2, {icon: <IconCloseCircle style={{ color: 'red', fontSize: 20 }}/>,          color: 'red',   text: 'Image Building Failed'}],
    [3, {icon: <IconSync spin style={{ fontSize: 20 }} />,                         color: 'green', text: 'Starting Executor'}],
    [4, {icon: <IconSync spin style={{ fontSize: 20 }} />,                         color: 'green', text: 'Running Executor'}],
    [5, {icon: <IconCheckCircle style={{ color:'green', fontSize: 20 }} />,        color: 'green', text: 'Finished'}],
    [6, {icon: <IconCloseCircle style={{color: 'red', fontSize: 20 }}/>,           color: 'red',   text: 'Executor Killed'}],
    [7, {icon: <IconCloseCircle style={{color: 'red', fontSize: 20 }} />,          color: 'red',   text: 'Executor Failed'}],
    [8, {icon: <IconExclamationCircle style={{color: '#ffcd00', fontSize: 20}} />, color: 'gray',  text: 'Unknown'}],
    [9, {icon: <IconCloseCircle style={{color: 'red', fontSize: 20 }} />,          color: 'red',   text: 'Launch Failed'}] 
]);

interface Job {
    id: number;
    jupyter_file_name: string;
    job_status: number;
    created_at: string;
    updated_at: string;
}


const columns: TableColumnProps[] = [
    {
        title: 'ID',
        dataIndex: 'id',
        width: 50,
    },
    {
        title: 'Filename',
        dataIndex: 'jupyter_file_name',
        ellipsis: true,
    },
    {
        title: '', // this will be later overwritten by a refresh icon.
        dataIndex: 'job_status',
        width:  40,
        render: (col, item: any, index) => {
            let status = statusMap.get(col)
            
            if (status == undefined) {
                return (<IconLoading style={{fontSize: 20}} />)
            }

            return (
                status.icon
            )
        }
    },
];

const handleDownloadOutput = async (record: Job) => {
    const id = record.id;
    const request = JSON.stringify({ id });
    const settings = ServerConnection.makeSettings();
    const result = await showDialog({
        title: "Download the output of the Job?",
        body: `Job ID: ${id}`,
        buttons: [Dialog.okButton(), Dialog.cancelButton()]
    });

    if (result.button.accept) {
        try {
            const response = await ServerConnection.makeRequest(settings.baseUrl + "manatee/output", {
                body: request,
                method: "POST"
            }, settings);

            if (response.status !== 200) {
                await showDialog({
                    title: "Download Failed",
                    body: "Server error during download.",
                    buttons: [Dialog.okButton()]
                });
                console.error(response);
                return;
            }

            const { done, value } = await response.body!.getReader().read();
            if (done || !value) {
                console.error("stream is closed");
                return;
            }

            const decoder = new TextDecoder('utf-8');
            const result = JSON.parse(decoder.decode(value));
            if (result.code === 0) {
                await showDialog({
                    title: "Download Successful",
                    body: `Filename: ${result.filename}`,
                    buttons: [Dialog.okButton()]
                });
            } else {
                await showDialog({
                    title: "Download Failed",
                    body: `Error: ${result.msg}`,
                    buttons: [Dialog.okButton()]
                });
            }
        } catch (err) {
            console.error("Error during download:", err);
        }
    }
};

const handleGetAttestation = async (record: Job) => {
    const id = record.id
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
        }, settings).then(async response => {
            if (response.status !== 200) {
                showDialog({
                    title: "Get Attestation Report Failed",
                    buttons: [Dialog.okButton(), Dialog.cancelButton()]
                });
                console.error(response)
                return;
            }
            try {
                const result = await response.json();
                if (result.code === 0) {
                    await showDialog({
                        title: "Get Attestation Report Successful",
                        body: 'OIDC Token: ' + result.token,
                        buttons: [Dialog.okButton(), Dialog.cancelButton()]
                    });
                } else {
                    await showDialog({
                        title: "Get Attestation Report Failed",
                        body: 'Error: ' + result.msg,
                        buttons: [Dialog.okButton(), Dialog.cancelButton()]
                    });
                }
            } catch (e) {
                console.error("Failed to parse response JSON", e);
                await showDialog({
                    title: "Invalid Response",
                    body: 'Could not parse JSON: ' + String(e),
                    buttons: [Dialog.okButton()]
                });
            }
        
        });
    }
};


const JobTableComponent = (): JSX.Element => {
    const [data, setData] = useState<Job[]>([]);
    const [loading, setLoading] = useState(false);
    const [pagination, setPagination] = useState({
        simple: true,
        sizeCanChange: false,
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


    useEffect(() => {
        const interval = setInterval(() => {
        fetchJobs(pagination.current, pagination.pageSize);
        }, 10000); // Refresh every 10 seconds

        return () => clearInterval(interval); // Cleanup on unmount
    }, [pagination.current, pagination.pageSize]);      

    // Create a mutable copy of columns
    const tableColumns = [...columns];
    // Set the refresh icon with access to pagination
    tableColumns[2].title = (
        <IconRefresh 
            style={{ fontSize: 20 }} 
            onClick={() => fetchJobs(pagination.current, pagination.pageSize)}
        />
    );

    return (
        <ConfigProvider locale={enUS}>            
            <Table 
                rowKey="id"
                columns={tableColumns}
                data={data} 
                onChange={onChangeTable} 
                loading={loading} 
                pagination={pagination}
                pagePosition={'bottomCenter'}
                expandedRowRender={(record) => {
                    const isFinished = record.job_status === 5;
                    let status = statusMap.get(record.job_status);
                    let text = "Unknown"
                    let color = "gray"
                    if (status != undefined) {
                        text = status.text
                        color = status.color
                    }
                    const descriptionData = [
                        { label: 'Job ID', value: record.id },
                        { label: 'Jupyter File', value: record.jupyter_file_name },
                        { label: 'Job Status', value: <Tag color={color}>{text}</Tag>  },
                        { label: 'Created At', value: record.created_at },
                        { label: 'Updated At', value: record.updated_at },
                        { label: 'Download', value: 
                            <div>
                                <Button
                                    icon={<IconDownload />}
                                    size='mini'
                                    type="primary"
                                    onClick={() => handleDownloadOutput(record)}
                                    disabled={!isFinished}
                                    style={{width: '80px', margin: '3px' }}
                                >
                                    Output
                                </Button>
                                <Button
                                    icon={<IconCheckSquare />}
                                    size='mini'
                                    type="primary"
                                    onClick={() => handleGetAttestation(record)}
                                    disabled={!isFinished}
                                    style={{width: '80px', margin: '3px'}}
                                >
                                    Report
                                </Button>
                            </div> 
                        },
                    ];

                    return <Descriptions
                        size='small'
                        colon=':' layout='inline-horizontal' data={descriptionData}
                        border
                        column={1}
                    />
                }}
                expandProps={{
                    expandRowByClick: true,
                }}
            />
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