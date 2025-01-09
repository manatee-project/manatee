import google.cloud.storage as gcs
from google.cloud import resourcemanager_v3
from google.auth import load_credentials_from_dict
from enum import Enum
import pandas as pd
from urllib.parse import urlparse
import os
import logging
import io

logger = logging.getLogger(__name__)

class Gcp():
    def __init__(self):
        self.project_id = ""
        self.pool_name = ""
        self.project_number = ""
        self.service_account = ""
        
    def init(self, project_id, pool_name, service_account):
        self.project_id = project_id
        self.pool_name = pool_name
        self.service_account = service_account
        self.project_number = self.get_project_number(project_id)
        
    def get_project_number(self, project_id):
        client = resourcemanager_v3.ProjectsClient()
        project = client.get_project(name=f"projects/{project_id}")
        return project.name.split('/')[1]

gcp = Gcp()

class Stage(Enum):
    UNKNOWN = 0
    STAGE1 = 1
    STAGE2 = 2

class DataRepo():
    def __init__(self, stage_1_bucket, stage_2_bucket):
        self.stage1 = RemoteStorage.init(Stage.STAGE1, stage_1_bucket)
        self.stage2 = RemoteStorage.init(Stage.STAGE2, stage_2_bucket)

    def get_data(self, filename):
        if self.get_stage() == 1:
            return self.stage1.get_data(filename)
        elif self.get_stage() == 2:
            return self.stage2.get_data(filename)
        else:
            logger.warning("Unknown stage")
            return filename
        
    def get_stage(self):
        stage = int(os.getenv('EXECUTION_STAGE', '').strip('\'"'))
        return stage

class RemoteStorage():
    def __init__(self):
        pass

    def get_data(self, filename):
        pass
    
    @staticmethod
    def init(stage, url):
        try:
            o = urlparse(url, allow_fragments=False)
        except Exception as e:
            raise ValueError("Invalid URL: " + url)
        
        if o.scheme == "gs":
            return RemoteStorageGCS(stage, o.netloc, o.path)
        elif o.scheme == "s3":
            raise NotImplementedError("S3 storage not implemented")
        elif o.scheme == "https":
            raise NotImplementedError("HTTPS storage not implemented")
        else:
            raise ValueError("Invalid scheme: " + o.scheme)  
    

class RemoteStorageGCS(RemoteStorage):
    def __init__(self, stage, bucket_name, path):
        super().__init__()
        self.bucket = bucket_name
        self.path = path

        if stage == Stage.STAGE1:
            self.client = gcs.Client()
        elif stage == Stage.STAGE2:
            credentials_dict = {
              "type": "external_account",
              "audience": "//iam.googleapis.com/projects/%s/locations/global/workloadIdentityPools/%s/providers/attestation-verifier"%(gcp.project_number, gcp.pool_name),
              "subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
              "token_url": "https://sts.googleapis.com/v1/token",
              "credential_source": {
                "file": "/run/container_launcher/attestation_verifier_claims_token"
              },
              "service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s@%s.iam.gserviceaccount.com:generateAccessToken"%(gcp.service_account, gcp.project_id),
            }
            credentials, _ = load_credentials_from_dict(credentials_dict)
            self.client = gcs.Client(credentials=credentials)

    def get_data(self, filename):
        # join the path and filename
        full_path = os.path.join(self.path, filename)
        blob = self.client.get_bucket(self.bucket).blob(full_path)
        data = blob.download_as_text()
        return data
