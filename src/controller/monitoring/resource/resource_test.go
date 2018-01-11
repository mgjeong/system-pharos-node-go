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
	GET_CPU_USAGE_CMD           = "cat /proc/stat | grep cpu"
	GET_MEM_USAGE_CMD           = "cat /proc/meminfo"
	GET_DISK_USAGE_CMD          = "df -m"

	MODEL_NAME = "model name	: Intel(R) Core(TM) i7-4790 CPU @ 3.60GHz\n"
	UNAME      = "Linux 4.10.0-42-generic x86_64\n"
	CPU_USAGE  = "cpu 101622 702 40379 12153720 11897 0 1222 0 0 0\ncpu0 13661 262 4994 1520352 413 0 250 0 0 0\ncpu1 13453 43 4188 1521872 346 0 136 0 0 0"
	MEM_USAGE  = "MemTotal: 8127136 kB\nMemFree: 1189944 kB\nMemAvailable: 3407004 kB"
	DISK_USAGE = "Filesystem 1M-blocks Used Available Use% Mounted on\nudev 3947 0 3947 0% /dev\ntmpfs 794 50 744 7% /run"
)

var (
	processor   = "Intel(R) Core(TM) i7-4790 CPU @ 3.60GHz"
	os          = "Linux 4.10.0-42-generic x86_64"
	procStatCPU = "cpu 101622 702 40379 12153720 11897 0 1222 0 0 0\ncpu0 13661 262 4994 1520352 413 0 250 0 0 0\ncpu1 13453 43 4188 1521872 346 0 136 0 0 0"
	procMeminfo = "MemTotal: 8127136 kB\nMemFree: 1189944 kB\nMemAvailable: 3407004 kB"
	df          = "Filesystem 1M-blocks Used Available Use% Mounted on\nudev 3947 0 3947 0% /dev\ntmpfs 794 50 744 7% /run"
)

func TestGetResourceInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_PROCESSOR_MODELNAME_CMD).Return(MODEL_NAME, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_OS_CMD).Return(UNAME, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPU_USAGE_CMD).Return(CPU_USAGE, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEM_USAGE_CMD).Return(MEM_USAGE, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISK_USAGE_CMD).Return(DISK_USAGE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	res, err := Executor.GetResourceInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := map[string]interface{}{
		"processor": processor,
		"os":        os,
		"cpu":       procStatCPU,
		"disk":      df,
		"mem":       procMeminfo,
	}
	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Errorf("Exepcted ret %s Actual ret %s", compareReturnVal, res)
	}
}

func TestGetPerformanceInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPU_USAGE_CMD).Return(CPU_USAGE, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEM_USAGE_CMD).Return(MEM_USAGE, nil),
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISK_USAGE_CMD).Return(DISK_USAGE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	res, err := Executor.GetPerformanceInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := map[string]interface{}{
		"cpu":  procStatCPU,
		"disk": df,
		"mem":  procMeminfo,
	}
	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Errorf("Expected ret %s Actual ret %s", compareReturnVal, res)
	}
}

func TestGetProcessorModel_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockCommand(ctrl)

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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPU_USAGE_CMD).Return(CPU_USAGE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj
	res, err := getCPUUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := procStatCPU
	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestGetCPUUsageWithShellError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPU_USAGE_CMD).Return("", errors.NotFound{"/proc/cpuinfo: No such file or directory"}),
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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_CPU_USAGE_CMD).Return("", nil),
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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEM_USAGE_CMD).Return(MEM_USAGE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj

	res, err := getMemUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	compareReturnVal := procMeminfo

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestGetMemUsageWithShellError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEM_USAGE_CMD).Return("", errors.NotFound{"/proc/meminfo: No such file or directory"}),
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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_MEM_USAGE_CMD).Return("", nil),
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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISK_USAGE_CMD).Return(DISK_USAGE, nil),
	)

	// pass mockObj to a real object.
	shellExecutor = shellMockObj

	res, err := getDiskUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	compareReturnVal := df

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestGetDiskUsageWithShellError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISK_USAGE_CMD).Return("", errors.NotFound{}),
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

	shellMockObj := shellmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		shellMockObj.EXPECT().ExecuteCommand(BASH, BASH_C_OPTION, GET_DISK_USAGE_CMD).Return("", nil),
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