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

import { 
  JupyterFrontEnd,
  JupyterFrontEndPlugin,
  ILayoutRestorer,
} from '@jupyterlab/application';
import { IDocumentManager } from '@jupyterlab/docmanager';
import { ITranslator } from '@jupyterlab/translation';
import { DataCleanRoomSidebar } from './sidebar';


async function activate(app: JupyterFrontEnd, docManager: IDocumentManager, translator: ITranslator, restorer: ILayoutRestorer | null) {
  console.log("JupyterLab extension jupyterlab_manatee is activated!");

  const sidebar = new DataCleanRoomSidebar({manager: docManager});

  app.shell.add(sidebar, 'right',  {rank: 0});

  if (restorer) {
    restorer.add(sidebar, "data-clean-room-side-bar");
  }
}

/**
 * Initialization data for the jupyterlab-manatee extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab_manatee:plugin',
  description: 'This is an open-source JupyterLab extension for ManaTEE framework',
  autoStart: true,
  requires: [IDocumentManager, ITranslator],
  optional: [ILayoutRestorer],
  activate: activate
};

export default plugin;
