package worker

import (
	"errors"

	"github.com/noncepad/worker-pool/pool"
)

func (sw *simpleWorker) Run(job pool.Job[Request]) (Result, error) {

	return Result{}, errors.New("not implemented yet")
}
