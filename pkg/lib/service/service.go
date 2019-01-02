package service

import (
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/pkg/errors"
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

func SliceEqual(sliceA []string, sliceB []string) bool {
	if len(sliceA) != len(sliceB) {
		return false
	}

	for _, a := range sliceA {
		match := false
		for _, b := range sliceB {
			if a == b {
				match = true
				break
			}
		}
		if match == false {
			return false
		}
	}

	return true
}

func PopulateValidationErr(prevErr, newErr error) error {
	if prevErr == nil {
		return newErr
	}
	return errors.Wrap(prevErr, newErr.Error())
}

type Handler struct {
	sHandler StateHandler
}

func (h *Handler) SetStateHandler(sHandler StateHandler) *Handler {
	h.sHandler = sHandler
	return h
}

func (h *Handler) NoState() error {
	h.sHandler.MutateState(Unprovisioned)

	return h.sHandler.NoState()
}

func (h *Handler) Unprovisioned() error {
	h.sHandler.MutateReadySince(0)
	h.sHandler.MutateState(Provisioning)

	return h.sHandler.Unprovisioned()
}

func (h *Handler) Provisioning() error {
	HandleResourceStatus(h.sHandler, "")

	return h.sHandler.Provisioning()
}

func (h *Handler) Provisioned() error {
	return h.sHandler.Provisioned()
}

type BaseStateHandler struct {
	Service *rds.RDS
	Logr    *log.Entry
}

func (s *BaseStateHandler) Logger() *log.Entry {
	return s.Logr
}

func (s *BaseStateHandler) Svc() *rds.RDS {
	return s.Service
}

type StateHandler interface {
	State() string
	ReadySince() int64
	MutateState(string)
	MutateReadySince(int64)
	NoState() error
	ResourceReady(string) bool

	Unprovisioned() error
	Provisioning() error
	Provisioned() error
	Logger() *log.Entry
}

func HandleResourceStatus(handler StateHandler, resourceStatus string) {
	if handler.ResourceReady(resourceStatus) {
		handler.Logger().Debug("db resource is ready")

		handler.MutateReadySince(CalculateReadySince(handler.Logger(), handler.ReadySince()))
		handler.MutateState(StateFromReadySince(handler.Logger(), handler.ReadySince()))
	} else {
		handler.Logger().Debug("db resource is not ready")
		handler.MutateReadySince(0)
	}
}

func Handle(handler Handler) (reconcile.Result, error) {
	state := handler.sHandler.State()
	result := reconcile.Result{}

	handler.sHandler.Logger().Debugf("current state is %s", state)
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

		if handler.sHandler.State() == Provisioned {
			result.RequeueAfter = 10 * time.Second
		}
	}

	return result, nil
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
