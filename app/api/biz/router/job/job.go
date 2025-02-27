// Code generated by hertz generator. DO NOT EDIT.

package job

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	job "github.com/manatee-project/manatee/app/api/biz/handler/job"
)

/*
 This file will register all the routes of the services in the master idl.
 And it will update automatically when you use the "update" command for the idl.
 So don't modify the contents of the file, or your code will be deleted when it is updated.
*/

// Register register routes based on the IDL 'api.${HTTP Method}' annotation.
func Register(r *server.Hertz) {

	root := r.Group("/", rootMw()...)
	{
		_v1 := root.Group("/v1", _v1Mw()...)
		{
			_job := _v1.Group("/job", _jobMw()...)
			{
				_attestation := _job.Group("/attestation", _attestationMw()...)
				_attestation.POST("/", append(_queryjobattestationreportMw(), job.QueryJobAttestationReport)...)
			}
			{
				_delete := _job.Group("/delete", _deleteMw()...)
				_delete.POST("/", append(_deletejobMw(), job.DeleteJob)...)
			}
			{
				_output := _job.Group("/output", _outputMw()...)
				{
					_download := _output.Group("/download", _downloadMw()...)
					_download.POST("/", append(_downloadjoboutputMw(), job.DownloadJobOutput)...)
				}
			}
			{
				_query := _job.Group("/query", _queryMw()...)
				_query.POST("/", append(_queryjobMw(), job.QueryJob)...)
			}
			{
				_submit := _job.Group("/submit", _submitMw()...)
				_submit.POST("/", append(_submitjobMw(), job.SubmitJob)...)
			}
		}
	}
}
