package service

import (
	"time"

	log "github.com/sirupsen/logrus"
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
