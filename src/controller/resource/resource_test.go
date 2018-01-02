package resource

import (
	"commons/errors"
	shellmocks "controller/shellcommand/mocks"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

const (
	BASH                        = "bash"
	BASH_C_OPTION               = "-c"
	GET_PROCESSOR_MODELNAME_CMD = "grep -m1 ^'model name' /proc/cpuinfo"
	GET_OS_CMD                  = "uname -mrs"
	GET_CPUUSAGE_CMD            = "cat /proc/stat | grep cpu"
	GET_MEMTOTAL_CMD            = "cat /proc/meminfo | grep MemTotal: | awk '{print $2}'"
	GET_MEMFREE_CMD             = "cat /proc/meminfo | grep MemFree: | awk '{print $2}'"
	GET_DISKTOTAL_CMD           = "df -m | awk '{print $2}'"
	GET_DISKFREE_CMD            = "df -m | awk '{print $4}'"

	MODEL_NAME = "model name	: Intel(R) Core(TM) i7-4790 CPU @ 3.60GHz\n"
	UNAME      = "Linux 4.10.0-42-generic x86_64\n"
	CPU_USAGE  = "cpu 101622 702 40379 12153720 11897 0 1222 0 0 0\ncpu0 13661 262 4994 1520352 413 0 250 0 0 0\ncpu1 13453 43 4188 1521872 346 0 136 0 0 0"
	MEM_TOTAL  = "8127128"
	MEM_FREE   = "1658736"
	DISK_TOTAL = "1M-blocks\n3947\n794\n472576\n3969\n5\n3969\n794\n16\n" //sum : 486070
	DISK_FREE  = "Available\n3947\n776\n411281\n3958\n5\n3969\n793\n3\n"  //sum : 424732
)

var (
	processor     = "Intel(R) Core(TM) i7-4790 CPU @ 3.60GHz"
	os            = "Linux 4.10.0-42-generic x86_64"
	memUsagePerc  = "79.59%%"
	diskUsagePerc = "12.62%%"
	cpuUsagePerc  = "1.17%%"
)

func TestGetResourceInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_PROCESSOR_MODELNAME_CMD).Return(MODEL_NAME, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_OS_CMD).Return(UNAME, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPUUSAGE_CMD).Return(CPU_USAGE, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEMTOTAL_CMD).Return(MEM_TOTAL, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEMFREE_CMD).Return(MEM_FREE, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISKTOTAL_CMD).Return(DISK_TOTAL, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISKFREE_CMD).Return(DISK_FREE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	res, err := Resource.GetResourceInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := map[string]interface{}{
		"processor": processor,
		"os":        os,
		"cpu":       cpuUsagePerc,
		"disk":      diskUsagePerc,
		"mem":       memUsagePerc,
	}
	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Errorf("Exepcted ret %s Actual ret %s", compareReturnVal, res)
	}
}

func TestGetPerformanceInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPUUSAGE_CMD).Return(CPU_USAGE, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEMTOTAL_CMD).Return(MEM_TOTAL, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEMFREE_CMD).Return(MEM_FREE, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISKTOTAL_CMD).Return(DISK_TOTAL, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISKFREE_CMD).Return(DISK_FREE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	res, err := Resource.GetPerformanceInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := map[string]interface{}{
		"cpu":  cpuUsagePerc,
		"disk": diskUsagePerc,
		"mem":  memUsagePerc,
	}
	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Errorf("Expected ret %s Actual ret %s", compareReturnVal, res)
	}
}

func TestGetProcessorModel_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_PROCESSOR_MODELNAME_CMD).Return(MODEL_NAME, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	res, err := getProcessorModel()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := processor
	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Errorf("Expected ret %s Actual ret %s", compareReturnVal, res)
	}
}

func TestGetProcessorModelWithShellError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_PROCESSOR_MODELNAME_CMD).Return("", errors.NotFound{"/proc/cpuinfo: No such file or directory"}),
	)

	shellExecutor = shellMockObj
	_, err := getProcessorModel()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.NotFound{}) {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", reflect.TypeOf(err))
	}
}

func TestGetProcessorModelWithEmptyModelName_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_PROCESSOR_MODELNAME_CMD).Return("", nil),
	)

	shellExecutor = shellMockObj
	_, err := getProcessorModel()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.Unknown{}) {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", reflect.TypeOf(err))
	}
}

func TestGetOS_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_OS_CMD).Return(UNAME, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	res, err := getOS()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := os
	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestGetOSWithShellError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_OS_CMD).Return("", errors.Unknown{}),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	_, err := getOS()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.Unknown{}) {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", reflect.TypeOf(err))
	}
}

func TestGetCPUUsage_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPUUSAGE_CMD).Return(CPU_USAGE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	res, err := getCPUUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := cpuUsagePerc
	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestGetCPUUsageWithShellError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPUUSAGE_CMD).Return("", errors.NotFound{"/proc/cpuinfo: No such file or directory"}),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	_, err := getCPUUsage()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.NotFound{}) {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", reflect.TypeOf(err))
	}
}

func TestGetCPUUsageWithEmptyCPUInfo_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPUUSAGE_CMD).Return("", nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	_, err := getCPUUsage()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.Unknown{}) {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", reflect.TypeOf(err))
	}
}

func TestGetMemUsage_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEMTOTAL_CMD).Return(MEM_TOTAL, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEMFREE_CMD).Return(MEM_FREE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj

	res, err := getMemUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	compareReturnVal := memUsagePerc

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestGetMemUsageWithShellError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEMTOTAL_CMD).Return("", errors.NotFound{"/proc/meminfo: No such file or directory"}),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	_, err := getMemUsage()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.NotFound{}) {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", reflect.TypeOf(err))
	}
}

func TestGetMemUsageWithEmptyMemInfo_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEMTOTAL_CMD).Return("", nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	_, err := getMemUsage()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.Unknown{}) {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", reflect.TypeOf(err))
	}
}

func TestGetDiskUsage_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISKTOTAL_CMD).Return(DISK_TOTAL, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISKFREE_CMD).Return(DISK_FREE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj

	res, err := getDiskUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	compareReturnVal := diskUsagePerc

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestGetDiskUsageWithShellError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISKTOTAL_CMD).Return("", errors.NotFound{}),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	_, err := getDiskUsage()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.NotFound{}) {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", reflect.TypeOf(err))
	}
}

func TestGetDiskUsageWithEmptyMemInfo_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockShellInterface(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISKTOTAL_CMD).Return("", nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	_, err := getDiskUsage()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	} else if reflect.TypeOf(err) != reflect.TypeOf(errors.Unknown{}) {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", reflect.TypeOf(err))
	}
}
