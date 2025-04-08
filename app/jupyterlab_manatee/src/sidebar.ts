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

import { SidePanel, trustedIcon } from '@jupyterlab/ui-components';
import { ITranslator, nullTranslator } from '@jupyterlab/translation';
import { IDocumentManager } from '@jupyterlab/docmanager';
import { DataCleanRoomSources } from './sources';
// import { DataCleanRoomInputs } from './inputs';
import { DataCleanRoomJobs } from './jobs';

export class DataCleanRoomSidebar extends SidePanel {
    constructor(options: DataCleanRoomSidebar.IOptions) {
        const { manager } = options;
        const translator = options.translator || nullTranslator;
        super({ translator });

        const jobsPanel = new DataCleanRoomJobs({ translator });
        const sourcesPanel = new DataCleanRoomSources({ manager, translator });
        this.addClass("jp-manatee-sidebar-view")

        this.title.icon = trustedIcon;
        this.id = "jp-DCRSource-sidebar"
        this.addWidget(jobsPanel);
        this.addWidget(sourcesPanel);
    }
}

export namespace DataCleanRoomSidebar {
    export interface IOptions {
        manager: IDocumentManager;
        translator?: ITranslator;
    }
}
