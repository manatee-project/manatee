// Copyright 2024 TikTok Pte. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errno

import (
	"fmt"
)

const (
	SuccessCode    = 0
	ServiceErrCode = iota + 10000
	ReachJobLimitErrCode
)

const (
	SuccessMsg          = "Success"
	ServiceErrMsg       = "Service internal error"
	ReachJobLimitErrMsg = "The number of in progress jobs has reached the limit"
)

type ErrNo struct {
	ErrCode int32
	ErrMsg  string
}

func (e ErrNo) Error() string {
	return fmt.Sprintf("err_code=%d, err_msg=%s", e.ErrCode, e.ErrMsg)
}

func NewErrNo(code int32, msg string) ErrNo {
	return ErrNo{code, msg}
}

func (e ErrNo) WithMessage(msg string) ErrNo {
	e.ErrMsg = msg
	return e
}

var (
	Success          = NewErrNo(SuccessCode, SuccessMsg)
	ServiceErr       = NewErrNo(ServiceErrCode, ServiceErrMsg)
	ReachJobLimitErr = NewErrNo(ReachJobLimitErrCode, ReachJobLimitErrMsg)
)
