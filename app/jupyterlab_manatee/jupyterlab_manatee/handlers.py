# Copyright 2024 TikTok Pte. Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json
import os
import tarfile
from http import HTTPStatus

import aiohttp
import aiofiles
import base64
import tornado
from aiohttp import FormData
from jupyter_server.base.handlers import JupyterHandler
from jupyter_server.utils import url_path_join

def ignore_hidden_files(tarinfo):
    if os.path.basename(tarinfo.name).startswith('.') or os.path.basename(tarinfo.name).startswith('lost+found'):
        return None
    else:
        return tarinfo

def make_tarfile(output_filename, source_dir, arcname):
    with tarfile.open(output_filename, "w:gz") as tar:
        tar.add(source_dir, arcname=arcname, filter=ignore_hidden_files)

def get_data_clean_room_url():
        # Get the value of an environment variable
        value = os.getenv('DATA_CLEAN_ROOM_HOST', '')
        # Check if 'http' protocol is present or not. If not, add 'http://'
        if not value.startswith(('http://', 'https://')):
            value = 'http://' + value
        return value

# Developer should develop the authenticator within hub image to pass user token to the single user pod through environment variable.
def get_user_token():
    token = os.getenv('USER_TOKEN', '')
    return token

async def make_proxied_post_request(logger, endpoint, body, headers) -> str:
    """
    Make a proxied POST request to the endpoint, and return the response.
    """
    url = url_path_join(get_data_clean_room_url(), endpoint)
    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(url, json=body, headers=headers) as response:
                if response.status != 200:
                    logger.error(response)
                    raise tornado.web.HTTPError(500, reason="Error from Data Clean Room API")
                return await response.text()
    except Exception as e:
        logger.error(e)
        raise tornado.web.HTTPError(500, reason="Failed to make a request to Data Clean Room API")


class DataCleanRoomJobHandler(JupyterHandler):
    """
    A Job Handler for Data Clean Room API.
    """

    def _build_form_data(self, workspace_file, creator, jupyter_filename, envs) -> FormData:
        data = FormData()
        data.add_field('file',
                        value=open(workspace_file, 'rb'),
                        filename='workspace.tar.gz',
                        content_type='application/gzip')
        data.add_field("envs", json.dumps(envs), content_type="application/json")
        data.add_field('creator', creator)
        data.add_field('filename', jupyter_filename)
        return data

    async def post_file(self, endpoint, body, workspace_filename, envs, headers) -> str:
        """
        Post file to Data Clean Room API.
        """
        url = url_path_join(get_data_clean_room_url(), endpoint)
        try:
            async with aiohttp.ClientSession() as session:
                data = self._build_form_data(workspace_filename, body['creator'], body['filename'], envs)
                async with session.post(url, data=data, headers=headers, allow_redirects=False) as response:
                    if response.status == HTTPStatus.TEMPORARY_REDIRECT:
                        # when redirect, post manually again
                        data = self._build_form_data(workspace_filename, body['creator'], body['filename'], envs)
                        redirect_url = url_path_join(get_data_clean_room_url(), response.headers['Location']) 
                        async with session.post(redirect_url, data=data, headers=headers) as redirect_resp:
                            return await redirect_resp.text()
                    if response.status == HTTPStatus.OK:
                        return await response.text()
        except Exception as e:
            self.log.error(e)
            raise tornado.web.HTTPError(500, reason="Failed to make a request to Data Clean Room API")

    @tornado.web.authenticated
    async def post(self):
        """
        Submit job to Data Clean Room API.
        """
        creator = self.current_user.username

        request_body = json.loads(self.request.body.decode('utf-8'))
        request_body['creator'] = creator
        
        # Get extra envs from environment variables
        extra_env_prefix = "MANATEE_EXTRA_ENV_"
        envs = []
        for key, value in os.environ.items():
            if key.startswith(extra_env_prefix):
                envs.append({
                    "key": key.replace(extra_env_prefix,""),
                    "value": value
                })

        headers = {
            "Authorization" : get_user_token()
        }

        if "filename" not in request_body or "path" not in request_body:
            raise tornado.web.HTTPError(500, reason="Missing arguments")
        # Pack the work directory into a tar archive
        tar_filename = "/tmp/workspace.tar.gz"
        make_tarfile(tar_filename, os.getcwd(), creator + '-workspace')

        self.finish(await self.post_file("v1/job/submit", request_body, tar_filename, envs, headers))

    @tornado.web.authenticated
    async def get(self):
        """
        List jobs from Data Clean Room API
        """
        creator = self.current_user.username

        request_body = {
            "page": int(self.get_argument("page", 1)),
            "page_size": int(self.get_argument("page_size", 10)),
            "creator": creator
        }

        headers = {
            "Authorization" : get_user_token()
        }

        self.finish(await make_proxied_post_request(self.log, "v1/job/query", request_body, headers))

class DataCleanRoomOutputHandler(JupyterHandler):
    @tornado.web.authenticated
    async def post(self):
        request_body = json.loads(self.request.body.decode('utf-8'))
        creator = self.current_user.username
        request_body['creator'] = creator
        headers = {
            "Authorization" : get_user_token()
        }

        await self.download_file("v1/job/output/download", request_body, headers)

    async def download_file(self, endpoint, request_body, headers):
        url = url_path_join(get_data_clean_room_url(), endpoint)
        offset = 0
        chunk = 1024 * 1024 * 3 # 3 MB
        request_body['chunk'] = chunk

        download_resp_str = await make_proxied_post_request(self.log, endpoint, request_body, headers)
        download_resp = json.loads(download_resp_str)
        filename = download_resp['filename']
        if download_resp['code'] != 0:
            self.finish(download_resp)
            return 
        signed_url = download_resp['signed_url']
        filename = download_resp['filename']
        async with aiohttp.ClientSession() as session:
            async with session.get(signed_url) as response:
                if response.status == 200:
                    async with aiofiles.open(filename, 'wb') as f:
                        while True:
                            chunk = await response.content.read(4096)
                            if not chunk:
                                break
                            await f.write(chunk)
                    self.finish(json.dumps({
                        "code": 0,
                        "msg": "Success",
                        "filename": filename
                    }).encode('utf-8'))
                else:
                    raise tornado.web.HTTPError(500, reason="Failed to get download output file  through signed url")


class DataCleanRoomAttestationHandler(JupyterHandler):
    @tornado.web.authenticated
    async def get(self):
        """
        Get attestation report from Data Clean Room API
        """
        creator = self.current_user.username

        request_body = {
            "id": int(self.get_argument("id", 1)),
            "creator": creator
        }

        headers = {
            "Authorization" : get_user_token()
        }
        attestation_resp_str = await make_proxied_post_request(self, "v1/job/attestation/", request_body, headers)
        attestation_resp = json.loads(attestation_resp_str)
        async with aiohttp.ClientSession() as session:
            async with session.get(attestation_resp['signed_url']) as response:
                if response.status == 200:
                    content = await response.read()
                    content_str = content.decode('utf-8')
                    self.finish(json.dumps({
                        "code": 0,
                        "msg": "Success",
                        "token": content_str
                    }).encode('utf-8'))
                else:
                    raise tornado.web.HTTPError(500, reason="Failed to get attestation report through signed url")