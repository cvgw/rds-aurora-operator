package service

import (
	"time"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	Unprovisioned     = "unprovisioned"
	Provisioning      = "provisioning"
	Provisioned       = "ready"
	DBInstanceReady   = "available"
	DBClusterReady    = "available"
	RequiredReadyTime = 120 * 1000000000
	WaitSleepTime     = 10
)

type StateHandler interface {
	State() string
	NoState() error
	Unprovisioned() error
	Provisioning() error
	Provisioned() error
	Logger() *log.Entry
}

func HandleState(handler StateHandler) (reconcile.Result, error) {
	state := handler.State()
	result := reconcile.Result{}

	handler.Logger().Debugf("current state is %s", state)
	switch state {
	case "":
		err := handler.NoState()
		if err != nil {
			return reconcile.Result{RequeueAfter: 1 * time.Second}, err
		}
	case Unprovisioned:
		err := handler.Unprovisioned()
		if err != nil {
			return reconcile.Result{RequeueAfter: 1 * time.Second}, err
		}
	case Provisioning:
		err := handler.Provisioning()
		if err != nil {
			return reconcile.Result{RequeueAfter: 1 * time.Second}, err
		}

		result.RequeueAfter = 10 * time.Second
	case Provisioned:
		err := handler.Provisioned()
		if err != nil {
			return reconcile.Result{RequeueAfter: 1 * time.Second}, err
		}

		if handler.State() == Provisioned {
			result.RequeueAfter = 10 * time.Second
		}
	}

	return result, nil
}

func ChangeState(logger *log.Entry, state string) string {
	logger.Debugf("setting state to %s", state)
	return state
}

func CalculateReadySince(logger *log.Entry, readySince int64) int64 {
	if readySince == 0 {
		ready := time.Now().UnixNano()
		logger.Debugf("setting ready since %d", ready)
		return ready
	}
	return readySince
}

func StateFromReadySince(logger *log.Entry, readySince int64) string {
	readyTime := time.Now().UnixNano() - readySince
	logger.Debugf("readyTime %d", readyTime)
	if readyTime >= RequiredReadyTime {
		return ChangeState(logger, Provisioned)
	}

	logger.Debug("waiting for db resource to be ready for minimum time")
	return Provisioning
}
