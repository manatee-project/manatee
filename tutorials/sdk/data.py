import google.cloud.storage as gcs
from enum import Enum
import pandas as pd
from urllib.parse import urlparse
import os
import logging
import io

logger = logging.getLogger(__name__)

class Stage(Enum):
    UNKNOWN = 0
    STAGE1 = 1
    STAGE2 = 2

class RemoteStorage():
    def __init__(self):
        pass

    def get_stage(self):
        stage = os.getenv('EXECUTION_STAGE', '')
        if stage == '1':
            return 1
        elif stage == '2':
            return 2
        else:
            return 0

    def get_filename(self, filename: str):
        if self.get_stage() == 1:
            return filename + ".s1"
        elif self.get_stage() == 2:
            return filename + ".s2"
        else:
            logger.warning("Unknown stage")
            return filename

class RemoteStorageGCS(RemoteStorage):
    def __init__(self, bucket_name, path):
        super().__init__()
        self.bucket_name = bucket_name
        self.path = path
        self.client = gcs.Client()
        self.bucket = self.client.get_bucket(bucket_name)

    def get_data(self, filename):
        actual_filename = self.get_filename(filename)
        # join the path and filename
        full_path = os.path.join(self.path, actual_filename)
        blob = self.bucket.blob(full_path)
        data = blob.download_as_string()
        return pd.read_csv(io.BytesIO(data))

def init(url):
    try:
        o = urlparse(url, allow_fragments=False)
    except Exception as e:
        raise ValueError("Invalid URL: " + url)
    
    if o.scheme == "gs":
        return RemoteStorageGCS(o.netloc, o.path)
    # elif o.scheme == "s3":
    #    return RemoteStorageS3(o.netloc, o.path)
    # else:
    #     # local
    #     return RemoteStorageLocal(o.path)
    else:
        raise ValueError("Invalid scheme: " + o.scheme)  