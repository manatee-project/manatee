namespace go job

enum JobStatus {
    Created = 0
    ImageBuilding = 1
    ImageBuildingFailed = 2
    VMWaiting = 3
    VMRunning = 4
    VMFinished = 5
    VMKilled = 6
    VMFailed = 7
    VMOther = 8
    VMLaunchFailed = 9
}

struct Job {
    1: i64 id
    2: string uuid
    3: string creator
    4: JobStatus job_status
    5: string jupyter_file_name
    6: string created_at
    7: string updated_at
}

struct Env {
    1: string key
    2: string value
}

struct SubmitJobRequest{
    1: string jupyter_file_name (api.body="filename", api.vd="len($) > 0 && len($) < 128 && regexp('^.*\\.ipynb$') && !regexp('.*\\.\\..*')")
    2: string creator (api.body="creator", api.vd="len($) > 0 && len($) < 32 && !regexp('.*\\.\\..*')") 
    3: list<Env> envs (api.body="envs", api.json="envs")
    4: i64 cpu_count = 2 (api.body="cpu_count", api.vd="len($) > 0", vt.gt = "1", vt.lt = "256")
    5: i64 disk_size = 50 (api.body="disk_size", api.vd="len($) > 0", vt.gt = "20", vt.lt = "1024")
    255: required string access_token     (api.header="Authorization")
}

struct SubmitJobResponse{
    1: i32 code
    2: string msg
    3: string uuid
}

struct QueryJobRequest {
    1: i64 page (api.body="page", api.query="page",api.vd="$>0")
    2: i64 page_size (api.body="page_size", api.query="page_size", api.vd="$ > 0 || $ <= 100")
    3: string creator (api.body="creator", api.vd="len($) > 0 && len($) < 32 && !regexp('.*\\.\\..*')")
    255: required string access_token     (api.header="Authorization")
}

struct QueryJobResponse {
    1: i32 code
    2: string msg
    3: list<Job> jobs
    4: i64 total
}

struct DeleteJobRequest {
    1: string uuid (api.body="uuid", api.query="uuid")
    2: string creator (api.body="creator", api.vd="len($) > 0 && len($) < 32 && !regexp('.*\\.\\..*')")
    255: required string access_token     (api.header="Authorization")
}

struct DeleteJobResponse {
    1: i32 code
    2: string msg
}

struct DownloadJobOutputRequest {
    1: i64 id (api.body="id", api.query="id", api.vd="$>0")
    2: string creator (api.body="creator", api.vd="len($) > 0 && len($) < 32 && !regexp('.*\\.\\..*')")
    255: required string access_token     (api.header="Authorization")
}

struct DownloadJobOutputResponse {
    1: i32 code
    2: string msg
    3: string signed_url
    4: string filename
}

struct QueryJobAttestationRequest {
    1: i64 id (api.body="id", api.query="id", api.vd="$>0")
    2: string creator (api.body="creator", api.vd="len($) > 0 && len($) < 32 && !regexp('.*\\.\\..*')")
}

struct QueryJobAttestationResponse {
    1: i32 code
    2: string msg
    3: string signed_url
}

service JobHandler {
    SubmitJobResponse SubmitJob(1:SubmitJobRequest req)(api.post="/v1/job/submit/")
    QueryJobResponse QueryJob(1:QueryJobRequest req)(api.post="/v1/job/query/")
    DeleteJobResponse DeleteJob(1:DeleteJobRequest req)(api.post="/v1/job/delete/")
    DownloadJobOutputResponse DownloadJobOutput(1:DownloadJobOutputRequest req) (api.post="/v1/job/output/download/")
    QueryJobAttestationResponse QueryJobAttestationReport(1:QueryJobAttestationRequest req)  (api.post="/v1/job/attestation/")
}